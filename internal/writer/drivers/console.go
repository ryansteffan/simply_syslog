package drivers

import (
	"fmt"

	"github.com/ryansteffan/simply_syslog/internal/buffer"
	"github.com/ryansteffan/simply_syslog/internal/config"
)

type ConsoleWriter struct {
	Name    string
	Enabled bool
	Options struct {
		Format string
	}
}

func NewConsoleWriter() *ConsoleWriter {
	return &ConsoleWriter{
		Enabled: false,
	}
}

// IsEnabled implements [Driver].
func (c *ConsoleWriter) IsEnabled() bool {
	return c.Enabled
}

// GetName implements [Driver].
func (c *ConsoleWriter) GetName() string {
	return c.Name
}

// Setup implements [Driver].
func (c *ConsoleWriter) Setup(config config.Writer) error {
	c.Name = config.Name
	c.Enabled = config.Enabled
	if format, ok := config.Options["format"]; ok {
		c.Options.Format = format
	} else {
		panic("logger format option is required for console writer")
	}
	return nil
}

// Write implements [Driver].
func (c *ConsoleWriter) Write(data buffer.BufferTransferData) error {
	if c.Options.Format == "raw" {
		fmt.Println(data.RawMessage)
	}
	return nil
}

var _ Driver = (*ConsoleWriter)(nil)
