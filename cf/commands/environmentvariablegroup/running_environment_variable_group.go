package environmentvariablegroup

import (
	"github.com/cloudfoundry/cli/cf/api/environment_variable_groups"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type RunningEnvironmentVariableGroup struct {
	ui                           terminal.UI
	config                       core_config.ReadWriter
	environmentVariableGroupRepo environment_variable_groups.EnvironmentVariableGroupsRepository
}

func NewRunningEnvironmentVariableGroup(ui terminal.UI, config core_config.ReadWriter, environmentVariableGroupRepo environment_variable_groups.EnvironmentVariableGroupsRepository) (cmd RunningEnvironmentVariableGroup) {
	cmd.ui = ui
	cmd.config = config
	cmd.environmentVariableGroupRepo = environmentVariableGroupRepo
	return
}

func (cmd RunningEnvironmentVariableGroup) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "running-environment-variable-group",
		Description: T("Retrieve the contents of the running environment variable group"),
		ShortName:   "revg",
		Usage:       T("CF_NAME running-environment-variable-group"),
	}
}

func (cmd RunningEnvironmentVariableGroup) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) ([]requirements.Requirement, error) {
	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}
	return reqs, nil
}

func (cmd RunningEnvironmentVariableGroup) Run(c *cli.Context) {
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
