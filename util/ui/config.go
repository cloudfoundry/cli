package ui

import "code.cloudfoundry.org/cli/util/configv3"

//go:generate counterfeiter . Config

// Config is the UI configuration.
type Config interface {
	// ColorEnabled enables or disabled color
	ColorEnabled() configv3.ColorSetting
	// Locale is the language to translate the output to
	Locale() string
	// IsTTY returns true when the ui has a TTY
	IsTTY() bool
	// TerminalWidth returns the width of the terminal
	TerminalWidth() int
}
