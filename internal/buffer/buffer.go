package buffer

import (
	"sync"
	"time"

	"github.com/ryansteffan/simply_syslog/internal/config"
	"github.com/ryansteffan/simply_syslog/internal/parser"
	"github.com/ryansteffan/simply_syslog/internal/pipeline"
	"github.com/ryansteffan/simply_syslog/pkg/applogger"
)

type BufferTransferData struct {
	RawMessage []string
	ParsedData []map[string]string
	Meta       []map[string]string
}

type Buffer struct {
	RawMessages []string
	ParsedData  []map[string]string
	Meta        []map[string]string
	Logger      applogger.Logger
	Mutex       sync.Mutex

	lastUpdateTime time.Time
	maxLifetime    time.Duration
	currentItems   int
	maxItems       int
}

func NewBuffer(maxItems int, maxLifetime time.Duration, logger applogger.Logger) *Buffer {
	return &Buffer{
		RawMessages: []string{},
		ParsedData:  []map[string]string{},
		Meta:        []map[string]string{},
		Logger:      logger,

		lastUpdateTime: time.Now(),
		maxLifetime:    maxLifetime,
		currentItems:   0,
		maxItems:       maxItems,
	}
}

func (b *Buffer) Add(rawMessage string, parsedData map[string]string, meta map[string]string) {
	b.Mutex.Lock()
	defer b.Mutex.Unlock()
	b.RawMessages = append(b.RawMessages, rawMessage)
	b.ParsedData = append(b.ParsedData, parsedData)
	b.Meta = append(b.Meta, meta)
	b.currentItems++
	b.lastUpdateTime = time.Now()
}

func (b *Buffer) ShouldFlush() bool {
	b.Mutex.Lock()
	defer b.Mutex.Unlock()
	if b.currentItems == 0 {
		return false
	}
	return b.currentItems >= b.maxItems || time.Since(b.lastUpdateTime) >= b.maxLifetime
}

func (b *Buffer) IsEmpty() bool {
	b.Mutex.Lock()
	defer b.Mutex.Unlock()
	return b.currentItems == 0
}

func (b *Buffer) GetData() BufferTransferData {
	b.Mutex.Lock()
	defer b.Mutex.Unlock()
	bufferSnapshot := BufferTransferData{
		RawMessage: append([]string(nil), b.RawMessages...),
		ParsedData: append([]map[string]string(nil), b.ParsedData...),
		Meta:       append([]map[string]string(nil), b.Meta...),
	}
	return bufferSnapshot
}
func (b *Buffer) Reset() {
	b.Mutex.Lock()
	defer b.Mutex.Unlock()
	b.RawMessages = []string{}
	b.ParsedData = []map[string]string{}
	b.Meta = []map[string]string{}
	b.currentItems = 0
	b.lastUpdateTime = time.Now()
}

func BufferProcessor(api pipeline.ProcessorAPI[parser.ParserTransferData, BufferTransferData]) {
	logger := api.GetNodeLogger()
	logger.Info("Buffer processor started")
	conf, err := config.GetConfig()
	if err != nil {
		api.SendError(err)
		return
	}

	bufferLifetime := time.Duration(conf.BufferMaxLifetime) * time.Second
	buffer := NewBuffer(conf.BufferMaxItems, bufferLifetime, logger)
	ctx := api.GetNodeContext()
	processorInputChan := api.GetInputChannel()
	if processorInputChan == nil {
		logger.Error("buffer processor input channel is nil")
		return
	}

	checkDuration := bufferLifetime / 6
	if conf.BufferMaxLifetime <= 0 || checkDuration <= 0 {
		checkDuration = 1 * time.Second
	}
	ticker := time.NewTicker(checkDuration)
	defer ticker.Stop()

	handleBufferFlush := func(reason string) {
		if !buffer.ShouldFlush() {
			return
		}

		logger.Info(reason)
		bufferData := buffer.GetData()
		err := api.Send(bufferData)
		if err != nil {
			errorMessage := "failed to send buffer data: " + err.Error()
			logger.Error(errorMessage)
			return
		}

		buffer.Reset()
	}

	for {
		select {
		// Handle custom receive logic, see pipeline.ProcessorAPI.Receive() for generic implementation
		case <-ctx.Done():
			logger.Info("buffer processor received shutdown signal")
			return
		case data, ok := <-processorInputChan:
			if !ok {
				logger.Info("buffer processor input channel closed")
				return
			}
			buffer.Add(data.RawMessage, data.ParsedData, data.Meta)
			logger.Debug("added message to buffer")
			handleBufferFlush("buffer reached flush conditions, sending data downstream")
		case <-ticker.C:
			logger.Debug("buffer ticker ticked, checking flush conditions")
			handleBufferFlush("buffer timer reached flush conditions, sending data downstream")
		}
	}
}
