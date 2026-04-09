package parser

import (
	"errors"

	"github.com/ryansteffan/simply_syslog/internal/pipeline"
	"github.com/ryansteffan/simply_syslog/internal/server"
)

func ParserProcessor(api pipeline.ProcessorAPI[server.ServerTransferData, string]) {
	logger := api.GetNodeLogger()
	for {
		data, ok := api.Receive()
		if !ok {
			api.SendError(errors.New("an error receiving messages in the parser took place"))
			return
		}
		logger.Debug("Parsing message: " + string(data.Message))
	}
}
