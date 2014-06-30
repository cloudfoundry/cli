package securitygroup

import (
	"github.com/cloudfoundry/cli/cf/api/security_groups"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/flag_helpers"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type AssignSecurityGroup struct {
	ui                terminal.UI
	securityGroupRepo security_groups.SecurityGroupRepo
}

func NewAssignSecurityGroup(ui terminal.UI, securityGroupRepo security_groups.SecurityGroupRepo) AssignSecurityGroup {
	return AssignSecurityGroup{
		ui:                ui,
		securityGroupRepo: securityGroupRepo,
	}
}

func (cmd AssignSecurityGroup) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "assign-security-group",
		Description: "Assign a security group to one or more spaces in one or more orgs",
		Usage:       "CF_NAME assign-security-group", // TODO: fix this
		Flags: []cli.Flag{
			flag_helpers.NewStringFlag("space", "Name of a single space to assign the group to"),
			flag_helpers.NewStringFlag("org", "Name of a single org for the --space flag"),
		},
	}
}

func (cmd AssignSecurityGroup) GetRequirements(requirementsFactory requirements.Factory, context *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(context.Args()) != 1 {
		cmd.ui.FailWithUsage(context)
	}

	spaceName := context.String("space")
	orgName := context.String("org")
	if orgName == "" || spaceName == "" {
		cmd.ui.FailWithUsage(context)
	}

	reqs = append(reqs, requirementsFactory.NewLoginRequirement())
	return
}

func (cmd AssignSecurityGroup) Run(context *cli.Context) {
	securityGroupName := context.Args()[0]
	_, err := cmd.securityGroupRepo.Read(securityGroupName)

	if err != nil {
		cmd.ui.Failed(err.Error())
	}
}
