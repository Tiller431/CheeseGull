package logger

import (
	"log"

	"github.com/fatih/color"
)

// Error logs an Exception if even.
func Error(message string, v ...interface{}) {
	if len(v) < 1 {
		log.Println(prefix(color.RedString("ERR")), message)
	} else {
		log.Printf(prefix(color.RedString("ERR"))+message+"\n", v)
	}
}
