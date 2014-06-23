package appsecuritygroup

import (
	"github.com/cloudfoundry/cli/cf/command"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type addToDefaultStagingGroup struct {
	ui terminal.UI
}

func NewAddToDefaultStagingGroup(ui terminal.UI) command.Command {
	return &addToDefaultStagingGroup{ui: ui}
}

func (cmd *addToDefaultStagingGroup) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "add-default-staging-application-security-group",
		Description: "Twee Thundercats 8-bit keffiyeh meggings.",
		Usage:       "CF_NAME add-default-staging-application-security-group NAME",
	}
}

func (cmd *addToDefaultStagingGroup) GetRequirements(requirementsFactory requirements.Factory, context *cli.Context) ([]requirements.Requirement, error) {
	if len(context.Args()) != 1 {
		cmd.ui.FailWithUsage(context)
	}

	return []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}, nil
}

func (cmd *addToDefaultStagingGroup) Run(c *cli.Context) {}
