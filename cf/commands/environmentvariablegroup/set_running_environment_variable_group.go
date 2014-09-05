package environmentvariablegroup

import (
	"github.com/cloudfoundry/cli/cf/api/environment_variable_groups"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	cf_errors "github.com/cloudfoundry/cli/cf/errors"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type SetRunningEnvironmentVariableGroup struct {
	ui                           terminal.UI
	config                       configuration.ReadWriter
	environmentVariableGroupRepo environment_variable_groups.EnvironmentVariableGroupsRepository
}

func NewSetRunningEnvironmentVariableGroup(ui terminal.UI, config configuration.ReadWriter, environmentVariableGroupRepo environment_variable_groups.EnvironmentVariableGroupsRepository) (cmd SetRunningEnvironmentVariableGroup) {
	cmd.ui = ui
	cmd.config = config
	cmd.environmentVariableGroupRepo = environmentVariableGroupRepo
	return
}

func (cmd SetRunningEnvironmentVariableGroup) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "set-running-environment-variable-group",
		Description: T("Pass parameters as JSON to create a running environment variable group"),
		ShortName:   "srevg",
		Usage:       T(`CF_NAME set-running-environment-variable-group '{"name":"value","name":"value"}'`),
	}
}

func (cmd SetRunningEnvironmentVariableGroup) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) ([]requirements.Requirement, error) {
	if len(c.Args()) != 1 {
		cmd.ui.FailWithUsage(c)
	}

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}
	return reqs, nil
}

func (cmd SetRunningEnvironmentVariableGroup) Run(c *cli.Context) {
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
