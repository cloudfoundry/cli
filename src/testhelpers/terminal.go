package testhelpers

import (
	"bytes"
	"io"
	"os"
	"fmt"
	"strings"
	"cf/configuration"
	term "cf/terminal"
	"time"
	"github.com/codegangsta/cli"
)

func CaptureOutput(f func ()) string {
	old := os.Stdout // keep backup of the real stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	outC := make(chan string)
	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	// back to normal state
	w.Close()
	os.Stdout = old // restoring the real stdout
	return <-outC
}

type FakeUI struct {
	Outputs []string
	Prompts []string
	PasswordPrompts []string
	Inputs  []string
	FailedWithUsage bool
}

func (ui *FakeUI) Say(message string, args ...interface{}) {
	ui.Outputs = append(ui.Outputs, fmt.Sprintf(message, args...))
	return
}

func (ui *FakeUI) Ask(prompt string, args ...interface{}) (answer string) {
	ui.Prompts = append(ui.Prompts, fmt.Sprintf(prompt, args...))
	answer = ui.Inputs[0]
	ui.Inputs = ui.Inputs[1:]
	return
}

func (ui *FakeUI) AskForPassword(prompt string, args ...interface{}) (answer string) {
	ui.PasswordPrompts = append(ui.PasswordPrompts, fmt.Sprintf(prompt, args...))
	answer = ui.Inputs[0]
	ui.Inputs = ui.Inputs[1:]
	return
}

func (ui *FakeUI) Ok() {
	ui.Say("OK")
}

func (ui *FakeUI) Failed(message string) {
	ui.Say("FAILED")
	ui.Say(message)
	return
}

func (ui *FakeUI) FailWithUsage(ctxt *cli.Context, cmdName string) {
	ui.FailedWithUsage = true
	ui.Failed("Incorrect Usage.")
}

func (ui *FakeUI) DumpOutputs() string {
	return "****************************\n" + strings.Join(ui.Outputs, "\n")
}

func (ui *FakeUI) ClearOutputs() {
	ui.Outputs = []string{}
}

func (ui *FakeUI) ShowConfiguration(config *configuration.Configuration) {
	ui.Say("API endpoint: %s (API version: %s)",
		config.Target,
		config.ApiVersion)

	if !config.IsLoggedIn() {
		ui.Say("Logged out. Use '%s' to login.", "cf login USERNAME")
		return
	} else {
		ui.Say("user:            %s", config.UserEmail())
	}

	if config.HasOrganization() {
		ui.Say("org:             %s", config.Organization.Name)
	}

	if config.HasSpace() {
		ui.Say("app space:       %s", config.Space.Name)
	}
}

func (ui FakeUI) LoadingIndication() {
}

func (c FakeUI) Wait(duration time.Duration) {
	time.Sleep(duration)
}

func (ui *FakeUI) showBaseConfig(config *configuration.Configuration) {

}

func (ui *FakeUI) DisplayTable(table [][]string, coloringFunc term.ColoringFunction) {
	if coloringFunc == nil {
		coloringFunc = term.DefaultColoringFunc
	}

	for row, line := range table {
		output := ""
		for col, value := range line {
			output = output + coloringFunc(value, row, col) + "  "
		}
		ui.Say(output)
	}
}
