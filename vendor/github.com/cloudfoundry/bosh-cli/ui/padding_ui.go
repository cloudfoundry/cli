package ui

import (
	. "github.com/cloudfoundry/bosh-cli/ui/table"
)

type paddingUIMode int

const (
	paddingUIModeNone paddingUIMode = iota
	paddingUIModeRaw
	paddingUIModeAuto
	paddingUIModeAskText
)

type paddingUI struct {
	parent   UI
	prevMode paddingUIMode
}

func NewPaddingUI(parent UI) UI {
	return &paddingUI{parent: parent}
}

func (ui *paddingUI) ErrorLinef(pattern string, args ...interface{}) {
	ui.padBefore(paddingUIModeAuto)
	ui.parent.ErrorLinef(pattern, args...)
}

func (ui *paddingUI) PrintLinef(pattern string, args ...interface{}) {
	ui.padBefore(paddingUIModeAuto)
	ui.parent.PrintLinef(pattern, args...)
}

func (ui *paddingUI) BeginLinef(pattern string, args ...interface{}) {
	ui.padBefore(paddingUIModeRaw)
	ui.parent.BeginLinef(pattern, args...)
}

func (ui *paddingUI) EndLinef(pattern string, args ...interface{}) {
	ui.padBefore(paddingUIModeRaw)
	ui.parent.EndLinef(pattern, args...)
}

func (ui *paddingUI) PrintBlock(block []byte) {
	ui.padBefore(paddingUIModeRaw)
	ui.parent.PrintBlock(block)
}

func (ui *paddingUI) PrintErrorBlock(block string) {
	ui.padBefore(paddingUIModeRaw)
	ui.parent.PrintErrorBlock(block)
}

func (ui *paddingUI) PrintTable(table Table) {
	ui.padBefore(paddingUIModeAuto)
	ui.parent.PrintTable(table)
}

func (ui *paddingUI) AskForText(label string) (string, error) {
	ui.padBefore(paddingUIModeAskText)
	return ui.parent.AskForText(label)
}

func (ui *paddingUI) AskForChoice(label string, options []string) (int, error) {
	ui.padBefore(paddingUIModeAuto)
	return ui.parent.AskForChoice(label, options)
}

func (ui *paddingUI) AskForPassword(label string) (string, error) {
	ui.padBefore(paddingUIModeAskText)
	return ui.parent.AskForPassword(label)
}

func (ui *paddingUI) AskForConfirmation() error {
	ui.padBefore(paddingUIModeAuto)
	return ui.parent.AskForConfirmation()
}

func (ui *paddingUI) IsInteractive() bool {
	return ui.parent.IsInteractive()
}

func (ui *paddingUI) Flush() {
	ui.parent.Flush()
}

func (ui *paddingUI) padBefore(currMode paddingUIMode) {
	switch {
	case ui.prevMode == paddingUIModeNone:
		// do nothing on the first time UI is called
	case ui.prevMode == paddingUIModeAskText && currMode == paddingUIModeAskText:
		// do nothing
	case ui.prevMode == paddingUIModeRaw && currMode == paddingUIModeRaw:
		// do nothing
	default:
		ui.parent.PrintLinef("")
	}
	ui.prevMode = currMode
}
