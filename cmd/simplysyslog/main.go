package main

import (
	"os"
	"sync"

	"github.com/ryansteffan/simply_syslog/internal/config"
	"github.com/ryansteffan/simply_syslog/internal/server"
	"github.com/ryansteffan/simply_syslog/pkg/applogger"
)

var serverWaitGroups sync.WaitGroup

func main() {
	logger, err := applogger.NewLogger("simply-syslog", applogger.DEBUG, applogger.CONSOLE)
	if err != nil {
		panic(err.Error())
	}

	conf, err := config.LoadConfig("./config/config.json")
	if err != nil {
		logger.Critical(err.Error())
		os.Exit(1)
	}

	udpChannel := make(chan server.ServerChannelMessage)

	udpServer, err := server.NewUDPServer(*conf, logger, udpChannel)

	if err != nil {
		panic(err.Error())
	}

	logger.Info("Starting UDP Server!")
	go udpServer.Start(&serverWaitGroups)

	go func() {
		serverWaitGroups.Wait()
	}()
}
