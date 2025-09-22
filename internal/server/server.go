package server

import "sync"

type Server interface {
	Start(wg *sync.WaitGroup) error
	Stop() error
	Restart() error
}

type ServerChannelMessage struct {
	Sig  byte
	Data []byte
}
