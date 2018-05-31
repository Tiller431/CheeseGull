package logger

import (
	"log"

	"github.com/Gigamons/cheesegull/config"
	"github.com/fatih/color"
)

// Debug logs a Debug information
func Debug(message string, v ...interface{}) {
	conf := config.Parse()
	if conf.Server.Debug {
		if len(v) < 1 {
			log.Println(prefix(color.YellowString("D")), message)
		} else {
			log.Printf(prefix(color.YellowString("D"))+message+"\n", v...)
		}
	}
}
