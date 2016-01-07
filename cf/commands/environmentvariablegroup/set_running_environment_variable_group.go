package environmentvariablegroup

import (
	"github.com/cloudfoundry/cli/cf/api/environment_variable_groups"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	cf_errors "github.com/cloudfoundry/cli/cf/errors"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
)

type SetRunningEnvironmentVariableGroup struct {
	ui                           terminal.UI
	config                       core_config.ReadWriter
	environmentVariableGroupRepo environment_variable_groups.EnvironmentVariableGroupsRepository
}

func init() {
	command_registry.Register(&SetRunningEnvironmentVariableGroup{})
}

func (cmd *SetRunningEnvironmentVariableGroup) MetaData() command_registry.CommandMetadata {
	return command_registry.CommandMetadata{
		Name:        "set-running-environment-variable-group",
		Description: T("Pass parameters as JSON to create a running environment variable group"),
		ShortName:   "srevg",
		Usage:       T(`CF_NAME set-running-environment-variable-group '{"name":"value","name":"value"}'`),
	}
}

func (cmd *SetRunningEnvironmentVariableGroup) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + command_registry.Commands.CommandUsage("set-running-environment-variable-group"))
	}

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}
	return reqs, nil
}

func (cmd *SetRunningEnvironmentVariableGroup) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.environmentVariableGroupRepo = deps.RepoLocator.GetEnvironmentVariableGroupsRepository()
	return cmd
}

func (cmd *SetRunningEnvironmentVariableGroup) Execute(c flags.FlagContext) {
	cmd.ui.Say(T("Setting the contents of the running environment variable group as {{.Username}}...", map[string]interface{}{
		"Username": terminal.EntityNameColor(cmd.config.Username())}))

	err := cmd.environmentVariableGroupRepo.SetRunning(c.Args()[0])
	if err != nil {
		suggestionText := ""

		httpError, ok := err.(cf_errors.HttpError)
		if ok && httpError.ErrorCode() == cf_errors.PARSE_ERROR {
			suggestionText = T(`

Your JSON string syntax is invalid.  Proper syntax is this:  cf set-running-environment-variable-group '{"name":"value","name":"value"}'`)
		}
		cmd.ui.Failed(err.Error() + suggestionText)
	}

	cmd.ui.Ok()
}
