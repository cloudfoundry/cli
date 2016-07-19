package environmentvariablegroup

import (
	"errors"
	"fmt"

	"code.cloudfoundry.org/cli/cf/api/environmentvariablegroups"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	cf_errors "code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/flags"
	. "code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
)

type SetStagingEnvironmentVariableGroup struct {
	ui                           terminal.UI
	config                       coreconfig.ReadWriter
	environmentVariableGroupRepo environmentvariablegroups.Repository
}

func init() {
	commandregistry.Register(&SetStagingEnvironmentVariableGroup{})
}

func (cmd *SetStagingEnvironmentVariableGroup) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "set-staging-environment-variable-group",
		Description: T("Pass parameters as JSON to create a staging environment variable group"),
		ShortName:   "ssevg",
		Usage: []string{
			T(`CF_NAME set-staging-environment-variable-group '{"name":"value","name":"value"}'`),
		},
	}
}

func (cmd *SetStagingEnvironmentVariableGroup) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + commandregistry.Commands.CommandUsage("set-staging-environment-variable-group"))
		return nil, fmt.Errorf("Incorrect usage: %d arguments of %d required", len(fc.Args()), 1)
	}

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}
	return reqs, nil
}

func (cmd *SetStagingEnvironmentVariableGroup) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.environmentVariableGroupRepo = deps.RepoLocator.GetEnvironmentVariableGroupsRepository()
	return cmd
}

func (cmd *SetStagingEnvironmentVariableGroup) Execute(c flags.FlagContext) error {
	cmd.ui.Say(T("Setting the contents of the staging environment variable group as {{.Username}}...", map[string]interface{}{
		"Username": terminal.EntityNameColor(cmd.config.Username())}))

	err := cmd.environmentVariableGroupRepo.SetStaging(c.Args()[0])
	if err != nil {
		suggestionText := ""

		httpError, ok := err.(cf_errors.HTTPError)
		if ok && httpError.ErrorCode() == cf_errors.MessageParseError {
			suggestionText = T(`

Your JSON string syntax is invalid.  Proper syntax is this:  cf set-staging-environment-variable-group '{"name":"value","name":"value"}'`)
		}
		return errors.New(err.Error() + suggestionText)
	}

	cmd.ui.Ok()
	return nil
}
