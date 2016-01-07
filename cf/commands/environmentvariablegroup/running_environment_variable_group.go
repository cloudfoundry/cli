package environmentvariablegroup

import (
	"github.com/cloudfoundry/cli/cf/api/environment_variable_groups"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
)

type RunningEnvironmentVariableGroup struct {
	ui                           terminal.UI
	config                       core_config.ReadWriter
	environmentVariableGroupRepo environment_variable_groups.EnvironmentVariableGroupsRepository
}

func init() {
	command_registry.Register(&RunningEnvironmentVariableGroup{})
}

func (cmd *RunningEnvironmentVariableGroup) MetaData() command_registry.CommandMetadata {
	return command_registry.CommandMetadata{
		Name:        "running-environment-variable-group",
		Description: T("Retrieve the contents of the running environment variable group"),
		ShortName:   "revg",
		Usage:       T("CF_NAME running-environment-variable-group"),
	}
}

func (cmd *RunningEnvironmentVariableGroup) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 0 {
		cmd.ui.Failed(T("Incorrect Usage. No argument required\n\n") + command_registry.Commands.CommandUsage("running-environment-variable-group"))
	}

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}
	return reqs, nil
}

func (cmd *RunningEnvironmentVariableGroup) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.environmentVariableGroupRepo = deps.RepoLocator.GetEnvironmentVariableGroupsRepository()
	return cmd
}

func (cmd *RunningEnvironmentVariableGroup) Execute(c flags.FlagContext) {
	cmd.ui.Say(T("Retrieving the contents of the running environment variable group as {{.Username}}...", map[string]interface{}{
		"Username": terminal.EntityNameColor(cmd.config.Username())}))

	runningEnvVars, err := cmd.environmentVariableGroupRepo.ListRunning()
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	cmd.ui.Ok()

	table := terminal.NewTable(cmd.ui, []string{T("Variable Name"), T("Assigned Value")})
	for _, envVar := range runningEnvVars {
		table.Add(envVar.Name, envVar.Value)
	}
	table.Print()
}
