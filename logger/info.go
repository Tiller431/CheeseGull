package logger

import (
	"log"

	"github.com/fatih/color"
)

// Info logs information.
func Info(message string, v ...interface{}) {
	if len(v) < 1 {
		log.Println(prefix(color.CyanString("I")), message)
	} else {
		log.Printf(prefix(color.CyanString("I"))+message+"\n", v)
	}
}
