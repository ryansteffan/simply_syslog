package main

import "github.com/ryansteffan/simply_syslog/pkg/applogger"

func main() {
	applogger.CreateLogger("TEST LOGGER", applogger.DEBUG, applogger.CONSOLE)
}
