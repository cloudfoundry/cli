package terminal

import "fmt"

type Color uint

const (
	red     Color = 31
	green         = 32
	yellow        = 33
	blue          = 34
	magenta       = 35
	cyan          = 36
)

func Colorize(message string, color Color, bold bool) string {
	attr := 0
	if bold {
		attr = 1
	}

	return fmt.Sprintf("\033[%d;%dm%s\033[0m", attr, color, message)
}

func Yellow(message string) string {
	return Colorize(message, yellow, true)
}

func Red(message string) string {
	return Colorize(message, red, true)
}
