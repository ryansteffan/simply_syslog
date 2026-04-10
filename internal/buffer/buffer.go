package buffer

import (
	"errors"

	"github.com/ryansteffan/simply_syslog/internal/parser"
	"github.com/ryansteffan/simply_syslog/internal/pipeline"
)

func BufferProcessor(api pipeline.ProcessorAPI[parser.ParserTransferData, string]) {
	logger := api.GetNodeLogger()
	logger.Info("Buffer processor started")
	for {
		data, ok := api.Receive()
		if !ok {
			api.SendError(errors.New("an error receiving messages in the buffer took place"))
			return
		}
		logger.Debug("Buffering message: " + data.RawMessage)
		// TODO: Add buffering logic here, for now pass through this node.
		api.Send(data.RawMessage)
	}
}
