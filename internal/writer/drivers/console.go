package drivers

import (
	"encoding/json"
	"fmt"

	"github.com/ryansteffan/simply_syslog/internal/buffer"
	"github.com/ryansteffan/simply_syslog/internal/config"
)

type ConsoleFormats string

const (
	ConsoleFormatRaw  ConsoleFormats = "raw"
	ConsoleFormatJSON ConsoleFormats = "json"
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
	switch ConsoleFormats(c.Options.Format) {

	case ConsoleFormatRaw:
		for _, rawMsg := range data.RawMessage {
			fmt.Println(rawMsg)
		}

	case ConsoleFormatJSON:
		output := make(map[string]interface{})
		for _, parsedData := range data.ParsedData {
			for key, value := range parsedData {
				output[key] = value
			}
		}
		jsonBytes, err := json.Marshal(output)
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(jsonBytes))

	default:
		return fmt.Errorf("invalid console format: %s", c.Options.Format)
	}

	return nil
}

var _ Driver = (*ConsoleWriter)(nil)
