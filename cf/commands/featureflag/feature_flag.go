package featureflag

import (
	"github.com/cloudfoundry/cli/cf/api/featureflags"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/flags"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type ShowFeatureFlag struct {
	ui       terminal.UI
	config   coreconfig.ReadWriter
	flagRepo featureflags.FeatureFlagRepository
}

func init() {
	commandregistry.Register(&ShowFeatureFlag{})
}

func (cmd *ShowFeatureFlag) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "feature-flag",
		Description: T("Retrieve an individual feature flag with status"),
		Usage: []string{
			T("CF_NAME feature-flag FEATURE_NAME"),
		},
	}
}

func (cmd *ShowFeatureFlag) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + commandregistry.Commands.CommandUsage("feature-flag"))
	}

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}

	return reqs
}

func (cmd *ShowFeatureFlag) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.flagRepo = deps.RepoLocator.GetFeatureFlagRepository()
	return cmd
}

func (cmd *ShowFeatureFlag) Execute(c flags.FlagContext) error {
	flagName := c.Args()[0]

	cmd.ui.Say(T("Retrieving status of {{.FeatureFlag}} as {{.Username}}...", map[string]interface{}{
		"FeatureFlag": terminal.EntityNameColor(flagName),
		"Username":    terminal.EntityNameColor(cmd.config.Username())}))

	flag, err := cmd.flagRepo.FindByName(flagName)
	if err != nil {
		return err
	}

	cmd.ui.Ok()
	cmd.ui.Say("")

	table := cmd.ui.Table([]string{T("Features"), T("State")})
	table.Add(flag.Name, cmd.flagBoolToString(flag.Enabled))

	table.Print()
	return nil
}

func (cmd ShowFeatureFlag) flagBoolToString(enabled bool) string {
	if enabled {
		return "enabled"
	}
	return "disabled"
}
