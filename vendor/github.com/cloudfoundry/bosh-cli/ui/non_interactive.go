package ui

import (
	. "github.com/cloudfoundry/bosh-cli/ui/table"
)

type nonInteractiveUI struct {
	parent UI
}

func NewNonInteractiveUI(parent UI) UI {
	return &nonInteractiveUI{parent: parent}
}

func (ui *nonInteractiveUI) ErrorLinef(pattern string, args ...interface{}) {
	ui.parent.ErrorLinef(pattern, args...)
}

func (ui *nonInteractiveUI) PrintLinef(pattern string, args ...interface{}) {
	ui.parent.PrintLinef(pattern, args...)
}

func (ui *nonInteractiveUI) BeginLinef(pattern string, args ...interface{}) {
	ui.parent.BeginLinef(pattern, args...)
}

func (ui *nonInteractiveUI) EndLinef(pattern string, args ...interface{}) {
	ui.parent.EndLinef(pattern, args...)
}

func (ui *nonInteractiveUI) PrintBlock(block []byte) {
	ui.parent.PrintBlock(block)
}

func (ui *nonInteractiveUI) PrintErrorBlock(block string) {
	ui.parent.PrintErrorBlock(block)
}

func (ui *nonInteractiveUI) PrintTable(table Table) {
	ui.parent.PrintTable(table)
}

func (ui *nonInteractiveUI) AskForText(label string) (string, error) {
	panic("Cannot ask for input in non-interactive UI")
}

func (ui *nonInteractiveUI) AskForChoice(label string, options []string) (int, error) {
	panic("Cannot ask for a choice in non-interactive UI")
}

func (ui *nonInteractiveUI) AskForPassword(label string) (string, error) {
	panic("Cannot ask for password in non-interactive UI")
}

func (ui *nonInteractiveUI) AskForConfirmation() error {
	// Always respond successfully
	return nil
}

func (ui *nonInteractiveUI) IsInteractive() bool {
	return false
}

func (ui *nonInteractiveUI) Flush() {
	ui.parent.Flush()
}
