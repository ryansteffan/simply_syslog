package applogger

type LogConfig struct {
	Name    string
	Level   LogLevel
	Handler string
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

func (logConf *LogConfig) Emergency(message string) {
}

var Loggers = make(map[string]*LogConfig)

var defaultLogger = &LogConfig{
	Name:    "DEFAULT",
	Level:   INFO,
	Handler: "console",
}

func GetDefaultLogger() *LogConfig {
	return defaultLogger
}

func CreateLogger(name string, level LogLevel, handler string) {

}
