package terminal

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	term "github.com/cloudfoundry/cli/cf/terminal"
)

const QuietPanic = "I should not print anything"

type FakeUI struct {
	Outputs                    []string
	UncapturedOutput           []string
	WarnOutputs                []string
	Prompts                    []string
	PasswordPrompts            []string
	Inputs                     []string
	FailedWithUsage            bool
	FailedWithUsageCommandName string
	PanickedQuietly            bool
	ShowConfigurationCalled    bool

	sayMutex sync.Mutex
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

func (ui *FakeUI) PrintCapturingNoOutput(message string, args ...interface{}) {
	ui.sayMutex.Lock()
	defer ui.sayMutex.Unlock()

	message = fmt.Sprintf(message, args...)
	ui.UncapturedOutput = append(ui.UncapturedOutput, strings.Split(message, "\n")...)
	return
}

func (ui *FakeUI) Say(message string, args ...interface{}) {
	ui.sayMutex.Lock()
	defer ui.sayMutex.Unlock()

	message = fmt.Sprintf(message, args...)
	ui.Outputs = append(ui.Outputs, strings.Split(message, "\n")...)
	return
}

func (ui *FakeUI) Warn(message string, args ...interface{}) {
	message = fmt.Sprintf(message, args...)
	ui.WarnOutputs = append(ui.WarnOutputs, strings.Split(message, "\n")...)
	ui.Say(message, args...)
	return
}

func (ui *FakeUI) Ask(prompt string, args ...interface{}) (answer string) {
	ui.Prompts = append(ui.Prompts, fmt.Sprintf(prompt, args...))
	if len(ui.Inputs) == 0 {
		panic("No input provided to Fake UI for prompt: " + fmt.Sprintf(prompt, args...))
	}

	answer = ui.Inputs[0]
	ui.Inputs = ui.Inputs[1:]
	return
}

func (ui *FakeUI) ConfirmDelete(modelType, modelName string) bool {
	return ui.Confirm(
		"Really delete the %s %s?%s",
		modelType,
		term.EntityNameColor(modelName),
		term.PromptColor(">"))
}

func (ui *FakeUI) ConfirmDeleteWithAssociations(modelType, modelName string) bool {
	return ui.ConfirmDelete(modelType, modelName)
}

func (ui *FakeUI) Confirm(prompt string, args ...interface{}) bool {
	response := ui.Ask(prompt, args...)
	switch strings.ToLower(response) {
	case "y", "yes":
		return true
	}
	return false
}

func (ui *FakeUI) AskForPassword(prompt string, args ...interface{}) (answer string) {
	ui.PasswordPrompts = append(ui.PasswordPrompts, fmt.Sprintf(prompt, args...))
	if len(ui.Inputs) == 0 {
		println("__________________PANIC__________________")
		println(ui.DumpOutputs())
		println(ui.DumpPrompts())
		println("_________________________________________")
		panic("No input provided to Fake UI for prompt: " + fmt.Sprintf(prompt, args...))
	}

	answer = ui.Inputs[0]
	ui.Inputs = ui.Inputs[1:]
	return
}

func (ui *FakeUI) Ok() {
	ui.Say("OK")
}

func (ui *FakeUI) Failed(message string, args ...interface{}) {
	ui.Say("FAILED")
	ui.Say(message, args...)
	panic(QuietPanic)
	return
}

func (ui *FakeUI) PanicQuietly() {
	ui.PanickedQuietly = true
}

func (ui *FakeUI) DumpWarnOutputs() string {
	return "****************************\n" + strings.Join(ui.WarnOutputs, "\n")
}

func (ui *FakeUI) DumpOutputs() string {
	return "****************************\n" + strings.Join(ui.Outputs, "\n")
}

func (ui *FakeUI) DumpPrompts() string {
	return "****************************\n" + strings.Join(ui.Prompts, "\n")
}

func (ui *FakeUI) ClearOutputs() {
	ui.Outputs = []string{}
}

func (ui *FakeUI) ShowConfiguration(config core_config.Reader) {
	ui.ShowConfigurationCalled = true
}

func (ui FakeUI) LoadingIndication() {
}

func (c FakeUI) Wait(duration time.Duration) {
	time.Sleep(duration)
}

func (ui *FakeUI) Table(headers []string) term.Table {
	return term.NewTable(ui, headers)
}

func (ui *FakeUI) NotifyUpdateIfNeeded(config core_config.Reader) {
	if !config.IsMinCliVersion(cf.Version) {
		ui.Say("Cloud Foundry API version {{.ApiVer}} requires CLI version " + config.MinCliVersion() + "  You are currently on version {{.CliVer}}. To upgrade your CLI, please visit: https://github.com/cloudfoundry/cli#downloads")
	}
}
