package featureflag

import (
	"code.cloudfoundry.org/cli/cf/api/featureflags"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/flags"
	. "code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
)

type ListFeatureFlags struct {
	ui       terminal.UI
	config   coreconfig.ReadWriter
	flagRepo featureflags.FeatureFlagRepository
}

func init() {
	commandregistry.Register(&ListFeatureFlags{})
}

func (cmd *ListFeatureFlags) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "feature-flags",
		Description: T("Retrieve list of feature flags with status of each flag-able feature"),
		Usage: []string{
			T("CF_NAME feature-flags"),
		},
	}
}

func (cmd *ListFeatureFlags) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
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

func (cmd *ListFeatureFlags) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.flagRepo = deps.RepoLocator.GetFeatureFlagRepository()
	return cmd
}

func (cmd *ListFeatureFlags) Execute(c flags.FlagContext) error {
	cmd.ui.Say(T("Retrieving status of all flagged features as {{.Username}}...", map[string]interface{}{
		"Username": terminal.EntityNameColor(cmd.config.Username())}))

	flags, err := cmd.flagRepo.List()
	if err != nil {
		return err
	}

	cmd.ui.Ok()
	cmd.ui.Say("")

	table := cmd.ui.Table([]string{T("Features"), T("State")})

	for _, flag := range flags {
		table.Add(
			flag.Name,
			cmd.flagBoolToString(flag.Enabled),
		)
	}

	err = table.Print()
	if err != nil {
		return err
	}
	return nil
}

func (cmd ListFeatureFlags) flagBoolToString(enabled bool) string {
	if enabled {
		return "enabled"
	}
	return "disabled"
}
