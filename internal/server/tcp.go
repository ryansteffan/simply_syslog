package server

import (
	"errors"
	"io"
	"net"
	"sync"

	"github.com/ryansteffan/simply_syslog/internal/config"
	"github.com/ryansteffan/simply_syslog/internal/syslog"
	"github.com/ryansteffan/simply_syslog/pkg/applogger"
)

type TCPSyslogServer struct {
	Conf     config.Config
	Logger   applogger.Logger
	Addr     *net.TCPAddr
	Channel  chan []byte
	Parser   syslog.SyslogParser
	listener *net.TCPListener
	stopped  bool
	mutex    sync.Mutex
}

func NewTCPServer(
	conf config.Config,
	logger applogger.Logger,
	channel chan []byte,
	parser syslog.SyslogParser,
) (Server, error) {
	address := conf.Data.Bind_Address + ":" + conf.Data.Tcp_Port
	addr, err := net.ResolveTCPAddr("tcp", address)

	logger.Info("Created TCP server on " + address)

	if err != nil {
		return nil, errors.New("there was an error resolving the tcp server address")
	}

	return &TCPSyslogServer{
		Conf:    conf,
		Logger:  logger,
		Addr:    addr,
		Channel: channel,
		Parser:  parser,
	}, nil
}

// Start implements Server.
func (t *TCPSyslogServer) Start(wg *sync.WaitGroup) error {
	defer wg.Done()

	listener, err := net.ListenTCP("tcp", t.Addr)

	if err != nil {
		return err
	}

	t.mutex.Lock()
	t.listener = listener
	t.stopped = false
	t.mutex.Unlock()

	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			t.mutex.Lock()
			isStopped := t.stopped
			t.mutex.Unlock()
			if isStopped {
				return nil
			}
			t.Logger.Error("Error accepting connection: " + err.Error())
			continue
		}
		go t.handleConnection(conn)
	}
}

func (t *TCPSyslogServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	for {

		// Read data from the connection
		buffer := make([]byte, t.Conf.Data.Max_Message_Size)
		size, err := conn.Read(buffer)

		// Handle connection closure
		if err == io.EOF {
			t.Logger.Info("Connection closed by client")
			return
		}

		if err != nil {
			t.Logger.Error("Error reading from connection: " + err.Error())
			return
		}

		msg := make([]byte, size)
		copy(msg, buffer[:size])

		t.Logger.Debug("Message received: " + string(msg))

		t.Channel <- msg
	}
}

// Stop implements Server.
func (t *TCPSyslogServer) Stop() error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.Logger.Info("Stopping TCP Server...")

	if t.stopped {
		return errors.New("server is already stopped")
	}

	t.stopped = true
	if t.listener != nil {
		t.listener.Close()
	}
	return nil
}

// Restart implements Server.
func (t *TCPSyslogServer) Restart() error {
	err := t.Stop()
	if err != nil {
		return err
	}
	// Start is a blocking function, so launch it in a goroutine.
	// The WaitGroup is local and satisfies Start's signature (Start calls defer wg.Done()).
	var wg sync.WaitGroup
	wg.Add(1)
	go t.Start(&wg)
	return nil
}

var _ Server = (*TCPSyslogServer)(nil)
