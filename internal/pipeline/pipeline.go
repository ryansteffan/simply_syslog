// Contains the implementation for a generic pipeline processing system.
// Nodes are added sequentially to the pipeline and data flows from
// one node to the next via channels. Each node has a processing function
// that defines its behavior.
package pipeline

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/ryansteffan/simply_syslog/pkg/applogger"
)

// PipelineRunner defines the interface that a node within the
// pipeline must implement to have the required control methods.
type PipelineRunner interface {
	Start() error
	Stop() error
	Restart() error
	GetIsRunning() bool
	ToggleIsRunning()
}

// A generic pipeline node that processes data of type T
// and outputs data of type K.
type PipelineNode[T any, K any] struct {
	// Define the observable properties of the node
	Name       string                        // Name of the node.
	InChannel  chan T                        // Input channel for receiving data of type T.
	OutChannel chan K                        // Output channel for sending data of type K.
	ProcFunc   func(ref *PipelineNode[T, K]) // Processing function for the node.
	// Define the internal behavior of the node
	isRunning bool               // Indicates if the node is currently running.
	stopCtx   context.Context    // Context for stopping the node.
	stopFunc  context.CancelFunc // Function to cancel the stop context.
	mutex     sync.Mutex         // Mutex for thread-safe access to isRunning.
	wg        sync.WaitGroup     // WaitGroup for managing goroutines.
}

// NewPipelineNode creates a new pipeline node instance and returns a pointer to it.
//
// When making a new pipeline node, the caller must provide the following:
//   - name: A string name for the node.
//   - inChan: A channel for receiving data of type T.
//   - outChan: A channel for sending data of type K.
//   - procFunc: A function that takes a reference to the PipelineNode
//     that runs the processing logic.
//
// If the node is not going to pass data or is not taking data in,
// the the respective channel can be set to nil.
//
// WARNING: It is the responsibility of the caller to ensure that the procFunc
// handles nil channels appropriately, as well as shutdown logic.
func NewPipelineNode[T any, K any](
	name string,
	inChan chan T,
	outChan chan K,
	procFunc func(ref *PipelineNode[T, K]),
) *PipelineNode[T, K] {
	stopCtx, stopFunc := context.WithCancel(context.Background())
	return &PipelineNode[T, K]{
		Name:       name,
		InChannel:  inChan,
		OutChannel: outChan,
		ProcFunc:   procFunc,
		isRunning:  false,
		stopCtx:    stopCtx,
		stopFunc:   stopFunc,
	}
}

// GetIsRunning returns the current running state of the pipeline node.
// This operation is thread-safe. It is the only way that isRunning
// should be accessed.
//
// GetIsRunning implements PipelineRunner.
func (p *PipelineNode[T, K]) GetIsRunning() bool {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return p.isRunning
}

// ToggleIsRunning toggles the running state of the pipeline node.
// This operation is thread-safe and as such it should be the only
// way that isRunning is modified.
//
// ToggleIsRunning implements PipelineRunner.
func (p *PipelineNode[T, K]) ToggleIsRunning() {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.isRunning = !p.isRunning
}

// Start begins the execution of the referenced pipeline node.
// If an error occurs, it is returned.
//
// Start implements PipelineRunner.
func (p *PipelineNode[T, K]) Start() error {
	if p.GetIsRunning() {
		return errors.New("pipeline node is already running")
	}
	p.ToggleIsRunning()
	go p.ProcFunc(p)
	return nil
}

// Stops the referenced pipeline node.
// If an error occurs, it is returned.
//
// Stop implements PipelineRunner.
func (p *PipelineNode[T, K]) Stop() error {
	if p.GetIsRunning() {
		p.stopFunc()
		p.ToggleIsRunning()
		return nil
	}
	return errors.New("pipeline node is not running")
}

// Stops the referenced pipeline node and then starts it again.
// The function returns any error encountered during stopping or starting.
//
// Restart implements PipelineRunner.
func (p *PipelineNode[T, K]) Restart() error {
	err := p.Stop()
	if err != nil {
		return err
	}

	err = p.Start()
	if err != nil {
		return err
	}

	return nil
}

// Validates that the PipelineNode struct implements the PipelineRunner interface.
var _ PipelineRunner = (*PipelineNode[any, any])(nil)

// A pipeline represents a collection of PipelineRunner nodes
// that are connected in sequential order.
type Pipeline struct {
	nodes  []PipelineRunner
	logger applogger.Logger
	wg     *sync.WaitGroup
}

// NewPipeline creates a new pipeline instance and returns a pointer to it.
func NewPipeline(wg *sync.WaitGroup, logger applogger.Logger) *Pipeline {
	return &Pipeline{
		nodes:  make([]PipelineRunner, 0),
		logger: logger,
		wg:     wg,
	}
}

// Adds a node to the pipeline.
//
// Note: Nodes should be added in the order
// that they will be executed.
//
// (ie. Node 1 --channel--> Node 2 --channel--> Node 3)
func (p *Pipeline) AddNode(node PipelineRunner) {
	p.nodes = append(p.nodes, node)
}

// Start all of the nodes in the pipeline.
// Nodes are started in reverse order to ensure that
// downstream nodes are ready to receive data from
// upstream nodes.
func (p *Pipeline) Start() error {
	for i := len(p.nodes) - 1; i >= 0; i-- {
		err := p.nodes[i].Start()
		if err != nil {
			p.logger.Error("Error starting pipeline node: " + err.Error())
			return err
		}
		p.logger.Info("Started pipeline node: " + fmt.Sprint(i))
		p.wg.Add(1)
	}
	return nil
}

// Stop all of the nodes in the pipeline.
// Nodes are stopped in order to make sure that the upstream
// nodes are not setting data to downstream nodes.
func (p *Pipeline) Stop() error {
	for i := 0; i < len(p.nodes); i++ {
		err := p.nodes[i].Stop()
		if err != nil {
			p.logger.Error("Error stopping pipeline node: " + err.Error())
			return err
		}
		p.logger.Info("Stopped pipeline node: " + fmt.Sprint(i))
		p.wg.Done()
	}
	return nil
}

// Restart all nodes in the pipeline.
//
// Note: All nodes stop and then all nodes start.
// Each node is not handled individually.
func (p *Pipeline) Restart() {
	p.Stop()
	p.Start()
}

func ExamplePipeline() {

}
