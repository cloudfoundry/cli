package ui

import (
	. "github.com/cloudfoundry/bosh-cli/ui/table"
)

type UI interface {
	ErrorLinef(pattern string, args ...interface{})
	PrintLinef(pattern string, args ...interface{})

	BeginLinef(pattern string, args ...interface{})
	EndLinef(pattern string, args ...interface{})

	PrintBlock([]byte) // takes []byte to avoid string copy
	PrintErrorBlock(string)

	PrintTable(Table)

	AskForText(label string) (string, error)
	AskForChoice(label string, options []string) (int, error)
	AskForPassword(label string) (string, error)

	// AskForConfirmation returns error if user doesnt want to continue
	AskForConfirmation() error

	IsInteractive() bool

	Flush()
}
