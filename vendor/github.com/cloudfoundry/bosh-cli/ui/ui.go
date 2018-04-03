package ui

import (
	"errors"
	"fmt"
	"io"
	"os"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	"github.com/mattn/go-isatty"
	"github.com/vito/go-interact/interact"

	. "github.com/cloudfoundry/bosh-cli/ui/table"
)

type WriterUI struct {
	outWriter io.Writer
	errWriter io.Writer
	logger    boshlog.Logger
	logTag    string
}

func NewConsoleUI(logger boshlog.Logger) *WriterUI {
	return NewWriterUI(os.Stdout, os.Stderr, logger)
}

func NewWriterUI(outWriter, errWriter io.Writer, logger boshlog.Logger) *WriterUI {
	return &WriterUI{
		outWriter: outWriter,
		errWriter: errWriter,

		logTag: "ui",
		logger: logger,
	}
}

func (ui *WriterUI) IsTTY() bool {
	file, ok := ui.outWriter.(*os.File)

	return ok && isatty.IsTerminal(file.Fd())
}

// ErrorLinef starts and ends a text error line
func (ui *WriterUI) ErrorLinef(pattern string, args ...interface{}) {
	message := fmt.Sprintf(pattern, args...)
	_, err := fmt.Fprintln(ui.errWriter, message)
	if err != nil {
		ui.logger.Error(ui.logTag, "UI.ErrorLinef failed (message='%s'): %s", message, err)
	}
}

// Printlnf starts and ends a text line
func (ui *WriterUI) PrintLinef(pattern string, args ...interface{}) {
	message := fmt.Sprintf(pattern, args...)
	_, err := fmt.Fprintln(ui.outWriter, message)
	if err != nil {
		ui.logger.Error(ui.logTag, "UI.PrintLinef failed (message='%s'): %s", message, err)
	}
}

// PrintBeginf starts a text line
func (ui *WriterUI) BeginLinef(pattern string, args ...interface{}) {
	message := fmt.Sprintf(pattern, args...)
	_, err := fmt.Fprint(ui.outWriter, message)
	if err != nil {
		ui.logger.Error(ui.logTag, "UI.BeginLinef failed (message='%s'): %s", message, err)
	}
}

// PrintEndf ends a text line
func (ui *WriterUI) EndLinef(pattern string, args ...interface{}) {
	message := fmt.Sprintf(pattern, args...)
	_, err := fmt.Fprintln(ui.outWriter, message)
	if err != nil {
		ui.logger.Error(ui.logTag, "UI.EndLinef failed (message='%s'): %s", message, err)
	}
}

func (ui *WriterUI) PrintBlock(block []byte) {
	_, err := ui.outWriter.Write(block)
	if err != nil {
		ui.logger.Error(ui.logTag, "UI.PrintBlock failed (message='%s'): %s", block, err)
	}
}

func (ui *WriterUI) PrintErrorBlock(block string) {
	_, err := fmt.Fprint(ui.outWriter, block)
	if err != nil {
		ui.logger.Error(ui.logTag, "UI.PrintErrorBlock failed (message='%s'): %s", block, err)
	}
}

func (ui *WriterUI) PrintTable(table Table) {
	err := table.Print(ui.outWriter)
	if err != nil {
		ui.logger.Error(ui.logTag, "UI.PrintTable failed: %s", err)
	}
}

func (ui *WriterUI) AskForText(label string) (string, error) {
	var text string

	err := interact.NewInteraction(label).Resolve(&text)
	if err != nil {
		return "", bosherr.WrapError(err, "Asking for text")
	}

	return text, nil
}

func (ui *WriterUI) AskForChoice(label string, options []string) (int, error) {
	var choices []interact.Choice

	for i, opt := range options {
		choices = append(choices, interact.Choice{Display: opt, Value: i})
	}

	var chosen int

	err := interact.NewInteraction(label, choices...).Resolve(&chosen)
	if err != nil {
		return 0, bosherr.WrapError(err, "Asking for choice")
	}

	return chosen, nil
}

func (ui *WriterUI) AskForPassword(label string) (string, error) {
	var password interact.Password

	err := interact.NewInteraction(label).Resolve(&password)
	if err != nil {
		return "", bosherr.WrapError(err, "Asking for password")
	}

	return string(password), nil
}

func (ui *WriterUI) AskForConfirmation() error {
	falseByDefault := false

	err := interact.NewInteraction("Continue?").Resolve(&falseByDefault)
	if err != nil {
		return bosherr.WrapError(err, "Asking for confirmation")
	}

	if falseByDefault == false {
		return errors.New("Stopped")
	}

	return nil
}

func (ui *WriterUI) IsInteractive() bool {
	return true
}

func (ui *WriterUI) Flush() {}
