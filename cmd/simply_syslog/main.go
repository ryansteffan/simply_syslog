package main

import (
	"github.com/ryansteffan/simply_syslog/internal/config"
)

func main() {
	config, err := config.LoadConfig("")
	if err != nil {
		return
	}
}
