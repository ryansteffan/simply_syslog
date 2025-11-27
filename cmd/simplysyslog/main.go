package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/ryansteffan/simply_syslog/internal/pipeline"
	"github.com/ryansteffan/simply_syslog/pkg/applogger"
)

func main() {
	pipelineLogger, err := applogger.NewConsoleLogger(
		applogger.DEBUG,
		applogger.LogConfig{
			Name:  "Pipeline Logger",
			Level: applogger.DEBUG,
		},
	)

	if err != nil {
		panic(err.Error())
	}

	pipelineWg := sync.WaitGroup{}

	serverPipeline := pipeline.NewPipeline(&pipelineWg, pipelineLogger)

	type OneToTwo struct {
		data string
	}

	oneToTwoChan := make(chan OneToTwo)

	serverPipeline.AddNode(
		pipeline.NewPipelineNode(
			"Node One",
			nil,
			oneToTwoChan,
			func(ref *pipeline.PipelineNode[any, OneToTwo]) {
				iter := 0
				for {
					ref.OutChannel <- OneToTwo{data: "Hello from Node One - Message# " + fmt.Sprint(iter)}
					time.Sleep(time.Second * 1)
					iter++
				}
			},
		),
	)

	serverPipeline.AddNode(
		pipeline.NewPipelineNode(
			"Node Two",
			oneToTwoChan,
			nil,
			func(ref *pipeline.PipelineNode[OneToTwo, any]) {
				for {
					msg := <-ref.InChannel
					pipelineLogger.Info("Node Two received message: " + msg.data)
				}
			},
		),
	)

	_ = serverPipeline.Start()
	pipelineWg.Wait()
}

// package main

// import (
// 	"flag"
// 	"fmt"
// 	"os"
// 	"os/signal"
// 	"strconv"
// 	"sync"

// 	"github.com/ryansteffan/simply_syslog/internal/buffer"
// 	"github.com/ryansteffan/simply_syslog/internal/config"
// 	"github.com/ryansteffan/simply_syslog/internal/server"
// 	"github.com/ryansteffan/simply_syslog/internal/syslog"
// 	"github.com/ryansteffan/simply_syslog/pkg/applogger"
// )

// type Args struct {
// 	UseEnv      bool
// 	UseEnvRegex bool
// 	ShowDebug   bool
// }

// type Channels struct {
// 	SyslogChannel            chan []byte
// 	WriteBufferInputChannel  chan buffer.ParsedSyslogData
// 	WriteBufferOutputChannel chan buffer.ParsedSyslogData
// 	OSSignalChannel          chan os.Signal
// }

// type Servers struct {
// 	UDPServer server.Server
// 	TCPServer server.Server
// }

// var serverWaitGroups sync.WaitGroup

// func main() {

// 	// Parse command line arguments
// 	args := ParseArgs()

// 	conf := LoadConfig(args)

// 	logger := CreateAppLogger(conf)
// 	logger.Info("Initialized logger")
// 	logger.Debug("Logger: " + fmt.Sprintf("%+v", logger))

// 	logger.Info("Loaded configuration from " + conf.FileLocation)

// 	// Create channels for communication between core components
// 	channels := CreateChannels(nil)
// 	logger.Debug("Initialized channels")
// 	logger.Debug("Channels: " + fmt.Sprintf("%+v", channels))

// 	// Register to receive OS signals
// 	signal.Notify(channels.OSSignalChannel)

// 	syslogParser := CreateSyslogParser(logger, args)
// 	logger.Info("Initialized syslog parser")
// 	logger.Debug("Syslog Parser: " + fmt.Sprintf("%+v", syslogParser))

// 	logger.Info(fmt.Sprintf(
// 		"Loaded %d syslog formats from %s",
// 		len(*syslogParser.Formats), "./config/regex.json",
// 	))

// 	writeBuffer := CreateWriteBuffer(conf, logger, channels)
// 	logger.Info("Initialized write buffer")
// 	logger.Debug("Write Buffer: " + fmt.Sprintf("%+v", writeBuffer))
// 	servers := CreateServers(conf, logger, channels.SyslogChannel, syslogParser)
// 	logger.Info("Initialized servers")
// 	logger.Debug("Servers: " + fmt.Sprintf("%+v", servers))

// 	// Start the write buffer
// 	serverWaitGroups.Add(1)
// 	go writeBuffer.StreamReader(&serverWaitGroups)
// 	logger.Info("Started write buffer")

// 	go writeBuffer.MonitorAge(&serverWaitGroups)
// 	logger.Info("Started write buffer age monitor")

// 	// Start the syslog handler
// 	serverWaitGroups.Add(1)
// 	go syslog.HandleSyslogMessages(
// 		syslogParser,
// 		channels.SyslogChannel,
// 		channels.WriteBufferInputChannel,
// 		logger,
// 		&serverWaitGroups,
// 	)
// 	logger.Info("Started syslog handler")

// 	// Start the servers
// 	if servers.UDPServer != nil {
// 		serverWaitGroups.Add(1)
// 		go servers.UDPServer.Start(&serverWaitGroups)
// 		logger.Info("Started UDP server on " + conf.Data.Bind_Address + ":" + conf.Data.Udp_Port)
// 	}
// 	if servers.TCPServer != nil {
// 		serverWaitGroups.Add(1)
// 		go servers.TCPServer.Start(&serverWaitGroups)
// 		logger.Info("Started TCP server on " + conf.Data.Bind_Address + ":" + conf.Data.Tcp_Port)
// 	}

// 	serverWaitGroups.Go(
// 		func() {
// 			defer serverWaitGroups.Done()
// 			signal := <-channels.OSSignalChannel
// 			switch signal {
// 			case os.Interrupt:
// 				logger.Info("Received interrupt signal, shutting down...")
// 				logger.Debug("Stopping PID: " + strconv.Itoa(os.Getpid()))
// 				servers.UDPServer.Stop()
// 				serverWaitGroups.Wait()
// 				os.Exit(0)
// 			// case os.Kill:
// 			// logger.Info("Received kill signal, shutting down...")
// 			// logger.Debug("Stopping PID: " + strconv.Itoa(os.Getpid()))
// 			// serverWaitGroups.Wait()
// 			// os.Exit(0)
// 			default:
// 				logger.Info("Received unknown signal: " + signal.String())
// 				os.Exit(100)
// 			}
// 		})

// 	// Stop the main function for exiting.
// 	serverWaitGroups.Wait()
// }

// func ParseArgs() Args {
// 	useEnvFlag := flag.Bool("env", false, "Load configuration from environment variables.")
// 	useEnvRegexFlag := flag.Bool("env-regex", false, "Load regex patterns from environment variables.")

// 	flag.Parse()

// 	return Args{
// 		UseEnv:      *useEnvFlag,
// 		UseEnvRegex: *useEnvRegexFlag,
// 	}
// }

// func CreateAppLogger(conf *config.Config) applogger.Logger {
// 	if conf.Data.Debug_Messages {
// 		logger, err := applogger.NewLogger("simply-syslog", applogger.DEBUG, applogger.CONSOLE)
// 		if err != nil {
// 			panic(err.Error())
// 		}
// 		return logger
// 	}
// 	logger, err := applogger.NewLogger("simply-syslog", applogger.INFO, applogger.CONSOLE)
// 	if err != nil {
// 		panic(err.Error())
// 	}
// 	return logger
// }

// func LoadConfig(args Args) *config.Config {
// 	if !args.UseEnv {
// 		conf, err := config.LoadConfig("./config/config.json")
// 		if err != nil {
// 			panic(err.Error())
// 		}
// 		return conf
// 	}
// 	conf, err := config.LoadConfig("ENV")
// 	if err != nil {
// 		panic(err.Error())
// 	}
// 	return conf
// }

// func CreateSyslogParser(logger applogger.Logger, args Args) *syslog.EvenDrivenSyslogParser {
// 	if !args.UseEnvRegex {
// 		syslogParser, err := syslog.NewEvenDrivenSyslogParser("./config/regex.json", logger)
// 		if err != nil {
// 			logger.Critical(err.Error())
// 			os.Exit(1)
// 		}
// 		return syslogParser
// 	}
// 	syslogParser, err := syslog.NewEvenDrivenSyslogParser("ENV", logger)
// 	if err != nil {
// 		logger.Critical(err.Error())
// 		os.Exit(1)
// 	}
// 	return syslogParser
// }

// func CreateServers(
// 	conf *config.Config,
// 	logger applogger.Logger,
// 	syslogChannel chan []byte,
// 	syslogParser *syslog.EvenDrivenSyslogParser,
// ) Servers {
// 	switch conf.Data.Protocol {
// 	case "UDP":
// 		server, err := server.NewUDPServer(*conf, logger, syslogChannel, syslogParser)
// 		if err != nil {
// 			logger.Critical(err.Error())
// 			os.Exit(1)
// 		}
// 		return Servers{
// 			UDPServer: server,
// 			TCPServer: nil,
// 		}
// 	case "TCP":
// 		server, err := server.NewTCPServer(*conf, logger, syslogChannel, syslogParser)
// 		if err != nil {
// 			logger.Critical(err.Error())
// 			os.Exit(1)
// 		}
// 		return Servers{
// 			UDPServer: nil,
// 			TCPServer: server,
// 		}
// 	case "BOTH":
// 		udpServer, err := server.NewUDPServer(*conf, logger, syslogChannel, syslogParser)
// 		if err != nil {
// 			logger.Critical(err.Error())
// 			os.Exit(1)
// 		}
// 		tcpServer, err := server.NewTCPServer(*conf, logger, syslogChannel, syslogParser)
// 		if err != nil {
// 			logger.Critical(err.Error())
// 			os.Exit(1)
// 		}
// 		return Servers{
// 			UDPServer: udpServer,
// 			TCPServer: tcpServer,
// 		}
// 	default:
// 		logger.Critical("Unsupported protocol: " + conf.Data.Protocol)
// 		os.Exit(1)
// 		return Servers{
// 			UDPServer: nil,
// 			TCPServer: nil,
// 		}
// 	}
// }

// func CreateWriteBuffer(
// 	conf *config.Config,
// 	logger applogger.Logger,
// 	channels Channels,
// ) *buffer.SyslogWriteBuffer {
// 	return buffer.NewSyslogWriteBuffer(
// 		conf.Data.Buffer_Length,
// 		conf.Data.Buffer_Lifespan,
// 		channels.WriteBufferInputChannel,
// 		channels.WriteBufferOutputChannel,
// 		buffer.WMF,
// 		&conf.Data.Syslog_Path,
// 		logger,
// 	)
// }

// func CreateChannels(bufferLen *int) Channels {
// 	if bufferLen == nil {
// 		return Channels{
// 			SyslogChannel:            make(chan []byte),
// 			WriteBufferInputChannel:  make(chan buffer.ParsedSyslogData),
// 			WriteBufferOutputChannel: make(chan buffer.ParsedSyslogData),
// 			OSSignalChannel:          make(chan os.Signal, 1),
// 		}
// 	}

// 	return Channels{
// 		SyslogChannel:            make(chan []byte, *bufferLen),
// 		WriteBufferInputChannel:  make(chan buffer.ParsedSyslogData, *bufferLen),
// 		WriteBufferOutputChannel: make(chan buffer.ParsedSyslogData, *bufferLen),
// 		OSSignalChannel:          make(chan os.Signal, 1),
// 	}
// }
