package rpc

import "code.cloudfoundry.org/cli/v8/util/ui"

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . CommandParser

// CommandParser interface for parsing commands from arguments
type CommandParser interface {
	ParseCommandFromArgs(*ui.UI, []string) (int, error)
}
