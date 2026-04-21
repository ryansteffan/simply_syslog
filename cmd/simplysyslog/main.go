package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

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
		"loaded config: udp_bind_address=%s udp_port=%s tcp_bind_address=%s tcp_port=%s tls_bind_address=%s tls_port=%s self_logging_level=%d",
		CONFIG.UDPServer.BindAddress,
		CONFIG.UDPServer.Port,
		CONFIG.TCPServer.BindAddress,
		CONFIG.TCPServer.Port,
		CONFIG.TLSServer.BindAddress,
		CONFIG.TLSServer.Port,
		CONFIG.SelfLoggingLevel,
	))

	logger.Info("starting syslog server")

	pl := pipeline.NewPipeline(context.Background(), logger)

	logger.Debug("created pipeline instance")
	logger.Debug("initializing pipeline channels")

	errChan := make(chan error, 10)
	serverToParserChan := make(chan server.ServerTransferData, 1000)
	parserToBufferChan := make(chan parser.ParserTransferData, 1000)
	bufferToWriterChan := make(chan buffer.BufferTransferData, 1000)

	logger.Debug("selecting server node based on config")

	notificationChan := make(chan os.Signal, 1)
	signal.Notify(notificationChan, os.Interrupt, syscall.SIGTERM)

	if CONFIG.UDPServer.Enabled {
		logger.Debug("UDP server enabled in config, adding UDP server node to pipeline")
		udpServerNode := pipeline.NewPipelineNode(
			"UDP Server",
			logger,
			nil,
			serverToParserChan,
			errChan,
			server.UDPServerProcessor,
		)
		pl.AddNode(udpServerNode)
	}

	if CONFIG.TCPServer.Enabled {
		logger.Debug("TCP server enabled in config, adding TCP server node to pipeline")
		tcpServerNode := pipeline.NewPipelineNode(
			"TCP Server",
			logger,
			nil,
			serverToParserChan,
			errChan,
			server.TCPServerProcessor,
		)
		pl.AddNode(tcpServerNode)
	}

	if CONFIG.TLSServer.Enabled {
		logger.Debug("TLS server enabled in config, adding TLS server node to pipeline")
		tlsServerNode := pipeline.NewPipelineNode(
			"TLS Server",
			logger,
			nil,
			serverToParserChan,
			errChan,
			server.TLSServerProcessor,
		)
		pl.AddNode(tlsServerNode)
	}

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

	// Handle graceful shutdown on interrupt signal
	go func() {
		<-notificationChan
		logger.Info("interrupt signal received, shutting down pipeline...")
		pl.Stop()
	}()

	pl.Wait()
	logger.Debug("pipeline wait completed")
}
