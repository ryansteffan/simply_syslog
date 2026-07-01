package drivers

import (
	"github.com/ryansteffan/simply_syslog/internal/buffer"
	"github.com/ryansteffan/simply_syslog/internal/config"
)

type Driver interface {
	GetName() string
	IsEnabled() bool
	Write(data buffer.BufferTransferData) error
	Setup(config config.Writer) error
}

var registeredWriters = map[string]Driver{
	"console": NewConsoleWriter(),
	"file":    NewFileWriter(),
	"duckdb":  NewDuckDBWriter(),
}

func GetRegisteredWriters() map[string]Driver {
	return registeredWriters
}
