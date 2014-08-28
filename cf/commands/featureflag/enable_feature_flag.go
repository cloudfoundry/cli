package featureflag

import (
	"fmt"

	"github.com/cloudfoundry/cli/cf/api/feature_flags"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type EnableFeatureFlag struct {
	ui       terminal.UI
	config   configuration.ReadWriter
	flagRepo feature_flags.FeatureFlagRepository
}

func NewEnableFeatureFlag(
	ui terminal.UI,
	config configuration.ReadWriter,
	flagRepo feature_flags.FeatureFlagRepository) (cmd EnableFeatureFlag) {
	return EnableFeatureFlag{
		ui:       ui,
		config:   config,
		flagRepo: flagRepo,
	}
}

func (cmd EnableFeatureFlag) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "enable-feature-flag",
		Description: T("Enable the use of a feature so that users have access to and can use the feature."),
		Usage:       T("CF_NAME enable-feature-flag FEATURE_NAME"),
	}
}

func (cmd EnableFeatureFlag) GetRequirements(requirementsFactory requirements.Factory, context *cli.Context) ([]requirements.Requirement, error) {
	if len(context.Args()) != 1 {
		cmd.ui.FailWithUsage(context)
	}

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}
	return reqs, nil
}

func (cmd EnableFeatureFlag) Run(c *cli.Context) {
	flag := c.Args()[0]

	cmd.ui.Say(T("Setting status of {{.FeatureFlag}} as {{.Username}}...", map[string]interface{}{
		"FeatureFlag": terminal.EntityNameColor(flag),
		"Username":    terminal.EntityNameColor(cmd.config.Username())}))

	apiErr := cmd.flagRepo.Update(flag, true)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
	}

	cmd.ui.Say("")
	cmd.ui.Ok()
	cmd.ui.Say("")
	cmd.ui.Say(fmt.Sprintf("Feature %s Enabled.", terminal.EntityNameColor(flag)))
	return
}
