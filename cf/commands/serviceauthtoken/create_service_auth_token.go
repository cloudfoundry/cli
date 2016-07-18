package serviceauthtoken

import (
	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/flags"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type CreateServiceAuthTokenFields struct {
	ui            terminal.UI
	config        coreconfig.Reader
	authTokenRepo api.ServiceAuthTokenRepository
}

func init() {
	commandregistry.Register(&CreateServiceAuthTokenFields{})
}

func (cmd *CreateServiceAuthTokenFields) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "create-service-auth-token",
		Description: T("Create a service auth token"),
		Usage: []string{
			T("CF_NAME create-service-auth-token LABEL PROVIDER TOKEN"),
		},
	}
}

func (cmd *CreateServiceAuthTokenFields) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	if len(fc.Args()) != 3 {
		cmd.ui.Failed(T("Incorrect Usage. Requires LABEL, PROVIDER and TOKEN as arguments\n\n") + commandregistry.Commands.CommandUsage("create-service-auth-token"))
	}

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewMaxAPIVersionRequirement(
			"create-service-auth-token",
			cf.ServiceAuthTokenMaximumAPIVersion,
		),
	}

	return reqs
}

func (cmd *CreateServiceAuthTokenFields) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.authTokenRepo = deps.RepoLocator.GetServiceAuthTokenRepository()
	return cmd
}

func (cmd *CreateServiceAuthTokenFields) Execute(c flags.FlagContext) error {
	cmd.ui.Say(T("Creating service auth token as {{.CurrentUser}}...",
		map[string]interface{}{
			"CurrentUser": terminal.EntityNameColor(cmd.config.Username()),
		}))

	serviceAuthTokenRepo := models.ServiceAuthTokenFields{
		Label:    c.Args()[0],
		Provider: c.Args()[1],
		Token:    c.Args()[2],
	}

	err := cmd.authTokenRepo.Create(serviceAuthTokenRepo)
	if err != nil {
		return err
	}

	cmd.ui.Ok()
	return nil
}
