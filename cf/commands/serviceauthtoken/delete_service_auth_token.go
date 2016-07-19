package serviceauthtoken

import (
	"fmt"

	"code.cloudfoundry.org/cli/cf"
	"code.cloudfoundry.org/cli/cf/flags"
	. "code.cloudfoundry.org/cli/cf/i18n"

	"code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
)

type DeleteServiceAuthTokenFields struct {
	ui            terminal.UI
	config        coreconfig.Reader
	authTokenRepo api.ServiceAuthTokenRepository
}

func init() {
	commandregistry.Register(&DeleteServiceAuthTokenFields{})
}

func (cmd *DeleteServiceAuthTokenFields) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["f"] = &flags.BoolFlag{ShortName: "f", Usage: T("Force deletion without confirmation")}

	return commandregistry.CommandMetadata{
		Name:        "delete-service-auth-token",
		Description: T("Delete a service auth token"),
		Usage: []string{
			T("CF_NAME delete-service-auth-token LABEL PROVIDER [-f]"),
		},
		Flags: fs,
	}
}

func (cmd *DeleteServiceAuthTokenFields) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 2 {
		cmd.ui.Failed(T("Incorrect Usage. Requires LABEL, PROVIDER as arguments\n\n") + commandregistry.Commands.CommandUsage("delete-service-auth-token"))
		return nil, fmt.Errorf("Incorrect usage: %d arguments of %d required", len(fc.Args()), 2)
	}

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewMaxAPIVersionRequirement(
			"delete-service-auth-token",
			cf.ServiceAuthTokenMaximumAPIVersion,
		),
	}

	return reqs, nil
}

func (cmd *DeleteServiceAuthTokenFields) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.authTokenRepo = deps.RepoLocator.GetServiceAuthTokenRepository()
	return cmd
}

func (cmd *DeleteServiceAuthTokenFields) Execute(c flags.FlagContext) error {
	tokenLabel := c.Args()[0]
	tokenProvider := c.Args()[1]

	if c.Bool("f") == false {
		if !cmd.ui.ConfirmDelete(T("service auth token"), fmt.Sprintf("%s %s", tokenLabel, tokenProvider)) {
			return nil
		}
	}

	cmd.ui.Say(T("Deleting service auth token as {{.CurrentUser}}",
		map[string]interface{}{
			"CurrentUser": terminal.EntityNameColor(cmd.config.Username()),
		}))
	token, err := cmd.authTokenRepo.FindByLabelAndProvider(tokenLabel, tokenProvider)

	switch err.(type) {
	case nil:
	case *errors.ModelNotFoundError:
		cmd.ui.Ok()
		cmd.ui.Warn(T("Service Auth Token {{.Label}} {{.Provider}} does not exist.", map[string]interface{}{"Label": tokenLabel, "Provider": tokenProvider}))
		return nil
	default:
		return err
	}

	err = cmd.authTokenRepo.Delete(token)
	if err != nil {
		return err
	}

	cmd.ui.Ok()
	return nil
}
