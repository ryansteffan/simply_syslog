package server

import (
	"net"

	"github.com/ryansteffan/simply_syslog/internal/config"
	"github.com/ryansteffan/simply_syslog/internal/pipeline"
)

type ServerTransferData struct {
	Message []byte
	Meta    map[string]string
}

func UDPServerProcessor(api pipeline.ProcessorAPI[string, ServerTransferData]) {
	CONFIG, err := config.GetConfig()
	logger := api.GetNodeLogger()

	if err != nil {
		api.SendError(err)
	}

	addr, err := net.ResolveUDPAddr("udp", CONFIG.BindAddress+":"+CONFIG.UDPPort)
	if err != nil {
		api.SendError(err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		api.SendError(err)
	}

	buffer := make([]byte, 1024)
	logger.Info("UDP server started on " + CONFIG.BindAddress + ":" + CONFIG.UDPPort)
	for {
		n, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			api.SendError(err)
			continue
		}

		message := make([]byte, n)
		copy(message, buffer[:n])
		logger.Debug("Received UDP message: " + string(message))
		api.Send(
			ServerTransferData{
				Message: message,
				Meta: map[string]string{
					"protocol": "udp",
				},
			},
		)
	}
}
