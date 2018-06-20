package logger

import "github.com/fatih/color"

func prefix(prefix string) string {
	return color.RedString("[") + prefix + color.RedString("]")
}
