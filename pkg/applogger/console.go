package applogger

import (
	"errors"
	"fmt"
	"os"
	"time"
)

type ConsoleLogger struct {
	Config        LogConfig
	Hostname      string
	Facility      int
	AppName       string
	MessageFormat string
	DateFormat    string
}

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
		MessageFormat: "<%d>%s %s: %s\n",
		DateFormat:    time.Stamp,
	}, nil
}

// Alert implements Logger.
func (c *ConsoleLogger) Alert(message string) {
	if c.Config.Level >= ALERT {
		pri := (c.Facility * 8) + ALERT
		fmt.Printf(c.MessageFormat, pri, time.Now().Format(c.DateFormat), c.Hostname, message)
	}
}

// Critical implements Logger.
func (c *ConsoleLogger) Critical(message string) {
	if c.Config.Level >= CRITICAL {
		pri := (c.Facility * 8) + CRITICAL
		fmt.Printf(c.MessageFormat, pri, time.Now().Format(c.DateFormat), c.Hostname, message)
	}
}

// Debug implements Logger.
func (c *ConsoleLogger) Debug(message string) {
	if c.Config.Level >= DEBUG {
		pri := (c.Facility * 8) + DEBUG
		fmt.Printf(c.MessageFormat, pri, time.Now().Format(c.DateFormat), c.Hostname, message)
	}
}

// Emergency implements Logger.
func (c *ConsoleLogger) Emergency(message string) {
	if c.Config.Level >= EMERGENCY {
		pri := (c.Facility * 8) + int(EMERGENCY)
		fmt.Printf(c.MessageFormat, pri, time.Now().Format(c.DateFormat), c.Hostname, message)
	}
}

// Error implements Logger.
func (c *ConsoleLogger) Error(message string) {
	if c.Config.Level >= ERROR {
		pri := (c.Facility * 8) + ERROR
		fmt.Printf(c.MessageFormat, pri, time.Now().Format(c.DateFormat), c.Hostname, message)
	}
}

// Info implements Logger.
func (c *ConsoleLogger) Info(message string) {
	if c.Config.Level >= INFO {
		pri := (c.Facility * 8) + INFO
		fmt.Printf(c.MessageFormat, pri, time.Now().Format(c.DateFormat), c.Hostname, message)
	}
}

// Notice implements Logger.
func (c *ConsoleLogger) Notice(message string) {
	if c.Config.Level >= NOTICE {
		pri := (c.Facility * 8) + NOTICE
		fmt.Printf(c.MessageFormat, pri, time.Now().Format(c.DateFormat), c.Hostname, message)
	}
}

// Warn implements Logger.
func (c *ConsoleLogger) Warn(message string) {
	if c.Config.Level >= WARN {
		pri := (c.Facility * 8) + WARN
		fmt.Printf(c.MessageFormat, pri, time.Now().Format(c.DateFormat), c.Hostname, message)
	}
}

var _ Logger = (*ConsoleLogger)(nil)
