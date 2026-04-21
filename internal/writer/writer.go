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
		logger.Debug(fmt.Sprintf("loaded writer config with %d writer(s)", len(writerConfig.Writers)))
	}

	logger.Debug(fmt.Sprintf("registered writer drivers: %d", len(drivers.GetRegisteredWriters())))
	for name, Writer := range drivers.GetRegisteredWriters() {
		logger.Debug(fmt.Sprintf("evaluating writer driver %q for setup", name))
		configured := false
		for _, config := range writerConfig.Writers {
			if config.Name == name {
				configured = true
				logger.Debug(fmt.Sprintf("setting up writer %q (enabled=%t)", name, config.Enabled))
				err := Writer.Setup(config)
				if err != nil {
					logger.Error(fmt.Sprintf("Error setting up writer %s: %v\n", name, err))
				} else {
					logger.Info(fmt.Sprintf("Writer %s set up successfully\n", name))
				}
			}
		}
		if !configured {
			logger.Debug(fmt.Sprintf("no writer config found for driver %q", name))
		}
	}

	for {
		select {
		case <-api.GetNodeContext().Done():
			logger.Info("shutting down writer")
			return
		default:
			data, ok := api.Receive()
			if !ok {
				api.SendError(fmt.Errorf("an error receiving messages in the writer took place"))
				continue
			}
			logger.Debug(fmt.Sprintf("writer stage received message with %d byte(s)", len(data.RawMessage)))
			for name, Writer := range drivers.GetRegisteredWriters() {
				if Writer.IsEnabled() {
					logger.Debug(fmt.Sprintf("writing message with driver %q", name))
					err := Writer.Write(data)
					if err != nil {
						logger.Error(fmt.Sprintf("Error writing message with writer %s: %v\n", name, err))
					}
				} else {
					logger.Debug(fmt.Sprintf("skipping disabled writer driver %q", name))
				}
			}
			logger.Debug("writer stage completed dispatch for current message")
		}
	}
}
