package appsecuritygroup

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
	"github.com/cloudfoundry/cli/cf/flag_helpers"
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
		Flags: []cli.Flag{
			flag_helpers.NewStringFlag("rules", "THIS IS ALL THE RULES!"),
		},
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
	rules := context.String("rules")

	cmd.ui.Say("Creating application security group %s", name)

	err := cmd.repo.Create(api.ApplicationSecurityGroupFields{Name: name, Rules: rules})
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	cmd.ui.Ok()
}
