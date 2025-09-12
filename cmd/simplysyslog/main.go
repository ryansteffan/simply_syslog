package main

import "github.com/ryansteffan/simply_syslog/pkg/applogger"

func main() {
	logger, err := applogger.NewLogger("simply-syslog", applogger.DEBUG, applogger.CONSOLE)
	if err != nil {
		panic(err.Error())
	}
	logger.Debug("DEBUG")
	logger.Info("INFO")
	logger.Notice("Notice")
	logger.Warn("WARN")
	logger.Critical("Critical")
	logger.Error("Error")
	logger.Emergency("Emergency")
	logger.Alert("ALERT")
}
