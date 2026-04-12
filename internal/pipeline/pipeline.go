package pipeline

import (
	"context"
	"fmt"
	"sync"

	"github.com/ryansteffan/simply_syslog/pkg/applogger"
)

// RunnerState represents the current state of a pipeline node.
type RunnerState int

const (
	Starting RunnerState = iota // The node is in the process of starting
	Running                     // The node is currently running
	Stopping                    // The node is in the process of stopping
	Stopped                     // The node has been stopped
)

type Runner interface {
	Start() error
	Stop() error
	Restart() error
	Wait()
}

type Node interface {
	Runner
	GetName() string
	GetLogger() applogger.Logger
	GetState() RunnerState
	GetParentPipeline() *Pipeline
	setParentPipeline(p *Pipeline)
}

type Pipeline struct {
	Nodes  []Node           // Order slice of pipeline nodes
	Logger applogger.Logger // Logger for the pipeline to use

	baseContext        context.Context    // Base context from which pipeline contexts are derived
	pipelineContext    context.Context    // Context for managing the pipeline's lifecycle
	pipelineCancelFunc context.CancelFunc // Function to cancel the pipeline's context
	mutex              sync.Mutex         // Mutex to protect state changes at the pipeline level
	state              RunnerState        // Current state of the pipeline
}

func NewPipeline(parentContext context.Context, logger applogger.Logger) *Pipeline {
	return &Pipeline{
		Nodes:       make([]Node, 0),
		Logger:      logger,
		baseContext: parentContext,
		state:       Stopped,
	}
}

func (p *Pipeline) AddNode(node Node) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if p.state != Stopped {
		return fmt.Errorf("cannot add node: pipeline is not stopped")
	}
	if p.Logger != nil {
		p.Logger.Debug(fmt.Sprintf("adding node %s to pipeline", node.GetName()))
	}
	node.setParentPipeline(p)
	p.Nodes = append(p.Nodes, node)
	return nil
}

// Restart implements [Runner].
func (p *Pipeline) Restart() error {
	if err := p.Stop(); err != nil {
		return err
	}
	return p.Start()
}

// Start implements [Runner].
func (p *Pipeline) Start() error {
	p.mutex.Lock()
	if p.state == Running || p.state == Starting {
		p.mutex.Unlock()
		return fmt.Errorf("pipeline is already running or starting")
	}
	p.state = Starting
	p.pipelineContext, p.pipelineCancelFunc = context.WithCancel(p.baseContext)
	nodes := make([]Node, len(p.Nodes))
	copy(nodes, p.Nodes)
	p.mutex.Unlock()

	if p.Logger != nil {
		p.Logger.Debug(fmt.Sprintf("starting pipeline with %d node(s)", len(nodes)))
	}

	startedNodes := make([]Node, 0)
	for _, node := range nodes {
		if p.Logger != nil {
			p.Logger.Debug(fmt.Sprintf("starting node %s", node.GetName()))
		}
		if err := node.Start(); err != nil {
			// If starting a node fails, stop all previously started nodes
			for _, startedNode := range startedNodes {
				startedNode.Stop()
			}

			p.mutex.Lock()
			p.state = Stopped
			if p.pipelineCancelFunc != nil {
				p.pipelineCancelFunc()
			}
			p.mutex.Unlock()

			return fmt.Errorf("failed to start node %s: %w", node.GetName(), err)
		}
		startedNodes = append(startedNodes, node)
		if p.Logger != nil {
			p.Logger.Debug(fmt.Sprintf("started node %s", node.GetName()))
		}
	}

	p.mutex.Lock()
	p.state = Running
	p.mutex.Unlock()

	if p.Logger != nil {
		p.Logger.Debug("pipeline state changed to running")
	}

	return nil
}

// Stop implements [Runner].
func (p *Pipeline) Stop() error {
	p.mutex.Lock()
	if p.state == Stopped {
		p.mutex.Unlock()
		return nil
	}

	if p.state != Running && p.state != Stopping {
		state := p.state
		p.mutex.Unlock()
		return fmt.Errorf("pipeline is not running (state: %v)", state)
	}

	if p.state == Stopping {
		p.mutex.Unlock()
		p.Wait()
		return nil
	}

	p.state = Stopping
	nodes := make([]Node, len(p.Nodes))
	copy(nodes, p.Nodes)
	p.mutex.Unlock()

	if p.Logger != nil {
		p.Logger.Debug(fmt.Sprintf("stopping pipeline with %d node(s)", len(nodes)))
	}

	var stopErrs []error
	for nodeIndex := len(nodes) - 1; nodeIndex >= 0; nodeIndex-- {
		node := nodes[nodeIndex]
		if p.Logger != nil {
			p.Logger.Debug(fmt.Sprintf("stopping node %s", node.GetName()))
		}
		if err := node.Stop(); err != nil {
			stopErrs = append(stopErrs, fmt.Errorf("failed to stop node %s: %w", node.GetName(), err))
		}
	}

	p.mutex.Lock()
	p.state = Stopped

	// Cleanup the pipeline context after nodes have been safely stopped
	if p.pipelineCancelFunc != nil {
		p.pipelineCancelFunc()
	}

	p.mutex.Unlock()

	if p.Logger != nil {
		p.Logger.Debug("pipeline state changed to stopped")
	}

	if len(stopErrs) > 0 {
		return fmt.Errorf("errors occurred while stopping nodes: %v", stopErrs)
	}

	return nil
}

// Wait implements [Runner].
func (p *Pipeline) Wait() {
	p.mutex.Lock()
	nodes := make([]Node, len(p.Nodes))
	copy(nodes, p.Nodes)
	p.mutex.Unlock()

	for _, node := range nodes {
		node.Wait()
	}
}

var _ Runner = (*Pipeline)(nil)

type PipelineNode[T any, K any] struct {
	name    string           // A unique name for the node
	process Processor[T, K]  // The function that handles processing data
	inChan  <-chan T         // The channel that the processor reads from
	outChan chan<- K         // The channel that the processor writes to
	errChan chan<- error     // The channel that the processor writes errors to
	logger  applogger.Logger // A logger for logging messages

	parent         *Pipeline          // Reference to the parent pipeline, attached when added to pipeline.
	state          RunnerState        // Current state of the node
	nodeContext    context.Context    // Context for managing the node's lifecycle
	nodeCancelFunc context.CancelFunc // Function to cancel the node's context
	wg             *sync.WaitGroup    // WaitGroup to manage goroutines
	mutex          sync.Mutex         // Mutex to protect state changes
}

func NewPipelineNode[T any, K any](
	name string,
	logger applogger.Logger,
	inChan <-chan T,
	outChan chan<- K,
	errChan chan<- error,
	process Processor[T, K],
) Node {
	return &PipelineNode[T, K]{
		name:    name,
		process: process,
		inChan:  inChan,
		outChan: outChan,
		errChan: errChan,
		logger:  logger,
		state:   Stopped,
		wg:      new(sync.WaitGroup),
	}
}

// GetName implements [Node].
func (p *PipelineNode[T, K]) GetName() string {
	return p.name
}

func (p *PipelineNode[T, K]) GetLogger() applogger.Logger {
	return p.logger
}

// GetParentPipeline implements [Node].
func (p *PipelineNode[T, K]) GetParentPipeline() *Pipeline {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return p.parent
}

// GetState implements [Node].
func (p *PipelineNode[T, K]) GetState() RunnerState {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return p.state
}

// setParentPipeline implements [Node].
func (p *PipelineNode[T, K]) setParentPipeline(parent *Pipeline) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.parent = parent
}

// Restart implements [Runner].
func (p *PipelineNode[T, K]) Restart() error {
	errors := make([]error, 0)
	if err := p.Stop(); err != nil {
		errors = append(errors, fmt.Errorf("failed to stop node %s: %w", p.name, err))
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors occurred during restart: %v", errors)
	}
	return p.Start()
}

// Start implements [Runner].
func (p *PipelineNode[T, K]) Start() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if p.state == Running || p.state == Starting {
		return fmt.Errorf("node %s is already running", p.name)
	}

	p.state = Starting
	p.logger.Debug(fmt.Sprintf("node %s transitioning to starting", p.name))

	if p.parent == nil {
		p.logger.Error("node not attached to pipeline, context derived from background")
		p.nodeContext, p.nodeCancelFunc = context.WithCancel(context.Background())
	} else {
		p.nodeContext, p.nodeCancelFunc = context.WithCancel(p.parent.pipelineContext)
	}

	ctx := p.nodeContext
	cancel := p.nodeCancelFunc

	p.wg.Add(1)
	go func() {
		defer func() {
			if cancel != nil {
				cancel() // Prevent context leaks if goroutine exits naturally
			}
			p.mutex.Lock()
			p.state = Stopped
			p.mutex.Unlock()
			p.logger.Debug(fmt.Sprintf("node %s transitioned to stopped", p.name))
			p.wg.Done()
		}()
		if p.process == nil {
			p.logger.Error("no process function defined for node")
			return
		}
		p.logger.Debug(fmt.Sprintf("node %s executing processor", p.name))
		p.process(&ProcessorAPIContext[T, K]{
			nodeState: p,
			ctx:       ctx,
			cancel:    cancel,
		})
	}()

	p.state = Running
	p.logger.Debug(fmt.Sprintf("node %s transitioned to running", p.name))

	return nil
}

// Stop implements [Runner].
func (p *PipelineNode[T, K]) Stop() error {
	// Acquire lock to validate and transition state, and capture cancel func.
	p.mutex.Lock()
	if p.state == Stopped {
		p.mutex.Unlock()
		return nil // Already stopped, return gracefully
	}

	if p.state != Running && p.state != Stopping {
		state := p.state
		p.mutex.Unlock()
		return fmt.Errorf("node %s is not running (state: %v)", p.name, state)
	}

	if p.state == Stopping {
		p.mutex.Unlock()
		p.wg.Wait()
		return nil
	}
	p.state = Stopping
	cancel := p.nodeCancelFunc
	p.mutex.Unlock()
	p.logger.Debug(fmt.Sprintf("node %s transitioning to stopping", p.name))

	if cancel != nil {
		cancel()
	}
	p.wg.Wait()

	p.mutex.Lock()
	p.state = Stopped
	p.mutex.Unlock()
	p.logger.Debug(fmt.Sprintf("node %s stop completed", p.name))

	return nil
}

// Wait implements [Runner].
func (p *PipelineNode[T, K]) Wait() {
	p.wg.Wait()
}

var _ Node = (*PipelineNode[any, any])(nil)

type ProcessorAPI[T any, K any] interface {
	GetNodeName() string
	GetNodeState() RunnerState
	GetNodeLogger() applogger.Logger
	GetNodeContext() context.Context
	GetNodeCancelFunc() context.CancelFunc
	GetParentPipeline() *Pipeline
	GetNodeWaitGroup() *sync.WaitGroup
	Send(data K) error
	SendError(err error) error
	Receive() (data T, ok bool)
}

type ProcessorAPIContext[T any, K any] struct {
	nodeState *PipelineNode[T, K]
	ctx       context.Context
	cancel    context.CancelFunc
}

// GetNodeCancelFunc implements [ProcessorAPI].
func (p *ProcessorAPIContext[T, K]) GetNodeCancelFunc() context.CancelFunc {
	return p.cancel
}

func (p *ProcessorAPIContext[T, K]) GetNodeLogger() applogger.Logger {
	return p.nodeState.logger
}

// GetNodeWaitGroup implements [ProcessorAPI].
func (p *ProcessorAPIContext[T, K]) GetNodeWaitGroup() *sync.WaitGroup {
	return p.nodeState.wg
}

// GetParentPipeline implements [ProcessorAPI].
func (p *ProcessorAPIContext[T, K]) GetParentPipeline() *Pipeline {
	return p.nodeState.parent
}

// GetNodeName implements [ProcessorAPI].
func (p *ProcessorAPIContext[T, K]) GetNodeName() string {
	return p.nodeState.name
}

// GetNodeContext implements [ProcessorAPI].
func (p *ProcessorAPIContext[T, K]) GetNodeContext() context.Context {
	return p.ctx
}

// GetNodeState implements [ProcessorAPI].
func (p *ProcessorAPIContext[T, K]) GetNodeState() RunnerState {
	p.nodeState.mutex.Lock()
	defer p.nodeState.mutex.Unlock()
	return p.nodeState.state
}

// Receive implements [ProcessorAPI].
func (p *ProcessorAPIContext[T, K]) Receive() (data T, ok bool) {
	if p.nodeState.inChan == nil {
		var zero T
		return zero, false
	}
	select {
	case <-p.ctx.Done():
		var zero T
		return zero, false
	case v, ok := <-p.nodeState.inChan:
		return v, ok
	}
}

// Send implements [ProcessorAPI].
func (p *ProcessorAPIContext[T, K]) Send(data K) error {
	if p.nodeState.outChan == nil {
		return fmt.Errorf("output channel is nil")
	}

	select {
	case <-p.ctx.Done():
		return p.ctx.Err()
	case p.nodeState.outChan <- data:
		return nil
	}
}

// SendError implements [ProcessorAPI].
func (p *ProcessorAPIContext[T, K]) SendError(err error) error {
	if p.nodeState.errChan == nil {
		return fmt.Errorf("error channel is nil")
	}
	select {
	case <-p.ctx.Done():
		return p.ctx.Err()
	case p.nodeState.errChan <- err:
		return nil
	}
}

var _ ProcessorAPI[any, any] = (*ProcessorAPIContext[any, any])(nil)

type Processor[T any, K any] func(api ProcessorAPI[T, K])
