package commands

import (
	"fmt"

	"code.cloudfoundry.org/cli/cf"
	"code.cloudfoundry.org/cli/cf/api/authentication"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/flags"
	. "code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
)

type Authenticate struct {
	ui            terminal.UI
	config        coreconfig.ReadWriter
	authenticator authentication.Repository
}

func init() {
	commandregistry.Register(&Authenticate{})
}

func (cmd *Authenticate) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "auth",
		Description: T("Authenticate user non-interactively"),
		Usage: []string{
			T("CF_NAME auth USERNAME PASSWORD\n\n"),
			terminal.WarningColor(T("WARNING:\n   Providing your password as a command line option is highly discouraged\n   Your password may be visible to others and may be recorded in your shell history")),
		},
		Examples: []string{
			T("CF_NAME auth name@example.com \"my password\" (use quotes for passwords with a space)"),
			T("CF_NAME auth name@example.com \"\\\"password\\\"\" (escape quotes if used in password)"),
		},
	}
}

func (cmd *Authenticate) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 2 {
		cmd.ui.Failed(T("Incorrect Usage. Requires 'username password' as arguments\n\n") + commandregistry.Commands.CommandUsage("auth"))
		return nil, fmt.Errorf("Incorrect usage: %d arguments of %d required", len(fc.Args()), 2)
	}

	reqs := []requirements.Requirement{
		requirementsFactory.NewAPIEndpointRequirement(),
	}

	return reqs, nil
}

func (cmd *Authenticate) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.authenticator = deps.RepoLocator.GetAuthenticationRepository()
	return cmd
}

func (cmd *Authenticate) Execute(c flags.FlagContext) error {
	cmd.config.ClearSession()
	cmd.authenticator.GetLoginPromptsAndSaveUAAServerURL()

	cmd.ui.Say(T("API endpoint: {{.APIEndpoint}}",
		map[string]interface{}{"APIEndpoint": terminal.EntityNameColor(cmd.config.APIEndpoint())}))
	cmd.ui.Say(T("Authenticating..."))

	err := cmd.authenticator.Authenticate(map[string]string{"username": c.Args()[0], "password": c.Args()[1]})
	if err != nil {
		return err
	}

	cmd.ui.Ok()
	cmd.ui.Say(T("Use '{{.Name}}' to view or set your target org and space",
		map[string]interface{}{"Name": terminal.CommandColor(cf.Name + " target")}))

	cmd.ui.NotifyUpdateIfNeeded(cmd.config)

	return nil
}
