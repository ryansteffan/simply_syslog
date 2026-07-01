package server

import (
	"fmt"
	"net"

	"github.com/ryansteffan/simply_syslog/internal/config"
	"github.com/ryansteffan/simply_syslog/internal/pipeline"
)

type ServerTransferData struct {
	Message []byte
	Meta    map[string]string
}

func UDPServerProcessor(api pipeline.ProcessorAPI[string, ServerTransferData]) {
	conf, err := config.GetConfig()
	logger := api.GetNodeLogger()

	if err != nil {
		api.SendError(err)
	}

	addr, err := net.ResolveUDPAddr("udp", conf.UDPServer.BindAddress+":"+conf.UDPServer.Port)
	if err != nil {
		api.SendError(err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		api.SendError(err)
	}

	buffer := make([]byte, conf.UDPServer.MaxMessageSize)
	logger.Info("UDP server started on " + conf.UDPServer.BindAddress + ":" + conf.UDPServer.Port)
	logger.Debug(fmt.Sprintf("UDP listener ready on %s", conn.LocalAddr()))

	ctx := api.GetNodeContext()
	go func() {
		<-ctx.Done()
		logger.Info("shutting down UDP server")
		conn.Close()
	}()

	for {
		n, addr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			select {
			case <-ctx.Done():
				return
			default:
				api.SendError(err)
				continue
			}
		}
		if n > conf.UDPServer.MaxMessageSize {
			logger.Warn(fmt.Sprintf("received UDP message from %s that exceeds the maximum message size of %d bytes, ignoring", addr.String(), conf.UDPServer.MaxMessageSize))
			continue
		}

		message := make([]byte, n)
		copy(message, buffer[:n])
		logger.Debug(fmt.Sprintf("received UDP message from %s with %d byte(s)", addr.String(), len(message)))
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
