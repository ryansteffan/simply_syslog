package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/ryansteffan/simply_syslog/internal/buffer"
	"github.com/ryansteffan/simply_syslog/internal/config"
	"github.com/ryansteffan/simply_syslog/internal/parser"
	"github.com/ryansteffan/simply_syslog/internal/pipeline"
	"github.com/ryansteffan/simply_syslog/internal/server"
	"github.com/ryansteffan/simply_syslog/internal/writer"
	"github.com/ryansteffan/simply_syslog/pkg/applogger"
)

func main() {
	// TODO: Add a more robust flag system.
	generateConfigFromEnv := flag.Bool(
		"env-gen-config",
		false,
		"Generate the configuration file from environment variables if it does not exist",
	)

	generateRegexFromEnv := flag.Bool(
		"env-gen-regex",
		false,
		"Generate the regex configuration file from environment variables if it does not exist",
	)

	generateWriterConfigFromEnv := flag.Bool(
		"env-gen-writer-config",
		false,
		"Generate the writer configuration file from environment variables if it does not exist",
	)

	flag.Parse()

	// Ensure that the config directory exists and generate the configs if required.
	if _, err := os.Stat("./config"); os.IsNotExist(err) {
		err := os.Mkdir("./config", 0755)
		if err != nil {
			panic(err)
		}
	}

	err := config.GenerateConfig(*generateConfigFromEnv)
	if err != nil {
		panic(err)
	}

	err = config.GenerateRegexConfig(*generateRegexFromEnv)
	if err != nil {
		panic(err)
	}

	err = config.GenerateWriterConfig(*generateWriterConfigFromEnv)
	if err != nil {
		panic(err)
	}

	CONFIG, err := config.GetConfig()
	if err != nil {
		panic(err)
	}

	logger, err := applogger.NewLogger(
		"Root Logger",
		CONFIG.SelfLoggingLevel,
		applogger.CONSOLE,
		nil,
	)

	if err != nil {
		panic(err)
	}

	logger.Debug(fmt.Sprintf(
		"loaded config: server_mode=%s bind_address=%s udp_port=%s tcp_port=%s tls_port=%s self_logging_level=%d",
		CONFIG.ServerMode,
		CONFIG.BindAddress,
		CONFIG.UDPPort,
		CONFIG.TCPPort,
		CONFIG.TLSPort,
		CONFIG.SelfLoggingLevel,
	))

	logger.Info("starting syslog server")

	pl := pipeline.NewPipeline(context.Background(), logger)

	logger.Debug("created pipeline instance")
	logger.Debug("initializing pipeline channels")

	// TODO: Add proper data types
	errChan := make(chan error, 10)
	serverToParserChan := make(chan server.ServerTransferData, 1000)
	parserToBufferChan := make(chan parser.ParserTransferData, 1000)
	bufferToWriterChan := make(chan buffer.BufferTransferData, 1000)

	// TODO: Make the server type more dynamic based on the config
	serverType := CONFIG.ServerMode
	logger.Debug(fmt.Sprintf("selecting server node for mode %s", serverType))

	var serverNode pipeline.Node
	switch serverType {
	case config.ServerModeUDP:
		serverNode = pipeline.NewPipelineNode(
			"UDP Server",
			logger,
			nil,
			serverToParserChan,
			errChan,
			server.UDPServerProcessor,
		)
	case config.ServerModeTCP:
		serverNode = pipeline.NewPipelineNode(
			"TCP Server",
			logger,
			nil,
			serverToParserChan,
			errChan,
			server.TCPServerProcessor,
		)
	default:
		logger.Critical("unsupported server type: " + string(serverType))
		panic("unsupported server type: " + string(serverType))
	}

	pl.AddNode(serverNode)

	logger.Debug("added server node to pipeline")

	parserNode := pipeline.NewPipelineNode(
		"Parser",
		logger,
		serverToParserChan,
		parserToBufferChan,
		errChan,
		parser.ParserProcessor,
	)

	pl.AddNode(parserNode)

	logger.Debug("added parser node to pipeline")

	bufferNode := pipeline.NewPipelineNode(
		"Buffer",
		logger,
		parserToBufferChan,
		bufferToWriterChan,
		errChan,
		buffer.BufferProcessor,
	)

	pl.AddNode(bufferNode)

	logger.Debug("added buffer node to pipeline")

	writerNode := pipeline.NewPipelineNode(
		"Writer",
		logger,
		bufferToWriterChan,
		nil,
		errChan,
		writer.WriterProcessor,
	)

	pl.AddNode(writerNode)

	logger.Debug("added writer node to pipeline")

	err = pl.Start()
	if err != nil {
		logger.Critical("failed to start pipeline: " + err.Error())
		panic(err)
	}

	logger.Debug("pipeline start returned successfully")

	// Create a error channel sink
	go func() {
		for err := range errChan {
			logger.Error("pipeline error: " + err.Error())
		}
	}()

	logger.Info("pipeline started successfully")

	pl.Wait()
	logger.Debug("pipeline wait completed")
}
