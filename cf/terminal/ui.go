package terminal

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"time"

	. "github.com/cloudfoundry/cli/cf/i18n"

	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/trace"
)

type ColoringFunction func(value string, row int, col int) string

func NotLoggedInText() string {
	return fmt.Sprintf(T("Not logged in. Use '{{.CFLoginCommand}}' to log in.", map[string]interface{}{"CFLoginCommand": CommandColor(cf.Name() + " " + "login")}))
}

type UI interface {
	PrintPaginator(rows []string, err error)
	Say(message string, args ...interface{})
	PrintCapturingNoOutput(message string, args ...interface{})
	Warn(message string, args ...interface{})
	Ask(prompt string, args ...interface{}) (answer string)
	AskForPassword(prompt string, args ...interface{}) (answer string)
	Confirm(message string, args ...interface{}) bool
	ConfirmDelete(modelType, modelName string) bool
	ConfirmDeleteWithAssociations(modelType, modelName string) bool
	Ok()
	Failed(message string, args ...interface{})
	PanicQuietly()
	ShowConfiguration(core_config.Reader)
	LoadingIndication()
	Wait(duration time.Duration)
	Table(headers []string) Table
	NotifyUpdateIfNeeded(core_config.Reader)
}

type terminalUI struct {
	stdin   io.Reader
	printer Printer
}

func NewUI(r io.Reader, printer Printer) UI {
	return &terminalUI{
		stdin:   r,
		printer: printer,
	}
}

func (c *terminalUI) PrintPaginator(rows []string, err error) {
	if err != nil {
		c.Failed(err.Error())
		return
	}

	for _, row := range rows {
		c.Say(row)
	}
}

func (c *terminalUI) PrintCapturingNoOutput(message string, args ...interface{}) {
	if len(args) == 0 {
		fmt.Printf("%s", message)
	} else {
		fmt.Printf(message, args...)
	}
}

func (c *terminalUI) Say(message string, args ...interface{}) {
	if len(args) == 0 {
		c.printer.Printf("%s\n", message)
	} else {
		c.printer.Printf(message+"\n", args...)
	}
}

func (c *terminalUI) Warn(message string, args ...interface{}) {
	message = fmt.Sprintf(message, args...)
	c.Say(WarningColor(message))
	return
}

func (c *terminalUI) ConfirmDeleteWithAssociations(modelType, modelName string) bool {
	return c.confirmDelete(T("Really delete the {{.ModelType}} {{.ModelName}} and everything associated with it?",
		map[string]interface{}{
			"ModelType": modelType,
			"ModelName": EntityNameColor(modelName),
		}))
}

func (c *terminalUI) ConfirmDelete(modelType, modelName string) bool {
	return c.confirmDelete(T("Really delete the {{.ModelType}} {{.ModelName}}?",
		map[string]interface{}{
			"ModelType": modelType,
			"ModelName": EntityNameColor(modelName),
		}))
}

func (c *terminalUI) confirmDelete(message string) bool {
	result := c.Confirm(message)

	if !result {
		c.Warn(T("Delete cancelled"))
	}

	return result
}

func (c *terminalUI) Confirm(message string, args ...interface{}) bool {
	response := c.Ask(message, args...)
	switch strings.ToLower(response) {
	case "y", "yes", T("yes"):
		return true
	}
	return false
}

func (c *terminalUI) Ask(prompt string, args ...interface{}) (answer string) {
	fmt.Println("")
	fmt.Printf(prompt+PromptColor(">")+" ", args...)

	rd := bufio.NewReader(c.stdin)
	line, err := rd.ReadString('\n')
	if err == nil {
		return strings.TrimSpace(line)
	}
	return ""
}

func (c *terminalUI) Ok() {
	c.Say(SuccessColor(T("OK")))
}

const QuietPanic = "This shouldn't print anything"

func (c *terminalUI) Failed(message string, args ...interface{}) {
	message = fmt.Sprintf(message, args...)

	if T == nil {
		c.Say(FailureColor("FAILED"))
		c.Say(message)

		trace.Logger.Print("FAILED")
		trace.Logger.Print(message)
		c.PanicQuietly()
	} else {
		c.Say(FailureColor(T("FAILED")))
		c.Say(message)

		trace.Logger.Print(T("FAILED"))
		trace.Logger.Print(message)
		c.PanicQuietly()
	}
}

func (c *terminalUI) PanicQuietly() {
	panic(QuietPanic)
}

func (ui *terminalUI) ShowConfiguration(config core_config.Reader) {
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

func (c *terminalUI) LoadingIndication() {
	c.printer.Print(".")
}

func (c *terminalUI) Wait(duration time.Duration) {
	time.Sleep(duration)
}

func (ui *terminalUI) Table(headers []string) Table {
	return NewTable(ui, headers)
}

func (ui *terminalUI) NotifyUpdateIfNeeded(config core_config.Reader) {
	if !config.IsMinCliVersion(cf.Version) {
		ui.Say("")
		ui.Say(T("Cloud Foundry API version {{.ApiVer}} requires CLI version {{.CliMin}}.  You are currently on version {{.CliVer}}. To upgrade your CLI, please visit: https://github.com/cloudfoundry/cli#downloads",
			map[string]interface{}{
				"ApiVer": config.ApiVersion(),
				"CliMin": config.MinCliVersion(),
				"CliVer": cf.Version,
			}))
	}
}
