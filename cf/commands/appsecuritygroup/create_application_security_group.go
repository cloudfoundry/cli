package appsecuritygroup

import (
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type CreateAppSecurityGroup struct {
	ui terminal.UI
}

func NewCreateAppSecurityGroup(ui terminal.UI) CreateAppSecurityGroup {
	return CreateAppSecurityGroup{
		ui: ui,
	}
}

func (cmd CreateAppSecurityGroup) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "create-application-security-group",
		Description: "<<< description goes here>>>",
		Usage:       "CF_NAME create-application-security-group NAME",
	}
}

func (cmd CreateAppSecurityGroup) GetRequirements(requirementsFactory requirements.Factory, context *cli.Context) ([]requirements.Requirement, error) {
	if len(context.Args()) != 1 {
		cmd.ui.FailWithUsage(context)
	}

	requirements := []requirements.Requirement{requirementsFactory.NewLoginRequirement()}
	return requirements, nil
}

func (cmd CreateAppSecurityGroup) Run(context *cli.Context) {
}
