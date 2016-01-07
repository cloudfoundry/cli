package serviceauthtoken

import (
	"fmt"

	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/flags"
	"github.com/cloudfoundry/cli/flags/flag"

	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type DeleteServiceAuthTokenFields struct {
	ui            terminal.UI
	config        core_config.Reader
	authTokenRepo api.ServiceAuthTokenRepository
}

func init() {
	command_registry.Register(&DeleteServiceAuthTokenFields{})
}

func (cmd *DeleteServiceAuthTokenFields) MetaData() command_registry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["f"] = &cliFlags.BoolFlag{ShortName: "f", Usage: T("Force deletion without confirmation")}

	return command_registry.CommandMetadata{
		Name:        "delete-service-auth-token",
		Description: T("Delete a service auth token"),
		Usage:       T("CF_NAME delete-service-auth-token LABEL PROVIDER [-f]"),
		Flags:       fs,
	}
}

func (cmd *DeleteServiceAuthTokenFields) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(fc.Args()) != 2 {
		cmd.ui.Failed(T("Incorrect Usage. Requires LABEL, PROVIDER as arguments\n\n") + command_registry.Commands.CommandUsage("delete-service-auth-token"))
	}

	reqs = append(reqs, requirementsFactory.NewLoginRequirement())
	return
}

func (cmd *DeleteServiceAuthTokenFields) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.authTokenRepo = deps.RepoLocator.GetServiceAuthTokenRepository()
	return cmd
}

func (cmd *DeleteServiceAuthTokenFields) Execute(c flags.FlagContext) {
	tokenLabel := c.Args()[0]
	tokenProvider := c.Args()[1]

	if c.Bool("f") == false {
		if !cmd.ui.ConfirmDelete(T("service auth token"), fmt.Sprintf("%s %s", tokenLabel, tokenProvider)) {
			return
		}
	}

	cmd.ui.Say(T("Deleting service auth token as {{.CurrentUser}}",
		map[string]interface{}{
			"CurrentUser": terminal.EntityNameColor(cmd.config.Username()),
		}))
	token, apiErr := cmd.authTokenRepo.FindByLabelAndProvider(tokenLabel, tokenProvider)

	switch apiErr.(type) {
	case nil:
	case *errors.ModelNotFoundError:
		cmd.ui.Ok()
		cmd.ui.Warn(T("Service Auth Token {{.Label}} {{.Provider}} does not exist.", map[string]interface{}{"Label": tokenLabel, "Provider": tokenProvider}))
		return
	default:
		cmd.ui.Failed(apiErr.Error())
	}

	apiErr = cmd.authTokenRepo.Delete(token)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Ok()
}
