package serviceauthtoken

import (
	"fmt"

	"code.cloudfoundry.org/cli/cf"
	"code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/flags"
	. "code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
)

type UpdateServiceAuthTokenFields struct {
	ui            terminal.UI
	config        coreconfig.Reader
	authTokenRepo api.ServiceAuthTokenRepository
}

func init() {
	commandregistry.Register(&UpdateServiceAuthTokenFields{})
}

func (cmd *UpdateServiceAuthTokenFields) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "update-service-auth-token",
		Description: T("Update a service auth token"),
		Usage: []string{
			T("CF_NAME update-service-auth-token LABEL PROVIDER TOKEN"),
		},
	}
}

func (cmd *UpdateServiceAuthTokenFields) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 3 {
		cmd.ui.Failed(T("Incorrect Usage. Requires LABEL, PROVIDER and TOKEN as arguments\n\n") + commandregistry.Commands.CommandUsage("update-service-auth-token"))
		return nil, fmt.Errorf("Incorrect usage: %d arguments of %d required", len(fc.Args()), 3)
	}

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewMaxAPIVersionRequirement(
			"update-service-auth-token",
			cf.ServiceAuthTokenMaximumAPIVersion,
		),
	}

	return reqs, nil
}

func (cmd *UpdateServiceAuthTokenFields) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.authTokenRepo = deps.RepoLocator.GetServiceAuthTokenRepository()
	return cmd
}

func (cmd *UpdateServiceAuthTokenFields) Execute(c flags.FlagContext) error {
	cmd.ui.Say(T("Updating service auth token as {{.CurrentUser}}...", map[string]interface{}{"CurrentUser": terminal.EntityNameColor(cmd.config.Username())}))

	serviceAuthToken, err := cmd.authTokenRepo.FindByLabelAndProvider(c.Args()[0], c.Args()[1])
	if err != nil {
		return err
	}

	serviceAuthToken.Token = c.Args()[2]

	err = cmd.authTokenRepo.Update(serviceAuthToken)
	if err != nil {
		return err
	}

	cmd.ui.Ok()
	return nil
}
