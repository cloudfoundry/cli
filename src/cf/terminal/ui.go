package terminal

import (
	"cf/configuration"
	"fmt"
	"github.com/codegangsta/cli"
	"os"
	"strings"
	"time"
)

type ColoringFunction func(value string, row int, col int) string

type UI interface {
	Say(message string, args ...interface{})
	Ask(prompt string, args ...interface{}) (answer string)
	AskForPassword(prompt string, args ...interface{}) (answer string)
	Ok()
	Failed(message string, err error)
	FailWithUsage(ctxt *cli.Context, cmdName string)
	ShowConfiguration(*configuration.Configuration)
	ShowConfigurationNoSpacesAvailable(config *configuration.Configuration)
	LoadingIndication()
	Wait(seconds time.Duration)
	DisplayTable(table [][]string, coloringFunc ColoringFunction)
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

	if message != "" && err == nil {
		c.Say(message)
	}

	if err != nil {
		c.Say(err.Error())
	}

	return
}

func (c TerminalUI) FailWithUsage(ctxt *cli.Context, cmdName string) {
	c.Failed("Incorrect Usage.\n", nil)
	cli.ShowCommandHelp(ctxt, cmdName)
	c.Say("")
	os.Exit(1)
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

func (c TerminalUI) LoadingIndication() {
	fmt.Print(".")
}

func (c TerminalUI) Wait(seconds time.Duration) {
	time.Sleep(1000 * 1000 * 1000 * seconds)
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

func (ui TerminalUI) DisplayTable(table [][]string, coloringFunc ColoringFunction) {
	if coloringFunc == nil {
		coloringFunc = DefaultColoringFunc
	}

	columnCount := len(table[0])
	maxSizes := make([]int, columnCount)

	for _, line := range table {
		for index, value := range line {
			if maxSizes[index] < len(value) {
				maxSizes[index] = len(value)
			}
		}
	}

	for row, line := range table {
		for col, value := range line {
			padding := strings.Repeat(" ", maxSizes[col]-len(value))
			value = coloringFunc(value, row, col)
			fmt.Printf("%s%s   ", value, padding)
		}
		fmt.Print("\n")
	}
}

func DefaultColoringFunc(value string, row int, col int) string {
	switch {
	case row == 0:
		return White(value)
	case col == 0 && row > 0:
		return Cyan(value)
	}

	return Grey(value)
}
