package commands

import (
	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/api/authentication"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type Authenticate struct {
	ui            terminal.UI
	config        configuration.ReadWriter
	authenticator authentication.AuthenticationRepository
}

func NewAuthenticate(ui terminal.UI, config configuration.ReadWriter, authenticator authentication.AuthenticationRepository) (cmd Authenticate) {
	cmd.ui = ui
	cmd.config = config
	cmd.authenticator = authenticator
	return
}

func (cmd Authenticate) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "auth",
		Description: T("Authenticate user non-interactively"),
		Usage: T("CF_NAME auth USERNAME PASSWORD\n\n") +
			terminal.WarningColor(T("WARNING:\n   Providing your password as a command line option is highly discouraged\n   Your password may be visible to others and may be recorded in your shell history\n\n")) + T("EXAMPLE:\n") + T("   CF_NAME auth name@example.com \"my password\" (use quotes for passwords with a space)\n") + T("   CF_NAME auth name@example.com \"\\\"password\\\"\" (escape quotes if used in password)"),
	}
}

func (cmd Authenticate) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 2 {
		cmd.ui.FailWithUsage(c)
	}

	reqs = append(reqs, requirementsFactory.NewApiEndpointRequirement())
	return
}

func (cmd Authenticate) Run(c *cli.Context) {
	cmd.config.ClearSession()
	cmd.authenticator.GetLoginPromptsAndSaveUAAServerURL()

	cmd.ui.Say(T("API endpoint: {{.ApiEndpoint}}",
		map[string]interface{}{"ApiEndpoint": terminal.EntityNameColor(cmd.config.ApiEndpoint())}))
	cmd.ui.Say(T("Authenticating..."))

	apiErr := cmd.authenticator.Authenticate(map[string]string{"username": c.Args()[0], "password": c.Args()[1]})
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say(T("Use '{{.Name}}' to view or set your target org and space",
		map[string]interface{}{"Name": terminal.CommandColor(cf.Name() + " target")}))
	return
}
