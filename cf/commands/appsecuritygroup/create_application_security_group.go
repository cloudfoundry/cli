package appsecuritygroup

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type CreateAppSecurityGroup struct {
	ui   terminal.UI
	repo api.AppSecurityGroup
}

func NewCreateAppSecurityGroup(ui terminal.UI, repo api.AppSecurityGroup) CreateAppSecurityGroup {
	return CreateAppSecurityGroup{
		ui:   ui,
		repo: repo,
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
	name := context.Args()[0]
	cmd.repo.Create(name)
}
