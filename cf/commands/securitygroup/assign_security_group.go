package securitygroup

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/api/security_groups"
	"github.com/cloudfoundry/cli/cf/api/security_groups/spaces"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type AssignSecurityGroup struct {
	ui                terminal.UI
	configRepo        configuration.Reader
	orgRepo           api.OrganizationRepository
	spaceRepo         api.SpaceRepository
	securityGroupRepo security_groups.SecurityGroupRepo
	spaceBinder       spaces.SecurityGroupSpaceBinder
}

func NewAssignSecurityGroup(
	ui terminal.UI,
	configRepo configuration.Reader,
	securityGroupRepo security_groups.SecurityGroupRepo,
	spaceRepo api.SpaceRepository,
	orgRepo api.OrganizationRepository,
	spaceBinder spaces.SecurityGroupSpaceBinder,
) AssignSecurityGroup {
	return AssignSecurityGroup{
		ui:                ui,
		configRepo:        configRepo,
		spaceRepo:         spaceRepo,
		orgRepo:           orgRepo,
		securityGroupRepo: securityGroupRepo,
		spaceBinder:       spaceBinder,
	}
}

func (cmd AssignSecurityGroup) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "assign-security-group",
		Description: "Assign a security group to one or more spaces in one or more orgs",
		Usage:       "CF_NAME assign-security-group", // TODO: fix this
	}
}

func (cmd AssignSecurityGroup) GetRequirements(requirementsFactory requirements.Factory, context *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(context.Args()) != 3 {
		cmd.ui.FailWithUsage(context)
	}

	reqs = append(reqs, requirementsFactory.NewLoginRequirement())
	return
}

func (cmd AssignSecurityGroup) Run(context *cli.Context) {
	securityGroupName := context.Args()[0]
	securityGroup, err := cmd.securityGroupRepo.Read(securityGroupName)

	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	orgName := context.Args()[1]
	org, err := cmd.orgRepo.FindByName(orgName)

	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	spaceName := context.Args()[2]
	space, err := cmd.spaceRepo.FindByNameInOrg(spaceName, org.Guid)

	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	err = cmd.spaceBinder.BindSpace(securityGroup.Guid, space.Guid)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}
}
