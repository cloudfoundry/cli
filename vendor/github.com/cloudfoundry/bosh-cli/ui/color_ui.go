package ui

import (
	"github.com/fatih/color"

	. "github.com/cloudfoundry/bosh-cli/ui/table"
)

type ColorUI struct {
	parent   UI
	okFunc   func(string, ...interface{}) string
	errFunc  func(string, ...interface{}) string
	boldFunc func(string, ...interface{}) string
}

func NewColorUI(parent UI) UI {
	return &ColorUI{
		parent:   parent,
		okFunc:   color.New(color.FgGreen).SprintfFunc(),
		errFunc:  color.New(color.FgRed).SprintfFunc(),
		boldFunc: color.New(color.Bold).SprintfFunc(),
	}
}

func (ui *ColorUI) ErrorLinef(pattern string, args ...interface{}) {
	ui.parent.ErrorLinef("%s", ui.errFunc(pattern, args...))
}

func (ui *ColorUI) PrintLinef(pattern string, args ...interface{}) {
	ui.parent.PrintLinef(pattern, args...)
}

func (ui *ColorUI) BeginLinef(pattern string, args ...interface{}) {
	ui.parent.BeginLinef(pattern, args...)
}

func (ui *ColorUI) EndLinef(pattern string, args ...interface{}) {
	ui.parent.EndLinef(pattern, args...)
}

func (ui *ColorUI) PrintBlock(block []byte) {
	ui.parent.PrintBlock(block)
}

func (ui *ColorUI) PrintErrorBlock(block string) {
	ui.parent.PrintErrorBlock(ui.errFunc("%s", block))
}

func (ui *ColorUI) PrintTable(table Table) {
	table.HeaderFormatFunc = ui.boldFunc

	for k, s := range table.Sections {
		for i, r := range s.Rows {
			for j, v := range r {
				table.Sections[k].Rows[i][j] = ui.colorValueFmt(v)
			}
		}
	}

	for i, r := range table.Rows {
		for j, v := range r {
			table.Rows[i][j] = ui.colorValueFmt(v)
		}
	}

	ui.parent.PrintTable(table)
}

func (ui *ColorUI) AskForText(label string) (string, error) {
	return ui.parent.AskForText(label)
}

func (ui *ColorUI) AskForChoice(label string, options []string) (int, error) {
	return ui.parent.AskForChoice(label, options)
}

func (ui *ColorUI) AskForPassword(label string) (string, error) {
	return ui.parent.AskForPassword(label)
}

func (ui *ColorUI) AskForConfirmation() error {
	return ui.parent.AskForConfirmation()
}

func (ui *ColorUI) IsInteractive() bool {
	return ui.parent.IsInteractive()
}

func (ui *ColorUI) Flush() {
	ui.parent.Flush()
}

func (ui *ColorUI) colorValueFmt(val Value) Value {
	if valFmt, ok := val.(ValueFmt); ok {
		if valFmt.Error {
			valFmt.Func = ui.errFunc
		} else {
			valFmt.Func = ui.okFunc
		}
		return valFmt
	}
	return val
}
