// Package applogger provides a set of simple logging utilities for
// applications, to log via syslog format to various different outputs.
//
// The recommended way to create a logger is via the NewLogger function,
// and specifying the desired log level and handler type. This sets up the
// logger to work with syslog by default.
//
// Handlers that are currently supported are CONSOLE and FILE.
// CONSOLE provides logging to the stdout, while the FILE handler
// writes logs to a specified file on disk.
//
// Example Usage:
//
//	logger, err := applogger.NewLogger("myapp", applogger.INFO, applogger.CONSOLE, nil)
//
//	if err != nil {
//	    panic(err.Error())
//	}
//
//	logger.Info("Application started")
//
// If you wish to implement additional handlers, you can create new types
// that satisfy the Logger interface defined in this package.
package applogger
