package securitygroup

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type DeleteSecurityGroup struct {
	ui                   terminal.UI
	appSecurityGroupRepo api.SecurityGroupRepo
	configRepo           configuration.Reader
}

func NewDeleteAppSecurityGroup(ui terminal.UI, configRepo configuration.Reader, appSecurityGroupRepo api.SecurityGroupRepo) DeleteSecurityGroup {
	return DeleteSecurityGroup{
		ui:                   ui,
		configRepo:           configRepo,
		appSecurityGroupRepo: appSecurityGroupRepo,
	}
}

func (cmd DeleteSecurityGroup) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "delete-security-group",
		Description: "<<< Description goes here >>>",
		Usage:       "CF_NAME delete-security-group NAME",
	}
}

func (cmd DeleteSecurityGroup) GetRequirements(requirementsFactory requirements.Factory, context *cli.Context) ([]requirements.Requirement, error) {
	if len(context.Args()) != 1 {
		cmd.ui.FailWithUsage(context)
	}

	requirements := []requirements.Requirement{requirementsFactory.NewLoginRequirement()}
	return requirements, nil
}

func (cmd DeleteSecurityGroup) Run(context *cli.Context) {
	name := context.Args()[0]

	group, err := cmd.appSecurityGroupRepo.Read(name)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	cmd.ui.Say("Deleting application security group '%s' as '%s'", name, cmd.configRepo.Username())

	err = cmd.appSecurityGroupRepo.Delete(group.Guid)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	cmd.ui.Ok()
}
