package applogger

type LogLevel int

const (
	EMERGENCY LogLevel = iota
	ALERT              = iota
	CRITICAL           = iota
	ERROR              = iota
	WARN               = iota
	NOTICE             = iota
	INFO               = iota
	DEBUG              = iota
)
