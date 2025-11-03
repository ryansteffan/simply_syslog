package server

import (
	"context"
	"errors"
	"net"
	"sync"

	"github.com/ryansteffan/simply_syslog/internal/config"
	"github.com/ryansteffan/simply_syslog/internal/syslog"
	"github.com/ryansteffan/simply_syslog/pkg/applogger"
)

type UDPSyslogServer struct {
	Conf       config.Config
	Logger     applogger.Logger
	Addr       *net.UDPAddr
	Channel    chan []byte
	Parser     syslog.SyslogParser
	cancleFunc func()
	cancleCtx  context.Context
	stoppped   bool
	mutex      sync.Mutex
}

// NewServer implements Server.
func NewUDPServer(
	conf config.Config,
	logger applogger.Logger,
	channel chan []byte,
	parser syslog.SyslogParser,
) (Server, error) {
	address := conf.Data.Bind_Address + ":" + conf.Data.Udp_Port
	addr, err := net.ResolveUDPAddr("udp", address)

	logger.Info("Created UDP server on " + address)

	if err != nil {
		return nil, errors.New("there was an error resolving the udp server address")
	}

	context, cancleFunc := context.WithCancel(context.Background())

	return &UDPSyslogServer{
		Conf:       conf,
		Logger:     logger,
		Addr:       addr,
		Channel:    channel,
		Parser:     parser,
		cancleFunc: cancleFunc,
		cancleCtx:  context,
	}, nil
}

// start implements Server.
func (u *UDPSyslogServer) Start(wg *sync.WaitGroup) error {
	defer wg.Done()

	u.mutex.Lock()
	u.stoppped = false
	u.mutex.Unlock()

	server, err := net.ListenUDP("udp", u.Addr)
	if err != nil {
		return errors.New("there was an error starting the udp server")
	}

	defer server.Close()

	maxMessageSize := u.Conf.Data.Max_Message_Size

	messageBuff := make([]byte, maxMessageSize)
	for {
		size, addr, err := server.ReadFromUDP(messageBuff)

		if err != nil {
			// addr may be nil on error; avoid nil deref
			if addr != nil {
				u.Logger.Error("There was an error receiving a message to the server from address " + addr.String())
			} else {
				u.Logger.Error("There was an error receiving a message to the server (addr unavailable): " + err.Error())
			}
			continue
		}

		// Copy the bytes before sending to channel to avoid reuse of the underlying buffer
		msg := make([]byte, size)
		copy(msg, messageBuff[:size])

		u.Logger.Debug("Message received: " + string(msg))

		select {
		case <-u.cancleCtx.Done():
			u.Logger.Info("UDP server is stopping, exiting message receive loop")
			return nil
		case u.Channel <- msg:
			u.Logger.Info("Message sent to processing channel from " + addr.String())
		default:
			u.Logger.Warn("Processing channel is full, dropping message from " + addr.String())
		}
	}
}

// stop implements Server.
func (u *UDPSyslogServer) Stop() error {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	u.Logger.Info("Stopping UDP Server...")

	if u.stoppped {
		return errors.New("server is already stopped")
	}
	u.cancleFunc()
	u.stoppped = true
	return nil
}

// restart implements Server.
func (u *UDPSyslogServer) Restart() error {
	panic("unimplemented")
}

var _ Server = (*UDPSyslogServer)(nil)
