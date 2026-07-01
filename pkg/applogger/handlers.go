package applogger

// Handlers defines the type for different logging handlers.
type Handlers int

const (
	// CONSOLE represents the console logging handler.
	CONSOLE Handlers = iota
	// FILE represents the file logging handler.
	FILE
)
