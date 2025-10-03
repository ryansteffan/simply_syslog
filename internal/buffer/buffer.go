// The buffer package contains the generic definition of an aged buffer
// and the code needed to create, and use it.
package buffer

import (
	"errors"
	"fmt"
	"time"
)

/*
Creates a generic buffer with a max size and a max age,
and then returns a pointer to it.
*/
func NewBuffer[T any](maxLength int, maxAge int) *Buffer[T] {
	return &Buffer[T]{
		MaxSize:     maxLength,
		MaxAge:      maxAge,
		CurrentSize: 0,
		CurrentAge:  time.Now(),
		Data:        make([]T, 0),
	}
}

/*
Represents a buffer with data of type T,
that has a max age, and a max length.
*/
type Buffer[T any] struct {
	MaxSize     int
	MaxAge      int
	CurrentSize int
	CurrentAge  time.Time
	Data        []T
}

/*
Adds a new item of type T to the referenced buffer.
*/
func (buff *Buffer[T]) AddItem(item T) error {
	if buff.CurrentSize > buff.MaxSize-1 {
		return errors.New("max buffer size exceeded")
	}
	buff.Data = append(buff.Data, item)
	buff.CurrentSize++
	return nil
}

/*
Sets the data in the buffer referenced to be nil,
and then allocates a fresh slice for new items
to be added.
*/
func (buff *Buffer[T]) ClearBuffer() {
	buff.Data = nil
	buff.Data = make([]T, buff.MaxSize)
}

/*
Prints a string representation of the referenced buffer's data.
*/
func (buff *Buffer[T]) PrintItems() {
	fmt.Println("Data:", buff.Data)
}
