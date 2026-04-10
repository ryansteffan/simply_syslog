package server

import (
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
	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			api.SendError(err)
			continue
		}

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
				logger.Debug("Received TCP message: " + string(data))
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
