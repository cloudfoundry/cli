package terminal

import (
	"fmt"
	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/trace"
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
	ConfirmDelete(modelType, modelName string) bool
	ConfirmDeleteWithAssociations(modelType, modelName string) bool
	Ok()
	Failed(message string, args ...interface{})
	FailWithUsage(ctxt *cli.Context, cmdName string)
	ConfigFailure(err error)
	ShowConfiguration(configuration.Reader)
	LoadingIndication()
	Wait(duration time.Duration)
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
	if len(args) == 0 {
		fmt.Printf("%s\n", message)
	} else {
		fmt.Printf(message+"\n", args...)
	}

	return
}

func (c terminalUI) Warn(message string, args ...interface{}) {
	message = fmt.Sprintf(message, args...)
	c.Say(WarningColor(message))
	return
}

func (c terminalUI) ConfirmDeleteWithAssociations(modelType, modelName string) bool {
	return c.confirmDelete("Really delete the %s %s and everything associated with it?%s", modelType, modelName)
}

func (c terminalUI) ConfirmDelete(modelType, modelName string) bool {
	return c.confirmDelete("Really delete the %s %s?%s", modelType, modelName)
}

func (c terminalUI) confirmDelete(message, modelType, modelName string) bool {
	result := c.Confirm(
		message,
		modelType,
		EntityNameColor(modelName),
		PromptColor(">"),
	)

	if !result {
		c.Warn("Delete cancelled")
	}

	return result
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
	fmt.Printf(prompt+PromptColor(">")+" ", args...)
	fmt.Fscanln(c.stdin, &answer)
	return
}

func (c terminalUI) Ok() {
	c.Say(SuccessColor("OK"))
}

const FailedWasCalled = "FailedWasCalled"

func (c terminalUI) Failed(message string, args ...interface{}) {
	message = fmt.Sprintf(message, args...)
	c.Say(FailureColor("FAILED"))
	c.Say(message)

	trace.Logger.Print("FAILED")
	trace.Logger.Print(message)
	panic(FailedWasCalled)
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

func (ui terminalUI) ShowConfiguration(config configuration.Reader) {
	if config.HasAPIEndpoint() {
		ui.Say("API endpoint: %s (API version: %s)",
			EntityNameColor(config.ApiEndpoint()),
			EntityNameColor(config.ApiVersion()))
	}

	if !config.IsLoggedIn() {
		ui.Say(NotLoggedInText())
		return
	} else {
		ui.Say("User:         %s", EntityNameColor(config.UserEmail()))
	}

	if !config.HasOrganization() && !config.HasSpace() {
		command := fmt.Sprintf("%s target -o ORG -s SPACE", cf.Name())
		ui.Say("No org or space targeted, use '%s'", CommandColor(command))
		return
	}

	if config.HasOrganization() {
		ui.Say("Org:          %s", EntityNameColor(config.OrganizationFields().Name))
	} else {
		command := fmt.Sprintf("%s target -o Org", cf.Name())
		ui.Say("Org:          No org targeted, use '%s'", CommandColor(command))
	}

	if config.HasSpace() {
		ui.Say("Space:        %s", EntityNameColor(config.SpaceFields().Name))
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
