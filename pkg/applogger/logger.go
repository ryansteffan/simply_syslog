package applogger

import (
	"errors"
)

// LogConfig defines the configuration for a logger.
//
// The LogLevel field determines the minimum level of messages
// that will be logged by the logger.
//
// Ex. If the log level is set to INFO, then only messages with
// severity INFO and above (NOTICE, WARN, ERROR, CRITICAL, ALERT, EMERGENCY)
// will be logged. DEBUG messages will be ignored.
type LogConfig struct {
	Name     string   // The name used to identify the logger.
	LogLevel LogLevel // The minimum level of messages to log.
}

// Logger defines the interface that all logging handlers must implement.
type Logger interface {
	Emergency(message string)
	Alert(message string)
	Critical(message string)
	Error(message string)
	Warn(message string)
	Notice(message string)
	Info(message string)
	Debug(message string)
	GetLogLevel() LogLevel
}

// Creates a new Logger instance based on the specified handler type and then returns it.
// If an error occurs during creation, it returns an error.
//
// It is recommended to use this function to create loggers rather than
// instantiating logger types directly. As this function handles the required
// setup for syslog.
//
// When making a logger you must specify:
//   - name: The name of the application or component using the logger.
//   - level: The minimum LogLevel for messages to be logged.
//   - handler: The type of logging handler to use.
//   - facility: The syslog facility code to use for the logger. If nil, the USER facility is used.
//
// Generally you will not need to specify a facility, unless you have specific
// requirements for logging, and as such it can remain nil.
//
// Example:
//
//	logger, err := applogger.NewLogger("myapp", applogger.INFO, applogger.CONSOLE, nil)
//
//	if err != nil {
//	    panic(err.Error())
//	}
//
//	logger.Info("Application started")
func NewLogger(name string, level LogLevel, handler Handlers, facility *int) (Logger, error) {
	var logger Logger
	var err error

	if facility == nil {
		defaultFacility := 1 // USER facility
		facility = &defaultFacility
	}

	config := LogConfig{
		Name:     name,
		LogLevel: level,
	}

	switch handler {
	case FILE:
		logger = &FileLogger{
			Config: config,
		}
		panic("unimplemented")
	case CONSOLE:
		logger, err = NewConsoleLogger(*facility, LogConfig{
			Name:     name,
			LogLevel: level,
		})
	}

	if err != nil {
		return nil, errors.New(err.Error())
	}

	return logger, nil
}
