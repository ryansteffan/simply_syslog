package server

import (
	"fmt"
	"net"

	"github.com/ryansteffan/simply_syslog/internal/config"
	"github.com/ryansteffan/simply_syslog/internal/pipeline"
)

func TCPServerProcessor(api pipeline.ProcessorAPI[string, ServerTransferData]) {
	logger := api.GetNodeLogger()
	CONFIG, err := config.GetConfig()
	if err != nil {
		api.SendError(err)
	}

	addr, err := net.ResolveTCPAddr("tcp", CONFIG.BindAddress+":"+CONFIG.TCPPort)
	if err != nil {
		api.SendError(err)
	}

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		api.SendError(err)
	}

	logger.Info("TCP server started on " + CONFIG.BindAddress + ":" + CONFIG.TCPPort)
	logger.Debug(fmt.Sprintf("TCP listener ready on %s", listener.Addr()))
	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			api.SendError(err)
			continue
		}
		logger.Debug(fmt.Sprintf("accepted TCP connection from %s", conn.RemoteAddr()))

		api.GetNodeWaitGroup().Add(1)
		go func() {
			defer api.GetNodeWaitGroup().Done()
			buffer := make([]byte, 1024)
			for {
				n, err := conn.Read(buffer)
				if err != nil {
					logger.Error("Error reading from TCP connection: " + err.Error())
					return
				}
				data := make([]byte, n)
				copy(data, buffer[:n])
				logger.Debug(fmt.Sprintf("received TCP message from %s with %d byte(s)", conn.RemoteAddr(), len(data)))
				api.Send(
					ServerTransferData{
						Message: data,
						Meta: map[string]string{
							"protocol": "tcp",
						},
					},
				)
			}
		}()
	}
}
