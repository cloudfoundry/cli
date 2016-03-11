package serviceauthtoken

import (
	"github.com/blang/semver"
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
)

type CreateServiceAuthTokenFields struct {
	ui            terminal.UI
	config        core_config.Reader
	authTokenRepo api.ServiceAuthTokenRepository
}

func init() {
	command_registry.Register(&CreateServiceAuthTokenFields{})
}

func (cmd *CreateServiceAuthTokenFields) MetaData() command_registry.CommandMetadata {
	return command_registry.CommandMetadata{
		Name:        "create-service-auth-token",
		Description: T("Create a service auth token"),
		Usage: []string{
			T("CF_NAME create-service-auth-token LABEL PROVIDER TOKEN"),
		},
	}
}

func (cmd *CreateServiceAuthTokenFields) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	if len(fc.Args()) != 3 {
		cmd.ui.Failed(T("Incorrect Usage. Requires LABEL, PROVIDER and TOKEN as arguments\n\n") + command_registry.Commands.CommandUsage("create-service-auth-token"))
	}

	maximumVersion, err := semver.Make("2.46.0")
	if err != nil {
		panic(err.Error())
	}

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewMaxAPIVersionRequirement("create-service-auth-token", maximumVersion),
	}

	return reqs
}

func (cmd *CreateServiceAuthTokenFields) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.authTokenRepo = deps.RepoLocator.GetServiceAuthTokenRepository()
	return cmd
}

func (cmd *CreateServiceAuthTokenFields) Execute(c flags.FlagContext) {
	cmd.ui.Say(T("Creating service auth token as {{.CurrentUser}}...",
		map[string]interface{}{
			"CurrentUser": terminal.EntityNameColor(cmd.config.Username()),
		}))

	serviceAuthTokenRepo := models.ServiceAuthTokenFields{
		Label:    c.Args()[0],
		Provider: c.Args()[1],
		Token:    c.Args()[2],
	}

	apiErr := cmd.authTokenRepo.Create(serviceAuthTokenRepo)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Ok()
}
