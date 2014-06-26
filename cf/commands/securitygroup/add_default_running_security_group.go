package securitygroup

import (
	"github.com/cloudfoundry/cli/cf/api/security_groups"
	"github.com/cloudfoundry/cli/cf/api/security_groups/defaults/running"
	"github.com/cloudfoundry/cli/cf/command"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type addToDefaultRunningGroup struct {
	ui                terminal.UI
	configRepo        configuration.Reader
	securityGroupRepo security_groups.SecurityGroupRepo
	runningGroupRepo  running.RunningSecurityGroupsRepo
}

func NewAddToDefaultRunningGroup(ui terminal.UI, configRepo configuration.Reader, securityGroupRepo security_groups.SecurityGroupRepo, runningGroupRepo running.RunningSecurityGroupsRepo) command.Command {
	return &addToDefaultRunningGroup{
		ui:                ui,
		configRepo:        configRepo,
		securityGroupRepo: securityGroupRepo,
		runningGroupRepo:  runningGroupRepo,
	}
}

func (cmd *addToDefaultRunningGroup) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "add-default-running-security-group",
		Description: "Add a security group to the list of security groups to be used for running Applications",
		Usage:       "CF_NAME add-default-running-security-group NAME",
	}
}

func (cmd *addToDefaultRunningGroup) GetRequirements(requirementsFactory requirements.Factory, context *cli.Context) ([]requirements.Requirement, error) {
	if len(context.Args()) != 1 {
		cmd.ui.FailWithUsage(context)
	}

	return []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}, nil
}

func (cmd *addToDefaultRunningGroup) Run(context *cli.Context) {
	name := context.Args()[0]

	securityGroup, err := cmd.securityGroupRepo.Read(name)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	cmd.ui.Say("Adding security group '%s' to defaults for running as '%s'", securityGroup.Name, cmd.configRepo.Username())
	err = cmd.runningGroupRepo.AddToDefaultRunningSet(securityGroup.Guid)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	cmd.ui.Ok()
}
