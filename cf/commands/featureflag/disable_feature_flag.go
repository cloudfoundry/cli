package featureflag

import (
	"fmt"

	"code.cloudfoundry.org/cli/cf/api/featureflags"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/flags"
	. "code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
)

type DisableFeatureFlag struct {
	ui       terminal.UI
	config   coreconfig.ReadWriter
	flagRepo featureflags.FeatureFlagRepository
}

func init() {
	commandregistry.Register(&DisableFeatureFlag{})
}

func (cmd *DisableFeatureFlag) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "disable-feature-flag",
		Description: T("Disable the use of a feature so that users have access to and can use the feature"),
		Usage: []string{
			T("CF_NAME disable-feature-flag FEATURE_NAME"),
		},
	}
}

func (cmd *DisableFeatureFlag) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + commandregistry.Commands.CommandUsage("disable-feature-flag"))
		return nil, fmt.Errorf("Incorrect usage: %d arguments of %d required", len(fc.Args()), 1)
	}

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}

	return reqs, nil
}

func (cmd *DisableFeatureFlag) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.flagRepo = deps.RepoLocator.GetFeatureFlagRepository()
	return cmd
}

func (cmd *DisableFeatureFlag) Execute(c flags.FlagContext) error {
	flag := c.Args()[0]

	cmd.ui.Say(T("Setting status of {{.FeatureFlag}} as {{.Username}}...", map[string]interface{}{
		"FeatureFlag": terminal.EntityNameColor(flag),
		"Username":    terminal.EntityNameColor(cmd.config.Username())}))

	err := cmd.flagRepo.Update(flag, false)
	if err != nil {
		return err
	}

	cmd.ui.Say("")
	cmd.ui.Ok()
	cmd.ui.Say("")
	cmd.ui.Say(T("Feature {{.FeatureFlag}} Disabled.", map[string]interface{}{
		"FeatureFlag": terminal.EntityNameColor(flag)}))
	return nil
}
