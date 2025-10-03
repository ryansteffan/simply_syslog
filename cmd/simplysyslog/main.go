package main

import (
	"flag"
	"fmt"
	"os"
	"sync"

	"github.com/ryansteffan/simply_syslog/internal/config"
	"github.com/ryansteffan/simply_syslog/internal/server"
	"github.com/ryansteffan/simply_syslog/internal/syslog"
	"github.com/ryansteffan/simply_syslog/pkg/applogger"
)

var serverWaitGroups sync.WaitGroup

type Args struct {
	UseEnv bool
}

func main() {

	args := ParseArgs()

	logger, err := applogger.NewLogger("simply-syslog", applogger.DEBUG, applogger.CONSOLE)
	if err != nil {
		panic(err.Error())
	}

	var conf *config.Config

	if !args.UseEnv {
		conf, err = config.LoadConfig("./config/config.json")
		if err != nil {
			logger.Critical(err.Error())
			os.Exit(1)
		}
	} else {
		conf, err = config.LoadConfig("ENV")
		if err != nil {
			logger.Critical(err.Error())
			os.Exit(1)
		}
	}

	logger.Info("Loaded config from: " + conf.FileLocation)

	logger.Debug("Config Data: " + fmt.Sprintf("%+v", conf))

	udpChannel := make(chan server.ServerChannelMessage)

	syslogParser, err := syslog.NewEvenDrivenSyslogParser("./config/regex.json")

	if err != nil {
		logger.Critical(err.Error())
		os.Exit(1)
	}

	logger.Info(fmt.Sprintf(
		"Loaded %d syslog formats from %s",
		len(*syslogParser.Formats), "./config/regex.json",
	))

	udpServer, err := server.NewUDPServer(*conf, logger, udpChannel, syslogParser)

	if err != nil {
		panic(err.Error())
	}

	logger.Info("Starting UDP Server...")

	serverWaitGroups.Add(1)
	go udpServer.Start(&serverWaitGroups)

	// Stop the main function for exiting.
	serverWaitGroups.Wait()
}

func ParseArgs() Args {
	useEnvFlag := flag.Bool("env", false, "Load configuration from environment variables.")

	flag.Parse()

	return Args{
		UseEnv: *useEnvFlag,
	}
}
