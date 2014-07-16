package terminal

import (
	"fmt"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"io"
	"os"
	"strings"
	"time"

	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/trace"
	"github.com/codegangsta/cli"
)

type ColoringFunction func(value string, row int, col int) string

func NotLoggedInText() string {
	return fmt.Sprintf(T("Not logged in. Use '{{.CFLoginCommand}}' to log in.", map[string]interface{}{"CFLoginCommand": CommandColor(cf.Name() + " " + "login")}))
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
	FailWithUsage(context *cli.Context)
	PanicQuietly()
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
	return c.confirmDelete(T("Really delete the {{.ModelType}} {{.ModelName}} and everything associated with it?",
		map[string]interface{}{
			"ModelType": modelType,
			"ModelName": EntityNameColor(modelName),
		}))
}

func (c terminalUI) ConfirmDelete(modelType, modelName string) bool {
	return c.confirmDelete(T("Really delete the {{.ModelType}} {{.ModelName}}?",
		map[string]interface{}{
			"ModelType": modelType,
			"ModelName": EntityNameColor(modelName),
		}))
}

func (c terminalUI) confirmDelete(message string) bool {
	result := c.Confirm(message)

	if !result {
		c.Warn(T("Delete cancelled"))
	}

	return result
}

func (c terminalUI) Confirm(message string, args ...interface{}) bool {
	response := c.Ask(message, args...)
	switch strings.ToLower(response) {
	case "y", T("yes"):
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
	c.Say(SuccessColor(T("OK")))
}

const QuietPanic = "This shouldn't print anything"

func (c terminalUI) Failed(message string, args ...interface{}) {
	message = fmt.Sprintf(message, args...)
	c.Say(FailureColor(T("FAILED")))
	c.Say(message)

	trace.Logger.Print(T("FAILED"))
	trace.Logger.Print(message)
	c.PanicQuietly()
}

func (c terminalUI) PanicQuietly() {
	panic(QuietPanic)
}

func (c terminalUI) FailWithUsage(context *cli.Context) {
	c.Say(FailureColor(T("FAILED")))
	c.Say(T("Incorrect Usage.\n"))
	cli.ShowCommandHelp(context, context.Command.Name)
	c.Say("")
	os.Exit(1)
}

func (ui terminalUI) ShowConfiguration(config configuration.Reader) {
	table := NewTable(ui, []string{"", ""})
	if config.HasAPIEndpoint() {
		table.Add(
			T("API endpoint:"),
			T("{{.ApiEndpoint}} (API version: {{.ApiVersionString}})",
				map[string]interface{}{
					"ApiEndpoint":      EntityNameColor(config.ApiEndpoint()),
					"ApiVersionString": EntityNameColor(config.ApiVersion()),
				}),
		)
	}

	if !config.IsLoggedIn() {
		table.Print()
		ui.Say(NotLoggedInText())
		return
	} else {
		table.Add(
			T("User:"),
			EntityNameColor(config.UserEmail()),
		)
	}

	if !config.HasOrganization() && !config.HasSpace() {
		table.Print()
		command := fmt.Sprintf("%s target -o ORG -s SPACE", cf.Name())
		ui.Say(T("No org or space targeted, use '{{.CFTargetCommand}}'",
			map[string]interface{}{
				"CFTargetCommand": CommandColor(command),
			}))
		return
	}

	if config.HasOrganization() {
		table.Add(
			T("Org:"),
			EntityNameColor(config.OrganizationFields().Name),
		)
	} else {
		command := fmt.Sprintf("%s target -o Org", cf.Name())
		table.Add(
			T("Org:"),
			T("No org targeted, use '{{.CFTargetCommand}}'",
				map[string]interface{}{
					"CFTargetCommand": CommandColor(command),
				}),
		)
	}

	if config.HasSpace() {
		table.Add(
			T("Space:"),
			EntityNameColor(config.SpaceFields().Name),
		)
	} else {
		command := fmt.Sprintf("%s target -s SPACE", cf.Name())
		table.Add(
			T("Space:"),
			T("No space targeted, use '{{.CFTargetCommand}}'", map[string]interface{}{"CFTargetCommand": CommandColor(command)}),
		)
	}

	table.Print()
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
