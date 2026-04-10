package parser

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/ryansteffan/simply_syslog/internal/config"
	"github.com/ryansteffan/simply_syslog/internal/pipeline"
	"github.com/ryansteffan/simply_syslog/internal/server"
	"github.com/ryansteffan/simply_syslog/pkg/applogger"
)

type ParserTransferData struct {
	RawMessage string
	ParsedData map[string]string
	Meta       map[string]string
}

func ParserProcessor(api pipeline.ProcessorAPI[server.ServerTransferData, ParserTransferData]) {
	logger := api.GetNodeLogger()
	regexConfig, err := config.GetRegexConfig()
	if err != nil {
		api.SendError(err)
	}

	if logger.GetLogLevel() <= applogger.DEBUG {
		logger.Debug(fmt.Sprintf("regexConfig: %v\n", regexConfig))
	}

	logger.Info("loaded regex patterns")

	compiledRegexes := make(map[string]*regexp.Regexp)
	for _, regex := range regexConfig.Regexes {
		expr, err := regexp.Compile(regex.Pattern)
		if err != nil {
			api.SendError(errors.New(
				"an error compiling the regex pattern: " + regex.Name +
					"with pattern: " + regex.Pattern +
					" error: " + err.Error(),
			))
			continue
		}
		compiledRegexes[regex.Name] = expr
	}
	logger.Info("compiled regex patterns")

	for {
		data, ok := api.Receive()
		if !ok {
			api.SendError(errors.New("an error receiving messages in the parser took place"))
			return
		}
		logger.Debug("Parsing message: " + string(data.Message))
		for name, regex := range compiledRegexes {
			if regex.Match(data.Message) {
				logger.Debug("Message matched regex: " + name)

				parseData := make(map[string]string)

				match := regex.FindStringSubmatch(string(data.Message))
				for i, name := range regex.SubexpNames() {
					if i != 0 && name != "" {
						parseData[name] = match[i]
					}
				}

				api.Send(
					ParserTransferData{
						RawMessage: string(data.Message),
						ParsedData: parseData,
						Meta: map[string]string{
							"protocol": data.Meta["protocol"],
							"regex":    name,
						},
					},
				)
			} else {
				logger.Debug("Message did not match regex: " + name)
			}
		}
	}
}
