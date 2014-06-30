package securitygroup

import (
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type AssignSecurityGroup struct {
	ui terminal.UI
}

func NewAssignSecurityGroup(ui terminal.UI) AssignSecurityGroup {
	return AssignSecurityGroup{ui: ui}
}

func (cmd AssignSecurityGroup) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "assign-security-group",
		Description: "Assign a security group to one or more spaces in one or more orgs",
		Usage:       "CF_NAME assign-security-group", // TODO: fix this
	}
}

func (cmd AssignSecurityGroup) GetRequirements(requirementsFactory requirements.Factory, context *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(context.Args()) != 1 {
		cmd.ui.FailWithUsage(context)
	}

	reqs = append(reqs, requirementsFactory.NewLoginRequirement())
	return
}

func (cmd AssignSecurityGroup) Run(context *cli.Context) {
}
