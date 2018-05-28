package logger

import (
	"log"

	"github.com/fatih/color"
)

// Info logs information.
func Info(message string, v ...interface{}) {
	log.Printf(prefix(color.CyanString("I"))+message, v)
}
