package applogger

// LogLevel defines the severity level for logging messages.
type LogLevel int

const (
	// EMERGENCY represents the highest severity level for logging.
	EMERGENCY LogLevel = iota
	// ALERT represents a high severity level for logging.
	ALERT
	// CRITICAL represents a critical severity level for logging.
	CRITICAL
	// ERROR represents an error severity level for logging.
	ERROR
	// WARN represents a warning severity level for logging.
	WARN
	// NOTICE represents a notice severity level for logging.
	NOTICE
	// INFO represents an informational severity level for logging.
	INFO
	// DEBUG represents a debug severity level for logging.
	DEBUG
)
