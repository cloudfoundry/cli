package terminal

import (
	"cf"
	"cf/configuration"
	"cf/trace"
	"fmt"
	"github.com/codegangsta/cli"
	"io"
	"os"
	"strings"
	"time"
)

type ColoringFunction func(value string, row int, col int) string

func NotLoggedInText() string {
	return fmt.Sprintf("Not logged in. Use '%s' to log in.", CommandColor(cf.Name()+" login"))
}

type UI interface {
	PrintPaginator(rows []string, err error)
	Say(message string, args ...interface{})
	Warn(message string, args ...interface{})
	Ask(prompt string, args ...interface{}) (answer string)
	AskForPassword(prompt string, args ...interface{}) (answer string)
	Confirm(message string, args ...interface{}) bool
	Ok()
	Failed(message string, args ...interface{})
	FailWithUsage(ctxt *cli.Context, cmdName string)
	ConfigFailure(err error)
	ShowConfiguration(*configuration.Configuration)
	LoadingIndication()
	Wait(duration time.Duration)
	DisplayTable(table [][]string)
	Table(headers []string) Table
}

type terminalUI struct {
	stdin io.Reader
}

func NewUI(r io.Reader) UI {
	return terminalUI{stdin: r}
}

func (c terminalUI) PrintPaginator(rows []string, err error) {
	if err != nil {
		c.Failed(err.Error())
		return
	}

	for _, row := range rows {
		c.Say(row)
	}
}

func (c terminalUI) Say(message string, args ...interface{}) {
	fmt.Printf(message+"\n", args...)
	return
}

func (c terminalUI) Warn(message string, args ...interface{}) {
	message = fmt.Sprintf(message, args...)
	c.Say(WarningColor(message))
	return
}

func (c terminalUI) Confirm(message string, args ...interface{}) bool {
	response := c.Ask(message, args...)
	switch strings.ToLower(response) {
	case "y", "yes":
		return true
	}
	return false
}

func (c terminalUI) Ask(prompt string, args ...interface{}) (answer string) {
	fmt.Println("")
	fmt.Printf(prompt+" ", args...)
	fmt.Fscanln(c.stdin, &answer)
	return
}

func (c terminalUI) Ok() {
	c.Say(SuccessColor("OK"))
}

func (c terminalUI) Failed(message string, args ...interface{}) {
	message = fmt.Sprintf(message, args...)
	c.Say(FailureColor("FAILED"))
	c.Say(message)

	trace.Logger.Print("FAILED")
	trace.Logger.Print(message)
	os.Exit(1)
}

func (c terminalUI) FailWithUsage(ctxt *cli.Context, cmdName string) {
	c.Say(FailureColor("FAILED"))
	c.Say("Incorrect Usage.\n")
	cli.ShowCommandHelp(ctxt, cmdName)
	c.Say("")
	os.Exit(1)
}

func (c terminalUI) ConfigFailure(err error) {
	c.Failed("Please use '%s api' to set an API endpoint and then '%s login' to login.", cf.Name(), cf.Name())
}

func (ui terminalUI) ShowConfiguration(config *configuration.Configuration) {
	ui.Say("API endpoint: %s (API version: %s)",
		EntityNameColor(config.Target),
		EntityNameColor(config.ApiVersion))

	if !config.IsLoggedIn() {
		ui.Say(NotLoggedInText())
	} else {
		ui.Say("User:         %s", EntityNameColor(config.UserEmail()))
	}

	if !config.HasOrganization() && !config.HasSpace() {
		command := fmt.Sprintf("%s target -o ORG -s SPACE", cf.Name())
		ui.Say("No org or space targeted, use '%s'", CommandColor(command))
		return
	}

	if config.HasOrganization() {
		ui.Say("Org:          %s", EntityNameColor(config.OrganizationFields.Name))
	} else {
		command := fmt.Sprintf("%s target -o Org", cf.Name())
		ui.Say("Org:          No org targeted, use '%s'", CommandColor(command))
	}

	if config.HasSpace() {
		ui.Say("Space:        %s", EntityNameColor(config.SpaceFields.Name))
	} else {
		command := fmt.Sprintf("%s target -s SPACE", cf.Name())
		ui.Say("Space:        No space targeted, use '%s'", CommandColor(command))
	}
}

func (c terminalUI) LoadingIndication() {
	fmt.Print(".")
}

func (c terminalUI) Wait(duration time.Duration) {
	time.Sleep(duration)
}

func (ui terminalUI) Table(headers []string) Table {
	return NewTable(ui, headers)
}

func (ui terminalUI) DisplayTable(table [][]string) {

	columnCount := len(table[0])
	maxSizes := make([]int, columnCount)

	for _, line := range table {
		for index, value := range line {
			cellLength := len(decolorize(value))
			if maxSizes[index] < cellLength {
				maxSizes[index] = cellLength
			}
		}
	}

	for row, line := range table {
		for col, value := range line {
			padding := strings.Repeat(" ", maxSizes[col]-len(decolorize(value)))
			value = tableColoringFunc(value, row, col)
			fmt.Printf("%s%s   ", value, padding)
		}
		fmt.Print("\n")
	}
}

func tableColoringFunc(value string, row int, col int) string {
	switch {
	case row == 0:
		return HeaderColor(value)
	case col == 0 && row > 0:
		return TableContentHeaderColor(value)
	default:
		return value
	}
}
