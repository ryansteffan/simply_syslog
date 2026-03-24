package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"sync"

	"github.com/ryansteffan/simply_syslog/internal/buffer"
	"github.com/ryansteffan/simply_syslog/internal/config"
	"github.com/ryansteffan/simply_syslog/internal/pipeline"
	"github.com/ryansteffan/simply_syslog/internal/server"
	"github.com/ryansteffan/simply_syslog/internal/syslog"
	"github.com/ryansteffan/simply_syslog/pkg/applogger"
)

type Args struct {
	UseEnv      bool
	UseEnvRegex bool
	ShowDebug   bool
}

type Channels struct {
	SyslogChannel            chan []byte
	WriteBufferInputChannel  chan buffer.ParsedSyslogData
	WriteBufferOutputChannel chan buffer.ParsedSyslogData
	OSSignalChannel          chan os.Signal
}

type Servers struct {
	UDPServer server.Server
	TCPServer server.Server
}

var serverWaitGroups sync.WaitGroup

func main() {

	// Parse command line arguments
	args := ParseArgs()

	conf := LoadConfig(args)

	logger := CreateAppLogger(conf)
	logger.Info("Initialized logger")
	logger.Debug("Logger: " + fmt.Sprintf("%+v", logger))

	logger.Info("Loaded configuration from " + conf.FileLocation)

	// Create channels for communication between core components
	channels := CreateChannels(nil)
	logger.Debug("Initialized channels")
	logger.Debug("Channels: " + fmt.Sprintf("%+v", channels))

	// Register to receive OS signals
	signal.Notify(channels.OSSignalChannel)

	// Create the Pipeline
	serverPipeline := pipeline.NewPipeline(&serverWaitGroups, logger)

	switch conf.Data.Protocol {
	case "UDP":
		// Create the UDP Server Node
		serverPipeline.AddNode(
			pipeline.NewPipelineNode(
				"UDP Server",
				nil,
				channels.SyslogChannel,
				func(inChannel chan any, outChannel chan []byte, stopCtx context.Context) {
					// Handle UDP Server
				},
			),
		)

	case "TCP":
		// Create the TCP Server Node
		serverPipeline.AddNode(
			pipeline.NewPipelineNode(
				"TCP Server",
				nil,
				channels.SyslogChannel,
				func(inChannel chan any, outChannel chan []byte, stopCtx context.Context) {
					// Handle TCP Server
				},
			),
		)

	case "BOTH":
		// Create the UDP Server Node
		serverPipeline.AddNode(
			pipeline.NewPipelineNode(
				"UDP Server",
				nil,
				channels.SyslogChannel,
				func(inChannel chan any, outChannel chan []byte, stopCtx context.Context) {
					// Handle UDP Server
				},
			),
		)

		// Create the TCP Server Node
		serverPipeline.AddNode(
			pipeline.NewPipelineNode(
				"TCP Server",
				nil,
				channels.SyslogChannel,
				func(inChannel chan any, outChannel chan []byte, stopCtx context.Context) {
					// Handle TCP Server
				},
			),
		)

	default:
		logger.Critical("Unsupported protocol: " + conf.Data.Protocol)
		os.Exit(1)
	}

	// Create the Syslog Parser Node
	serverPipeline.AddNode(
		pipeline.NewPipelineNode(
			"Syslog Parser",
			channels.SyslogChannel,
			channels.WriteBufferInputChannel,
			func(inChannel chan []byte, outChannel chan buffer.ParsedSyslogData, stopCtx context.Context) {
				// Handle Syslog Parsing
			},
		),
	)

	// Create the Write Buffer Node
	serverPipeline.AddNode(
		pipeline.NewPipelineNode(
			"Write Buffer",
			channels.WriteBufferInputChannel,
			channels.WriteBufferOutputChannel,
			func(inChannel chan buffer.ParsedSyslogData, outChannel chan buffer.ParsedSyslogData, stopCtx context.Context) {
				// Handle Write Buffering
			},
		),
	)

	// Start the Pipeline
	serverPipeline.Start()
	logger.Info("Started server pipeline")

	serverWaitGroups.Go(
		func() {
			signal := <-channels.OSSignalChannel
			switch signal {
			case os.Interrupt:
				logger.Info("Received interrupt signal, shutting down...")
				logger.Debug("Stopping PID: " + strconv.Itoa(os.Getpid()))
				serverPipeline.Stop()
			}
		})

	// Stop the main function from exiting.
	serverWaitGroups.Wait()
	os.Exit(0)
}

func ParseArgs() Args {
	useEnvFlag := flag.Bool("env", false, "Load configuration from environment variables.")
	useEnvRegexFlag := flag.Bool("env-regex", false, "Load regex patterns from environment variables.")

	flag.Parse()

	return Args{
		UseEnv:      *useEnvFlag,
		UseEnvRegex: *useEnvRegexFlag,
	}
}

func CreateAppLogger(conf *config.Config) applogger.Logger {
	facility := 5

	if conf.Data.Debug_Messages {
		logger, err := applogger.NewLogger("simply-syslog", applogger.DEBUG, applogger.CONSOLE, &facility)
		if err != nil {
			panic(err.Error())
		}
		return logger
	}
	logger, err := applogger.NewLogger("simply-syslog", applogger.INFO, applogger.CONSOLE, &facility)
	if err != nil {
		panic(err.Error())
	}
	return logger
}

func LoadConfig(args Args) *config.Config {
	if !args.UseEnv {
		conf, err := config.LoadConfig("./config/config.json")
		if err != nil {
			panic(err.Error())
		}
		return conf
	}
	conf, err := config.LoadConfig("ENV")
	if err != nil {
		panic(err.Error())
	}
	return conf
}

func CreateSyslogParser(logger applogger.Logger, args Args) *syslog.EvenDrivenSyslogParser {
	if !args.UseEnvRegex {
		syslogParser, err := syslog.NewEvenDrivenSyslogParser("./config/regex.json", logger)
		if err != nil {
			logger.Critical(err.Error())
			os.Exit(1)
		}
		return syslogParser
	}
	syslogParser, err := syslog.NewEvenDrivenSyslogParser("ENV", logger)
	if err != nil {
		logger.Critical(err.Error())
		os.Exit(1)
	}
	return syslogParser
}

func CreateServers(
	conf *config.Config,
	logger applogger.Logger,
	syslogChannel chan []byte,
	syslogParser *syslog.EvenDrivenSyslogParser,
) Servers {
	switch conf.Data.Protocol {
	case "UDP":
		server, err := server.NewUDPServer(*conf, logger, syslogChannel, syslogParser)
		if err != nil {
			logger.Critical(err.Error())
			os.Exit(1)
		}
		return Servers{
			UDPServer: server,
			TCPServer: nil,
		}
	case "TCP":
		server, err := server.NewTCPServer(*conf, logger, syslogChannel, syslogParser)
		if err != nil {
			logger.Critical(err.Error())
			os.Exit(1)
		}
		return Servers{
			UDPServer: nil,
			TCPServer: server,
		}
	case "BOTH":
		udpServer, err := server.NewUDPServer(*conf, logger, syslogChannel, syslogParser)
		if err != nil {
			logger.Critical(err.Error())
			os.Exit(1)
		}
		tcpServer, err := server.NewTCPServer(*conf, logger, syslogChannel, syslogParser)
		if err != nil {
			logger.Critical(err.Error())
			os.Exit(1)
		}
		return Servers{
			UDPServer: udpServer,
			TCPServer: tcpServer,
		}
	default:
		logger.Critical("Unsupported protocol: " + conf.Data.Protocol)
		os.Exit(1)
		return Servers{
			UDPServer: nil,
			TCPServer: nil,
		}
	}
}

func CreateWriteBuffer(
	conf *config.Config,
	logger applogger.Logger,
	channels Channels,
) *buffer.SyslogWriteBuffer {
	return buffer.NewSyslogWriteBuffer(
		conf.Data.Buffer_Length,
		conf.Data.Buffer_Lifespan,
		channels.WriteBufferInputChannel,
		channels.WriteBufferOutputChannel,
		buffer.WMF,
		&conf.Data.Syslog_Path,
		logger,
	)
}

func CreateChannels(bufferLen *int) Channels {
	if bufferLen == nil {
		return Channels{
			SyslogChannel:            make(chan []byte),
			WriteBufferInputChannel:  make(chan buffer.ParsedSyslogData),
			WriteBufferOutputChannel: make(chan buffer.ParsedSyslogData),
			OSSignalChannel:          make(chan os.Signal, 1),
		}
	}

	return Channels{
		SyslogChannel:            make(chan []byte, *bufferLen),
		WriteBufferInputChannel:  make(chan buffer.ParsedSyslogData, *bufferLen),
		WriteBufferOutputChannel: make(chan buffer.ParsedSyslogData, *bufferLen),
		OSSignalChannel:          make(chan os.Signal, 1),
	}
}
