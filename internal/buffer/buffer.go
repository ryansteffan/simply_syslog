// This package provides a generic buffer implementation,
// as well as methods, and functions to interact with it.
package buffer

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/ryansteffan/simply_syslog/pkg/applogger"
)

type WriteModes string

const (
	// Write to file mode.
	WMF WriteModes = "file"
	// Write to database mode.
	WMDB WriteModes = "db"
	// Write to stream mode.
	WMS WriteModes = "stream"
)

type WriteBuffer[T any] struct {
	Data       []T
	MaxSize    int
	Size       int
	MaxAge     int
	CurrentAge time.Time
	InChannel  chan T
	OutChannel chan T
	logger     applogger.Logger
	WriteMode  WriteModes
}

func NewWriteBuffer[T any](
	maxSize int,
	maxAge int,
	inChannel chan T,
	outChannel chan T,
	writeMode WriteModes,
	logger applogger.Logger,
) *WriteBuffer[T] {

	return &WriteBuffer[T]{
		Data:       make([]T, 0, maxSize),
		MaxSize:    maxSize,
		Size:       0,
		MaxAge:     maxAge,
		CurrentAge: time.Now(),
		InChannel:  inChannel,
		OutChannel: outChannel,
		WriteMode:  writeMode,
		logger:     logger,
	}
}

func (b *WriteBuffer[T]) Add(item T) error {
	if b.Size < b.MaxSize {
		b.Data = append(b.Data, item)
		b.Size++
		return nil
	}
	return errors.New("buffer overflow: max size reached")
}

func (b *WriteBuffer[T]) StreamReader(logger applogger.Logger, wg *sync.WaitGroup) {
	defer wg.Done()
	for message := range b.InChannel {
		if b.Size < b.MaxSize {
			b.Add(message)
		} else {
			logger.Info("Buffer full, writing to output...")
			b.writeHandler()
			b.Add(message)
		}
	}
}

func (b *WriteBuffer[T]) Flush() {
	b.Data = make([]T, 0, b.MaxSize)
	b.Size = 0
}

func (b *WriteBuffer[T]) writeHandler() {
	if strings.Contains(string(b.WriteMode), "file") {
		b.WriteToFile()
	}
	if strings.Contains(string(b.WriteMode), "db") {
		b.WriteToDB()
	}
	if strings.Contains(string(b.WriteMode), "stream") {
		b.StreamWriter()
	}
	b.Flush()
}

func (b *WriteBuffer[T]) StreamWriter() {
	for _, item := range b.Data {
		b.OutChannel <- item
	}
}

func (b *WriteBuffer[T]) MonitorAge() {}

func (b *WriteBuffer[T]) WriteToFile() {
	file, err := os.OpenFile("/var/log/syslog.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	b.logger.Debug("Opened file for writing...")
	if err != nil {
		b.logger.Error("Error opening file: " + err.Error())
		return
	}

	defer file.Close()

	for _, item := range b.Data {
		_, err := file.WriteString(fmt.Sprintf("%v\n", item))
		if err != nil {
			b.logger.Error("Error writing to file: " + err.Error())
		}
	}
}

func (b *WriteBuffer[T]) WriteToDB() {
	for _, item := range b.Data {
		b.logger.Debug("Writing to database: " + fmt.Sprintf("%v", item))
	}
}
