package applogger

type FileLogger struct {
	Config     LogConfig
	Path       string
	Buffered   bool
	Buffer     []string
	BufferSize int
}

// GetLogLevel implements [Logger].
func (f *FileLogger) GetLogLevel() LogLevel {
	panic("unimplemented")
}

// Alert implements Logger.
func (f *FileLogger) Alert(message string) {
	panic("unimplemented")
}

// Critical implements Logger.
func (f *FileLogger) Critical(message string) {
	panic("unimplemented")
}

// Debug implements Logger.
func (f *FileLogger) Debug(message string) {
	panic("unimplemented")
}

// Emergency implements Logger.
func (f *FileLogger) Emergency(message string) {
	panic("unimplemented")
}

// Error implements Logger.
func (f *FileLogger) Error(message string) {
	panic("unimplemented")
}

// Info implements Logger.
func (f *FileLogger) Info(message string) {
	panic("unimplemented")
}

// Notice implements Logger.
func (f *FileLogger) Notice(message string) {
	panic("unimplemented")
}

// Warn implements Logger.
func (f *FileLogger) Warn(message string) {
	panic("unimplemented")
}

var _ Logger = (*FileLogger)(nil)
