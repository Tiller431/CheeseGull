package logger

import (
	"log"

	"github.com/Gigamons/cheesegull/config"
)

// Debug logs a Debug information
func Debug(message string, v ...interface{}) {
	conf := config.Parse()
	if conf.Server.Debug {
		log.Printf("[D]"+message, v)
	}
}
