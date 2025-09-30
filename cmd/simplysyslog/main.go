package main

import (
	"flag"
	"fmt"
	"sync"
)

var serverWaitGroups sync.WaitGroup

type Args struct {
	UseEnv bool
}

func main() {

	fmt.Println("Hello World!")

	// args := ParseArgs()

	// logger, err := applogger.NewLogger("simply-syslog", applogger.DEBUG, applogger.CONSOLE)
	// if err != nil {
	// 	panic(err.Error())
	// }

	// var conf *config.Config

	// if !args.UseEnv {
	// 	conf, err = config.LoadConfig("./config/config.json")
	// 	if err != nil {
	// 		logger.Critical(err.Error())
	// 		os.Exit(1)
	// 	}
	// } else {
	// 	conf, err = config.LoadConfig("ENV")
	// 	if err != nil {
	// 		logger.Critical(err.Error())
	// 		os.Exit(1)
	// 	}
	// }

	// logger.Info("Loaded config from: " + conf.FileLocation)

	// udpChannel := make(chan server.ServerChannelMessage)

	// udpServer, err := server.NewUDPServer(*conf, logger, udpChannel)

	// if err != nil {
	// 	panic(err.Error())
	// }

	// logger.Info("Starting UDP Server...")

	// serverWaitGroups.Add(1)
	// go udpServer.Start(&serverWaitGroups)

	// // Stop the main function for exiting.
	// serverWaitGroups.Wait()
}

func ParseArgs() Args {
	useEnvFlag := flag.Bool("env", false, "Load configuration from environment variables.")

	flag.Parse()

	return Args{
		UseEnv: *useEnvFlag,
	}
}
