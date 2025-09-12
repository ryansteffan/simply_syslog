package applogger

import (
	"errors"
)

type LogConfig struct {
	Name  string
	Level LogLevel
}

type Logger interface {
	Emergency(message string)
	Alert(message string)
	Critical(message string)
	Error(message string)
	Warn(message string)
	Notice(message string)
	Info(message string)
	Debug(message string)
}

func NewLogger(name string, level LogLevel, handler Handler) (Logger, error) {
	var logger Logger
	var err error

	facility := 5 // Level 5 is the level for syslog.

	config := LogConfig{
		Name:  name,
		Level: level,
	}

	switch handler {
	case FILE:
		logger = &FileLogger{
			Config: config,
		}
		panic("unimplemented")
	case CONSOLE:
		logger, err = NewConsoleLogger(facility, LogConfig{
			Name:  name,
			Level: level,
		})
	}

	if err != nil {
		return nil, errors.New(err.Error())
	}

	return logger, nil
}
