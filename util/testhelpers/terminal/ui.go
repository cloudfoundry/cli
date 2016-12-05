package terminal

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"os"

	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	term "code.cloudfoundry.org/cli/cf/terminal"
)

type FakeUI struct {
	outputs                       []string
	uncapturedOutput              []string
	WarnOutputs                   []string
	Prompts                       []string
	PasswordPrompts               []string
	Inputs                        []string
	FailedWithUsage               bool
	FailedWithUsageCommandName    string
	ShowConfigurationCalled       bool
	NotifyUpdateIfNeededCallCount int

	sayMutex sync.Mutex
}

func (ui *FakeUI) Outputs() []string {
	ui.sayMutex.Lock()
	defer ui.sayMutex.Unlock()

	return ui.outputs
}

func (ui *FakeUI) UncapturedOutput() []string {
	ui.sayMutex.Lock()
	defer ui.sayMutex.Unlock()

	return ui.uncapturedOutput
}

func (ui *FakeUI) PrintPaginator(rows []string, err error) {
	if err != nil {
		ui.Failed(err.Error())
		return
	}

	for _, row := range rows {
		ui.Say(row)
	}
}

func (ui *FakeUI) Writer() io.Writer {
	return os.Stdout
}

func (ui *FakeUI) PrintCapturingNoOutput(message string, args ...interface{}) {
	ui.sayMutex.Lock()
	defer ui.sayMutex.Unlock()

	message = fmt.Sprintf(message, args...)
	ui.uncapturedOutput = append(ui.uncapturedOutput, strings.Split(message, "\n")...)
	return
}

func (ui *FakeUI) Say(message string, args ...interface{}) {
	ui.sayMutex.Lock()
	defer ui.sayMutex.Unlock()

	message = fmt.Sprintf(message, args...)
	ui.outputs = append(ui.outputs, strings.Split(message, "\n")...)
	return
}

func (ui *FakeUI) Warn(message string, args ...interface{}) {
	message = fmt.Sprintf(message, args...)
	ui.WarnOutputs = append(ui.WarnOutputs, strings.Split(message, "\n")...)
	ui.Say(message, args...)
	return
}

func (ui *FakeUI) Ask(prompt string) string {
	ui.Prompts = append(ui.Prompts, prompt)

	if len(ui.Inputs) == 0 {
		panic("No input provided to Fake UI for prompt: " + prompt)
	}

	answer := ui.Inputs[0]
	ui.Inputs = ui.Inputs[1:]
	return answer
}

func (ui *FakeUI) ConfirmDelete(modelType, modelName string) bool {
	return ui.Confirm(fmt.Sprintf(
		"Really delete the %s %s?%s",
		modelType,
		term.EntityNameColor(modelName),
		term.PromptColor(">")))
}

func (ui *FakeUI) ConfirmDeleteWithAssociations(modelType, modelName string) bool {
	return ui.ConfirmDelete(modelType, modelName)
}

func (ui *FakeUI) Confirm(prompt string) bool {
	response := ui.Ask(prompt)
	switch strings.ToLower(response) {
	case "y", "yes":
		return true
	}
	return false
}

func (ui *FakeUI) AskForPassword(prompt string) string {
	ui.PasswordPrompts = append(ui.PasswordPrompts, prompt)

	if len(ui.Inputs) == 0 {
		panic("No input provided to Fake UI for prompt: " + prompt)
	}

	answer := ui.Inputs[0]
	ui.Inputs = ui.Inputs[1:]
	return answer
}

func (ui *FakeUI) Ok() {
	ui.Say("OK")
}

func (ui *FakeUI) Failed(message string, args ...interface{}) {
	ui.Say("FAILED")
	ui.Say(message, args...)
}

func (ui *FakeUI) DumpWarnOutputs() string {
	return "****************************\n" + strings.Join(ui.WarnOutputs, "\n")
}

func (ui *FakeUI) DumpOutputs() string {
	return "****************************\n" + strings.Join(ui.Outputs(), "\n")
}

func (ui *FakeUI) DumpPrompts() string {
	return "****************************\n" + strings.Join(ui.Prompts, "\n")
}

func (ui *FakeUI) ClearOutputs() {
	ui.sayMutex.Lock()
	defer ui.sayMutex.Unlock()

	ui.outputs = []string{}
}

func (ui *FakeUI) ShowConfiguration(config coreconfig.Reader) error {
	ui.ShowConfigurationCalled = true
	return nil
}

func (ui *FakeUI) LoadingIndication() {
}

func (ui *FakeUI) Wait(duration time.Duration) {
	time.Sleep(duration)
}

func (ui *FakeUI) Table(headers []string) *term.UITable {
	return &term.UITable{
		UI:    ui,
		Table: term.NewTable(headers),
	}
}

func (ui *FakeUI) NotifyUpdateIfNeeded(config coreconfig.Reader) {
	ui.NotifyUpdateIfNeededCallCount += 1
}
