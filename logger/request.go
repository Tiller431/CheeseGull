package logger

import (
	"log"

	"github.com/fatih/color"
)

// Request logs Request information. (Like Info, but just for API Requests).
func Request(message string, v ...interface{}) {
	if len(v) < 1 {
		log.Println(prefix(color.GreenString("R")), message)
	} else {
		log.Printf(prefix(color.GreenString("R"))+message+"\n", v)
	}
}
