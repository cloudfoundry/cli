package ui

import (
	"os"

	"github.com/fatih/color"
	"golang.org/x/crypto/ssh/terminal"
)

const (
	red            color.Attribute = color.FgRed
	green                          = color.FgGreen
	yellow                         = color.FgYellow
	magenta                        = color.FgMagenta
	cyan                           = color.FgCyan
	grey                           = color.FgWhite
	defaultFgColor                 = 38
)

var (
	colorize               func(message string, textColor color.Attribute, bold bool) string
	TerminalSupportsColors = isTerminal()
	UserAskedForColors     = ""
)

// InitColorSupport checks if color is enabled and set the colorize function. The colorize function is used to add color coding to strings.
func InitColorSupport() {
	if colorsEnabled() {
		colorize = func(message string, textColor color.Attribute, bold bool) string {
			colorPrinter := color.New(textColor)
			if bold {
				colorPrinter = colorPrinter.Add(color.Bold)
			}
			f := colorPrinter.SprintFunc()
			return f(message)
		}
	} else {
		colorize = func(message string, _ color.Attribute, _ bool) string {
			return message
		}
	}
}

func colorsEnabled() bool {
	if os.Getenv("CF_COLOR") == "true" {
		return true
	}

	if os.Getenv("CF_COLOR") == "false" {
		return false
	}

	if UserAskedForColors == "true" {
		return true
	}

	return UserAskedForColors != "false" && TerminalSupportsColors
}

func isTerminal() bool {
	return terminal.IsTerminal(int(os.Stdout.Fd()))
}
