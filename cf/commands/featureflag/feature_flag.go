package featureflag

import (
	"github.com/cloudfoundry/cli/cf/api/feature_flags"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type ShowFeatureFlag struct {
	ui       terminal.UI
	config   core_config.ReadWriter
	flagRepo feature_flags.FeatureFlagRepository
}

func NewShowFeatureFlag(
	ui terminal.UI,
	config core_config.ReadWriter,
	flagRepo feature_flags.FeatureFlagRepository) (cmd ShowFeatureFlag) {
	return ShowFeatureFlag{
		ui:       ui,
		config:   config,
		flagRepo: flagRepo,
	}
}

func (cmd ShowFeatureFlag) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "feature-flag",
		Description: T("Retrieve an individual feature flag with status"),
		Usage:       T("CF_NAME feature-flag FEATURE_NAME"),
	}
}

func (cmd ShowFeatureFlag) GetRequirements(requirementsFactory requirements.Factory, context *cli.Context) ([]requirements.Requirement, error) {
	if len(context.Args()) != 1 {
		cmd.ui.FailWithUsage(context)
	}

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}
	return reqs, nil
}

func (cmd ShowFeatureFlag) Run(c *cli.Context) {
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
	} else {
		return "disabled"
	}
}
