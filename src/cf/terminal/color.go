package terminal

import (
	"fmt"
	"runtime"
)

type Color uint

const (
	red     Color = 31
	green         = 32
	yellow        = 33
	blue          = 34
	magenta       = 35
	cyan          = 36
)

func colorize(message string, color Color, bold bool) string {
	if runtime.GOOS == "windows" {
		return message
	}

	attr := 0
	if bold {
		attr = 1
	}

	return fmt.Sprintf("\033[%d;%dm%s\033[0m", attr, color, message)
}

func Yellow(message string) string {
	return colorize(message, yellow, true)
}

func Red(message string) string {
	return colorize(message, red, true)
}

func Green(message string) string {
	return colorize(message, green, true)
}

func Blue(message string) string {
	return colorize(message, blue, true)
}

func Cyan(message string) string {
	return colorize(message, cyan, true)
}

func Magenta(message string) string {
	return colorize(message, magenta, true)
}
