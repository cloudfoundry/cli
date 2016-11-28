package command

import "github.com/jessevdk/go-flags"

// ExtendedCommander extends the go-flags Command interface by forcing a Setup
// function on all commands. This setup function should setup all command
// dependencies.
type ExtendedCommander interface {
	flags.Commander
	Setup(Config, UI) error
}
