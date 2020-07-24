package environmentvariablegroup

import (
	"sort"

	"code.cloudfoundry.org/cli/cf/api/environmentvariablegroups"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/flags"
	. "code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
)

type RunningEnvironmentVariableGroup struct {
	ui                           terminal.UI
	config                       coreconfig.ReadWriter
	environmentVariableGroupRepo environmentvariablegroups.Repository
}

func init() {
	commandregistry.Register(&RunningEnvironmentVariableGroup{})
}

func (cmd *RunningEnvironmentVariableGroup) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "running-environment-variable-group",
		Description: T("Retrieve the contents of the running environment variable group"),
		ShortName:   "revg",
		Usage: []string{
			T("CF_NAME running-environment-variable-group"),
		},
	}
}

func (cmd *RunningEnvironmentVariableGroup) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	usageReq := requirements.NewUsageRequirement(commandregistry.CLICommandUsagePresenter(cmd),
		T("No argument required"),
		func() bool {
			return len(fc.Args()) != 0
		},
	)

	reqs := []requirements.Requirement{
		usageReq,
		requirementsFactory.NewLoginRequirement(),
	}
	return reqs, nil
}

func (cmd *RunningEnvironmentVariableGroup) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.environmentVariableGroupRepo = deps.RepoLocator.GetEnvironmentVariableGroupsRepository()
	return cmd
}

func (cmd *RunningEnvironmentVariableGroup) Execute(c flags.FlagContext) error {
	cmd.ui.Say(T("Retrieving the contents of the running environment variable group as {{.Username}}...", map[string]interface{}{
		"Username": terminal.EntityNameColor(cmd.config.Username())}))

	runningEnvVars, err := cmd.environmentVariableGroupRepo.ListRunning()
	if err != nil {
		return err
	}

	cmd.ui.Ok()

	table := cmd.ui.Table([]string{T("Variable Name"), T("Assigned Value")})
	sortedEnvVars := models.EnvironmentVariableList(runningEnvVars)
	sort.Sort(sortedEnvVars)
	for _, envVar := range sortedEnvVars {
		table.Add(envVar.Name, envVar.Value)
	}
	err = table.Print()
	if err != nil {
		return err
	}
	return nil
}
