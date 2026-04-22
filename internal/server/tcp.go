package server

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/ryansteffan/simply_syslog/internal/config"
	"github.com/ryansteffan/simply_syslog/internal/pipeline"
)

func TCPServerProcessor(api pipeline.ProcessorAPI[string, ServerTransferData]) {
	panic("TCP Server not implemented")
	logger := api.GetNodeLogger()
	CONFIG, err := config.GetConfig()
	if err != nil {
		api.SendError(err)
	}

	addr, err := net.ResolveTCPAddr("tcp", CONFIG.TCPServer.BindAddress+":"+CONFIG.TCPServer.Port)
	if err != nil {
		api.SendError(err)
	}

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		api.SendError(err)
	}

	ctx := api.GetNodeContext()
	wg := api.GetNodeWaitGroup()

	// HashSet for active connections
	connections := make(map[*net.TCPConn]struct{})
	var connectionsMutex sync.Mutex

	logger.Info("TCP server started on " + CONFIG.TCPServer.BindAddress + ":" + CONFIG.TCPServer.Port)

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		listener.Close()

		connectionsMutex.Lock()
		for conn := range connections {
			conn.Close()
		}
		connectionsMutex.Unlock()

		logger.Info("TCP server shut down")
	}()

	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			select {
			case <-ctx.Done():
				return
			default:
				api.SendError(err)
				continue
			}
		}

		connectionsMutex.Lock()
		connections[conn] = struct{}{}
		connectionsMutex.Unlock()

		wg.Add(1)
		go func(c *net.TCPConn) {
			defer wg.Done()
			defer func() {
				c.Close()
				connectionsMutex.Lock()
				delete(connections, c)
				connectionsMutex.Unlock()
			}()

			// TODO: Finish TCP server, add octet counting
			for {
				c.SetDeadline(time.Now().Add(10 * time.Second))
				buffer := make([]byte, CONFIG.TCPServer.MaxMessageSize)
				n, err := c.Read(buffer)
				if err != nil {
					return
				}
				if n > CONFIG.TCPServer.MaxMessageSize {
					logger.Warn(fmt.Sprintf("received TCP message from %s that exceeds the maximum message size of %d bytes, ignoring", c.RemoteAddr(), CONFIG.TCPServer.MaxMessageSize))
					continue
				}

			}
		}(conn)
	}
}
