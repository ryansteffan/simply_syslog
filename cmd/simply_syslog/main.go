package main

import (
	"log/slog"

	"github.com/ryansteffan/simply_syslog/pkg/applogger"
)

func main() {
	applogger.GetAppLogger(slog.LevelInfo)
	slog.Info("sfsadf")
}
