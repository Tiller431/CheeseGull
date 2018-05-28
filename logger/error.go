package logger

import (
	"log"

	"github.com/fatih/color"
)

// Error logs an Exception if even.
func Error(message string, v ...interface{}) {
	log.Printf(prefix(color.RedString("ERR"))+message, v)
}
