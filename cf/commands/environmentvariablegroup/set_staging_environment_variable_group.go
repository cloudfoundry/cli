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

type SetStagingEnvironmentVariableGroup struct {
	ui                           terminal.UI
	config                       configuration.ReadWriter
	environmentVariableGroupRepo environment_variable_groups.EnvironmentVariableGroupsRepository
}

func NewSetStagingEnvironmentVariableGroup(ui terminal.UI, config configuration.ReadWriter, environmentVariableGroupRepo environment_variable_groups.EnvironmentVariableGroupsRepository) (cmd SetStagingEnvironmentVariableGroup) {
	cmd.ui = ui
	cmd.config = config
	cmd.environmentVariableGroupRepo = environmentVariableGroupRepo
	return
}

func (cmd SetStagingEnvironmentVariableGroup) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "set-staging-environment-variable-group",
		Description: T("Pass parameters as JSON to create a staging environment variable group"),
		ShortName:   "ssevg",
		Usage:       T("CF_NAME set-staging-environment-variable-group"),
	}
}

func (cmd SetStagingEnvironmentVariableGroup) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) ([]requirements.Requirement, error) {
	if len(c.Args()) != 1 {
		cmd.ui.FailWithUsage(c)
	}

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}
	return reqs, nil
}

func (cmd SetStagingEnvironmentVariableGroup) Run(c *cli.Context) {
	cmd.ui.Say(T("Setting the contents of the staging environment variable group as {{.Username}}...", map[string]interface{}{
		"Username": terminal.EntityNameColor(cmd.config.Username())}))

	err := cmd.environmentVariableGroupRepo.SetStaging(c.Args()[0])
	if err != nil {
		suggestionText := ""

		httpError, ok := err.(cf_errors.HttpError)
		if ok && httpError.ErrorCode() == cf_errors.PARSE_ERROR {
			suggestionText = T(`

Your JSON string syntax is invalid.  Proper syntax is this:  cf set-staging-environment-variable-group '{"name":"value","name":"value"}'`)
		}
		cmd.ui.Failed(err.Error() + suggestionText)
	}

	cmd.ui.Ok()
}
