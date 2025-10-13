package server

import (
	"errors"
	"net"
	"sync"

	"github.com/ryansteffan/simply_syslog/internal/config"
	"github.com/ryansteffan/simply_syslog/internal/syslog"
	"github.com/ryansteffan/simply_syslog/pkg/applogger"
)

type UDPSyslogServer struct {
	Conf    config.Config
	Logger  applogger.Logger
	Addr    *net.UDPAddr
	Channel chan []byte
	Parser  syslog.SyslogParser
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

	return &UDPSyslogServer{
		Conf:    conf,
		Logger:  logger,
		Addr:    addr,
		Channel: channel,
		Parser:  parser,
	}, nil
}

// start implements Server.
func (u *UDPSyslogServer) Start(wg *sync.WaitGroup) error {
	defer wg.Done()

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

		u.Channel <- msg
	}
}

// stop implements Server.
func (u *UDPSyslogServer) Stop() error {
	panic("unimplemented")
}

// restart implements Server.
func (u *UDPSyslogServer) Restart() error {
	panic("unimplemented")
}

var _ Server = (*UDPSyslogServer)(nil)
