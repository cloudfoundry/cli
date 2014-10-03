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

type DisableFeatureFlag struct {
	ui       terminal.UI
	config   core_config.ReadWriter
	flagRepo feature_flags.FeatureFlagRepository
}

func NewDisableFeatureFlag(
	ui terminal.UI,
	config core_config.ReadWriter,
	flagRepo feature_flags.FeatureFlagRepository) (cmd DisableFeatureFlag) {
	return DisableFeatureFlag{
		ui:       ui,
		config:   config,
		flagRepo: flagRepo,
	}
}

func (cmd DisableFeatureFlag) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "disable-feature-flag",
		Description: T("Disable the use of a feature so that users have access to and can use the feature."),
		Usage:       T("CF_NAME disable-feature-flag FEATURE_NAME"),
	}
}

func (cmd DisableFeatureFlag) GetRequirements(requirementsFactory requirements.Factory, context *cli.Context) ([]requirements.Requirement, error) {
	if len(context.Args()) != 1 {
		cmd.ui.FailWithUsage(context)
	}

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}
	return reqs, nil
}

func (cmd DisableFeatureFlag) Run(c *cli.Context) {
	flag := c.Args()[0]

	cmd.ui.Say(T("Setting status of {{.FeatureFlag}} as {{.Username}}...", map[string]interface{}{
		"FeatureFlag": terminal.EntityNameColor(flag),
		"Username":    terminal.EntityNameColor(cmd.config.Username())}))

	apiErr := cmd.flagRepo.Update(flag, false)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
	}

	cmd.ui.Say("")
	cmd.ui.Ok()
	cmd.ui.Say("")
	cmd.ui.Say(T("Feature {{.FeatureFlag}} Disabled.", map[string]interface{}{
		"FeatureFlag": terminal.EntityNameColor(flag)}))
	return
}
