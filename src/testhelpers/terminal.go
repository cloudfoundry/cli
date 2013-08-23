package testhelpers

import (
	"bytes"
	"io"
	"os"
	"fmt"
	"strings"
	"cf/configuration"
)

func CaptureOutput(f func()) string {
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
	Inputs []string
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

func (ui *FakeUI) Ok() {
	ui.Say("OK")
}

func (ui *FakeUI) Failed(message string, err error) {
	ui.Say("FAILED")

	if message != "" {
		ui.Say(message)
	}

	if err != nil {
		ui.Say(err.Error())
	}
	return
}

func (ui *FakeUI)DumpOutputs()string{
	return "****************************\n" + strings.Join(ui.Outputs, "\n")
}

func (ui *FakeUI) ShowConfiguration(config *configuration.Configuration) {
	ui.showBaseConfig(config)

	if config.HasSpace() {
		ui.Say("  app space:       %s", config.Space.Name)
	} else {
		ui.Say("  No space targeted. Use 'cf target -s' to target a space.")
	}
}

func (ui *FakeUI) ShowConfigurationNoSpacesAvailable(config *configuration.Configuration) {
	ui.showBaseConfig(config)

	ui.Say("  No spaces found. Use 'cf create-space' as an Org Manager.")
}

func (ui *FakeUI) showBaseConfig(config *configuration.Configuration) {
	ui.Say("  API endpoint: %s (API version: %s)",
		config.Target,
		config.ApiVersion)

	if !config.IsLoggedIn() {
		ui.Say("  Logged out. Use '%s' to login.", "cf login USERNAME")
		return
	}

	ui.Say("  user:            %s", config.UserEmail())

	if config.HasOrganization() {
		ui.Say("  org:             %s", config.Organization.Name)
	} else {
		ui.Say("  No org targeted. Use 'cf target -o' to target an org.")
	}
}
