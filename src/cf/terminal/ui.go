package terminal

import (
	"cf/configuration"
	"fmt"
)

type UI interface {
	Say(message string, args ...interface{})
	Ask(prompt string, args ...interface{}) (answer string)
	AskForPassword(prompt string, args ...interface{}) (answer string)
	Ok()
	Failed(message string, err error)
	ShowConfiguration(*configuration.Configuration)
	ShowConfigurationNoSpacesAvailable(config *configuration.Configuration)
}

type TerminalUI struct {
}

func (c TerminalUI) Say(message string, args ...interface{}) {
	fmt.Printf(message+"\n", args...)
	return
}

func (c TerminalUI) Ask(prompt string, args ...interface{}) (answer string) {
	fmt.Println("")
	fmt.Printf(prompt+" ", args...)
	fmt.Scanln(&answer)
	return
}

func (c TerminalUI) Ok() {
	c.Say(Green("OK"))
}

func (c TerminalUI) Failed(message string, err error) {
	c.Say(Red("FAILED"))

	if message != "" {
		c.Say(message)
	}

	if err != nil {
		c.Say(err.Error())
	}
	return
}

func (ui TerminalUI) ShowConfiguration(config *configuration.Configuration) {
	ui.showBaseConfig(config)

	if config.HasSpace() {
		ui.Say("app space:       %s", Yellow(config.Space.Name))
	} else {
		ui.Say("No space targeted. Use 'cf target -s' to target a space.")
	}
}

func (ui TerminalUI) ShowConfigurationNoSpacesAvailable(config *configuration.Configuration) {
	ui.showBaseConfig(config)

	ui.Say("No spaces found. Use 'cf create-space' as an Org Manager.")
}

func (ui TerminalUI) showBaseConfig(config *configuration.Configuration) {
	ui.Say("API endpoint: %s (API version: %s)",
		Yellow(config.Target),
		Yellow(config.ApiVersion))

	if !config.IsLoggedIn() {
		ui.Say("Logged out. Use '%s' to login.", Yellow("cf login USERNAME"))
		return
	}

	ui.Say("user:            %s", Yellow(config.UserEmail()))

	if config.HasOrganization() {
		ui.Say("org:             %s", Yellow(config.Organization.Name))
	} else {
		ui.Say("No org targeted. Use 'cf target -o' to target an org.")
	}
}
