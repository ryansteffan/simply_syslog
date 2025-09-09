package applogger

type Handler int

const (
	CONSOLE Handler = iota
	FILE    Handler = iota
)
