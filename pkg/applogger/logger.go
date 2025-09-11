package applogger

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

func CreateLogger(name string, level LogLevel, handler Handler) Logger {
	var logger Logger

	config := LogConfig{
		Name:  name,
		Level: level,
	}

	switch handler {
	case FILE:
		logger = &FileLogger{
			Config: config,
		}
	case CONSOLE:
		logger = &ConsoleLogger{
			Config: config,
		}
	}

	return logger
}
