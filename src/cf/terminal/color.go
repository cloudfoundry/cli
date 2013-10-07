package terminal

import (
	"fmt"
	"regexp"
	"runtime"
)

type Color uint

const (
	red    Color = 31
	green        = 32
	yellow       = 33
	//	blue          = 34
	magenta = 35
	cyan    = 36
	grey    = 37
	white   = 38
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

func decolorize(message string) string {
	reg, err := regexp.Compile(`\x1B\[([0-9]{1,2}(;[0-9]{1,2})?)?[m|K]`)
	if err != nil {
		panic(err)
	}
	return string(reg.ReplaceAll([]byte(message), []byte("")))
}

func HeaderColor(message string) string {
	return colorize(message, white, true)
}

func TableContentColor(message string) string {
	return colorize(message, grey, false)
}

func CommandColor(message string) string {
	return colorize(message, yellow, true)
}

func StartedColor(message string) string {
	return colorize(message, grey, true)
}

func StoppedColor(message string) string {
	return colorize(message, grey, true)
}

func AdvisoryColor(message string) string {
	return colorize(message, yellow, true)
}

func CrashedColor(message string) string {
	return colorize(message, red, true)
}

func FailureColor(message string) string {
	return colorize(message, red, true)
}

func SuccessColor(message string) string {
	return colorize(message, green, true)
}

func EntityNameColor(message string) string {
	return colorize(message, cyan, true)
}

func PromptColor(message string) string {
	return colorize(message, cyan, true)
}

func TableContentHeaderColor(message string) string {
	return colorize(message, cyan, true)
}

func WarningColor(message string) string {
	return colorize(message, magenta, true)
}
