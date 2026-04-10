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

	for _, writer := range writerConfig.Writers {
		// Setup each of the writers
		drivers.GetRegisteredWriters()[writer.Name].Setup(writer)
		logger.Info(fmt.Sprintf("enabled writer: %s", writer.Name))
	}

	enabledWriters := make(map[string]drivers.Driver)
	for _, writer := range writerConfig.Writers {
		if drivers.GetRegisteredWriters()[writer.Name].IsEnabled() {
			enabledWriters[writer.Name] = drivers.GetRegisteredWriters()[writer.Name]
		}
	}

	for {
		data, ok := api.Receive()
		if !ok {
			api.SendError(fmt.Errorf("an error receiving messages in the writer took place"))
			return
		}
		logger.Debug("Writing message: " + data.RawMessage)
		for _, writer := range enabledWriters {
			writer.Write(data)
		}
	}
}
