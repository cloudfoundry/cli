package featureflag

import (
	"fmt"

	"code.cloudfoundry.org/cli/v7/cf/api/featureflags"
	"code.cloudfoundry.org/cli/v7/cf/commandregistry"
	"code.cloudfoundry.org/cli/v7/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/v7/cf/flags"
	. "code.cloudfoundry.org/cli/v7/cf/i18n"
	"code.cloudfoundry.org/cli/v7/cf/requirements"
	"code.cloudfoundry.org/cli/v7/cf/terminal"
)

type EnableFeatureFlag struct {
	ui       terminal.UI
	config   coreconfig.ReadWriter
	flagRepo featureflags.FeatureFlagRepository
}

func init() {
	commandregistry.Register(&EnableFeatureFlag{})
}

func (cmd *EnableFeatureFlag) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "enable-feature-flag",
		Description: T("Enable the use of a feature so that users have access to and can use the feature"),
		Usage: []string{
			T("CF_NAME enable-feature-flag FEATURE_NAME"),
		},
	}
}

func (cmd *EnableFeatureFlag) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + commandregistry.Commands.CommandUsage("enable-feature-flag"))
		return nil, fmt.Errorf("Incorrect usage: %d arguments of %d required", len(fc.Args()), 1)
	}

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}

	return reqs, nil
}

func (cmd *EnableFeatureFlag) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.flagRepo = deps.RepoLocator.GetFeatureFlagRepository()
	return cmd
}

func (cmd *EnableFeatureFlag) Execute(c flags.FlagContext) error {
	flag := c.Args()[0]

	cmd.ui.Say(T("Setting status of {{.FeatureFlag}} as {{.Username}}...", map[string]interface{}{
		"FeatureFlag": terminal.EntityNameColor(flag),
		"Username":    terminal.EntityNameColor(cmd.config.Username())}))

	err := cmd.flagRepo.Update(flag, true)
	if err != nil {
		return err
	}

	cmd.ui.Say("")
	cmd.ui.Ok()
	cmd.ui.Say("")
	cmd.ui.Say(T("Feature {{.FeatureFlag}} Enabled.", map[string]interface{}{
		"FeatureFlag": terminal.EntityNameColor(flag)}))
	return nil
}
