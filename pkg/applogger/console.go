package applogger

import (
	"errors"
	"fmt"
	"os"
	"time"
)

// ConsoleLogger defines the structure for a handler that can
// log messages to the console/stdout.
type ConsoleLogger struct {
	Config        LogConfig // The configuration for the logger.
	Hostname      string    // The hostname of the machine running the logger.
	Facility      int       // The syslog facility code used for the logger.
	AppName       string    // The AppName used in the log messages.
	MessageFormat string    // The format string for the log messages. (fmt.Printf syntax)
	DateFormat    string    // The date format string for timestamps. (time package syntax)
}

// NewConsoleLogger creates and returns a new ConsoleLogger instance.
// If an error occurs during creation, it returns an error.
//
// NOTE: It is recommended to use NewLogger to create a console logger,
// as it handles the required setup for syslog.
//
// When making a new console logger, the caller must provide:
//   - facility: The syslog facility code to use for the logger.
//   - config: The LogConfig structure defining the logger's configuration.
func NewConsoleLogger(facility int, config LogConfig) (*ConsoleLogger, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, errors.New("there was an error fetching the hostname from the kernel")
	}
	return &ConsoleLogger{
		Config:        config,
		Hostname:      hostname,
		Facility:      facility,
		AppName:       config.Name,
		MessageFormat: "<%d>%s %s: %s -> %s\n",
		DateFormat:    time.Stamp,
	}, nil
}

// Writes an alert message to the console with a specified message.
//
// Alert implements Logger.
func (c *ConsoleLogger) Alert(message string) {
	if c.Config.LogLevel >= ALERT {
		pri := (c.Facility * 8) + int(ALERT)
		fmt.Printf(c.MessageFormat, pri, time.Now().Format(c.DateFormat), c.Hostname, "ALERT", message)
	}
}

// Writes a critical message to the console with a specified message.
//
// Critical implements Logger.
func (c *ConsoleLogger) Critical(message string) {
	if c.Config.LogLevel >= CRITICAL {
		pri := (c.Facility * 8) + int(CRITICAL)
		fmt.Printf(c.MessageFormat, pri, time.Now().Format(c.DateFormat), c.Hostname, "CRITICAL", message)
	}
}

// Writes a debug message to the console with a specified message.
//
// Debug implements Logger.
func (c *ConsoleLogger) Debug(message string) {
	if c.Config.LogLevel >= DEBUG {
		pri := (c.Facility * 8) + int(DEBUG)
		fmt.Printf(c.MessageFormat, pri, time.Now().Format(c.DateFormat), c.Hostname, "DEBUG", message)
	}
}

// Writes an emergency message to the console with a specified message.
//
// Emergency implements Logger.
func (c *ConsoleLogger) Emergency(message string) {
	if c.Config.LogLevel >= EMERGENCY {
		pri := (c.Facility * 8) + int(EMERGENCY)
		fmt.Printf(c.MessageFormat, pri, time.Now().Format(c.DateFormat), c.Hostname, "EMERGENCY", message)
	}
}

// Writes an error message to the console with a specified message.
//
// Error implements Logger.
func (c *ConsoleLogger) Error(message string) {
	if c.Config.LogLevel >= ERROR {
		pri := (c.Facility * 8) + int(ERROR)
		fmt.Printf(c.MessageFormat, pri, time.Now().Format(c.DateFormat), c.Hostname, "ERROR", message)
	}
}

// Writes an info message to the console with a specified message.
//
// Info implements Logger.
func (c *ConsoleLogger) Info(message string) {
	if c.Config.LogLevel >= INFO {
		pri := (c.Facility * 8) + int(INFO)
		fmt.Printf(c.MessageFormat, pri, time.Now().Format(c.DateFormat), c.Hostname, "INFO", message)
	}
}

// Writes a notice message to the console with a specified message.
//
// Notice implements Logger.
func (c *ConsoleLogger) Notice(message string) {
	if c.Config.LogLevel >= NOTICE {
		pri := (c.Facility * 8) + int(NOTICE)
		fmt.Printf(c.MessageFormat, pri, time.Now().Format(c.DateFormat), c.Hostname, "NOTICE", message)
	}
}

// Writes a warning message to the console with a specified message.
//
// Warn implements Logger.
func (c *ConsoleLogger) Warn(message string) {
	if c.Config.LogLevel >= WARN {
		pri := (c.Facility * 8) + int(WARN)
		fmt.Printf(c.MessageFormat, pri, time.Now().Format(c.DateFormat), c.Hostname, "WARN", message)
	}
}

// Ensure ConsoleLogger implements the Logger interface.
var _ Logger = (*ConsoleLogger)(nil)
