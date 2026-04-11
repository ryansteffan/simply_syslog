package writer

import (
	"fmt"

	"github.com/ryansteffan/simply_syslog/internal/buffer"
	"github.com/ryansteffan/simply_syslog/internal/config"
	"github.com/ryansteffan/simply_syslog/internal/pipeline"
	"github.com/ryansteffan/simply_syslog/internal/writer/drivers"
	"github.com/ryansteffan/simply_syslog/pkg/applogger"
)

func WriterProcessor(api pipeline.ProcessorAPI[buffer.BufferTransferData, any]) {
	logger := api.GetNodeLogger()
	writerConfig, err := config.GetWriterConfig()
	if err != nil {
		api.SendError(err)
	}

	if logger.GetLogLevel() >= applogger.DEBUG {
		logger.Debug(fmt.Sprintf("writerConfig: %v\n", writerConfig))
	}

	for name, Writer := range drivers.GetRegisteredWriters() {
		for _, config := range writerConfig.Writers {
			if config.Name == name {
				err := Writer.Setup(config)
				if err != nil {
					logger.Error(fmt.Sprintf("Error setting up writer %s: %v\n", name, err))
				} else {
					logger.Info(fmt.Sprintf("Writer %s set up successfully\n", name))
				}
			}
		}
	}

	for {
		data, ok := api.Receive()
		if !ok {
			api.SendError(fmt.Errorf("an error receiving messages in the writer took place"))
			continue
		}
		for name, Writer := range drivers.GetRegisteredWriters() {
			if Writer.IsEnabled() {
				err := Writer.Write(data)
				if err != nil {
					logger.Error(fmt.Sprintf("Error writing message with writer %s: %v\n", name, err))
				}
			}
		}
		logger.Debug("Writing message: " + data.RawMessage)
	}
}
