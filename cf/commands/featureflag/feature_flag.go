package featureflag

import (
	"github.com/cloudfoundry/cli/cf/api/feature_flags"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
)

type ShowFeatureFlag struct {
	ui       terminal.UI
	config   core_config.ReadWriter
	flagRepo feature_flags.FeatureFlagRepository
}

func init() {
	command_registry.Register(&ShowFeatureFlag{})
}

func (cmd *ShowFeatureFlag) MetaData() command_registry.CommandMetadata {
	return command_registry.CommandMetadata{
		Name:        "feature-flag",
		Description: T("Retrieve an individual feature flag with status"),
		Usage:       T("CF_NAME feature-flag FEATURE_NAME"),
	}
}

func (cmd *ShowFeatureFlag) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + command_registry.Commands.CommandUsage("feature-flag"))
	}

	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}
	return reqs, nil
}

func (cmd *ShowFeatureFlag) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.flagRepo = deps.RepoLocator.GetFeatureFlagRepository()
	return cmd
}

func (cmd *ShowFeatureFlag) Execute(c flags.FlagContext) {
	flagName := c.Args()[0]

	cmd.ui.Say(T("Retrieving status of {{.FeatureFlag}} as {{.Username}}...", map[string]interface{}{
		"FeatureFlag": terminal.EntityNameColor(flagName),
		"Username":    terminal.EntityNameColor(cmd.config.Username())}))

	flag, err := cmd.flagRepo.FindByName(flagName)
	if err != nil {
		cmd.ui.Failed(err.Error())
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say("")

	table := terminal.NewTable(cmd.ui, []string{T("Features"), T("State")})
	table.Add(flag.Name, cmd.flagBoolToString(flag.Enabled))

	table.Print()
	return
}

func (cmd ShowFeatureFlag) flagBoolToString(enabled bool) string {
	if enabled {
		return "enabled"
	}
	return "disabled"
}
