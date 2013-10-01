package terminal

import (
	"cf"
	"cf/configuration"
	"fmt"
	"github.com/codegangsta/cli"
	"os"
	"strings"
	"time"
)

type ColoringFunction func(value string, row int, col int) string

func NotLoggedInText() string {
	return fmt.Sprintf("Not logged in. Use '%s' to log in.", CommandColor(cf.Name+" login"))
}

type UI interface {
	Say(message string, args ...interface{})
	Warn(message string, args ...interface{})
	Ask(prompt string, args ...interface{}) (answer string)
	AskForPassword(prompt string, args ...interface{}) (answer string)
	Ok()
	Failed(message string, args ...interface{})
	FailWithUsage(ctxt *cli.Context, cmdName string)
	ConfigFailure(err error)
	ShowConfiguration(configuration.Configuration)
	LoadingIndication()
	Wait(duration time.Duration)
	DisplayTable(table [][]string, coloringFunc ColoringFunction)
}

type TerminalUI struct {
}

func (c TerminalUI) Say(message string, args ...interface{}) {
	fmt.Printf(message+"\n", args...)
	return
}

func (c TerminalUI) Warn(message string, args ...interface{}) {
	message = fmt.Sprintf(message, args...)
	c.Say(WarningColor(message))
	return
}

func (c TerminalUI) Ask(prompt string, args ...interface{}) (answer string) {
	fmt.Println("")
	fmt.Printf(prompt+" ", args...)
	fmt.Scanln(&answer)
	return
}

func (c TerminalUI) Ok() {
	c.Say(SuccessColor("OK"))
}

func (c TerminalUI) Failed(message string, args ...interface{}) {
	message = fmt.Sprintf(message, args...)
	c.Say(FailureColor("FAILED"))
	c.Say(message)
	os.Exit(1)
}

func (c TerminalUI) FailWithUsage(ctxt *cli.Context, cmdName string) {
	c.Say(FailureColor("FAILED"))
	c.Say("Incorrect Usage.\n")
	cli.ShowCommandHelp(ctxt, cmdName)
	c.Say("")
	os.Exit(1)
}

func (c TerminalUI) ConfigFailure(err error) {
	c.Failed("Error loading config. Please reset the api '%s' and log in '%s'.\n%s",
		CommandColor(fmt.Sprintf("%s api", cf.Name)),
		CommandColor(fmt.Sprintf("%s login", cf.Name)),
		err.Error())
}

func (ui TerminalUI) ShowConfiguration(config configuration.Configuration) {
	ui.Say("API endpoint: %s (API version: %s)",
		EntityNameColor(config.Target),
		EntityNameColor(config.ApiVersion))

	if !config.IsLoggedIn() {
		ui.Say("Logged out. Use '%s' to login.", CommandColor(cf.Name+" login USERNAME"))
	} else {
		ui.Say("user:            %s", EntityNameColor(config.UserEmail()))
	}

	if config.HasOrganization() {
		ui.Say("org:             %s", EntityNameColor(config.Organization.Name))
	}

	if config.HasSpace() {
		ui.Say("space:           %s", EntityNameColor(config.Space.Name))
	}
}

func (c TerminalUI) LoadingIndication() {
	fmt.Print(".")
}

func (c TerminalUI) Wait(duration time.Duration) {
	time.Sleep(duration)
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
		return HeaderColor(value)
	case col == 0 && row > 0:
		return TableContentHeaderColor(value)
	}

	return TableContentColor(value)
}
