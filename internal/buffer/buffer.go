// This package provides a generic buffer implementation,
// as well as methods, and functions to interact with it.
package buffer

import (
	"strings"
	"time"
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
	WriteMode  WriteModes
}

func NewWriteBuffer[T any](
	maxSize int,
	maxAge int,
	inChannel chan T,
	outChannel chan T,
	writeMode WriteModes,
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
	}
}

func (b *WriteBuffer[T]) Add(item T) {
	if b.Size < b.MaxSize {
		b.Data = append(b.Data, item)
		b.Size++
	}
}

func (b *WriteBuffer[T]) StreamReader() {
	for message := range b.InChannel {
		b.Add(message)
	}
}

func (b *WriteBuffer[T]) writeHandler() {
	// switch b.WriteMode {
	// case WMF:
	// b.WriteToFile()
	// case WMDB:
	// b.WriteToDB()
	// case WMS:
	// b.StreamWriter()
	// }
	if strings.Contains(string(b.WriteMode), "file") {
	}
	if strings.Contains(string(b.WriteMode), "db") {
	}
	if strings.Contains(string(b.WriteMode), "stream") {
	}
}

func (b *WriteBuffer[T]) StreamWriter() {
	for _, item := range b.Data {
		b.OutChannel <- item
	}
}

func (b *WriteBuffer[T]) MonitorAge() {}

func (b *WriteBuffer[T]) WriteToFile() {}

func (b *WriteBuffer[T]) WriteToDB() {}
