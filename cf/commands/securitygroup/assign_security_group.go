package securitygroup

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/api/security_groups"
	sgbinder "github.com/cloudfoundry/cli/cf/api/security_groups/spaces"
	"github.com/cloudfoundry/cli/cf/api/spaces"
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
	spaceRepo         spaces.SpaceRepository
	securityGroupRepo security_groups.SecurityGroupRepo
	spaceBinder       sgbinder.SecurityGroupSpaceBinder
}

func NewAssignSecurityGroup(
	ui terminal.UI,
	configRepo configuration.Reader,
	securityGroupRepo security_groups.SecurityGroupRepo,
	spaceRepo spaces.SpaceRepository,
	orgRepo api.OrganizationRepository,
	spaceBinder sgbinder.SecurityGroupSpaceBinder,
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
		Description: T("Assign a security group to a space"),
		Usage:       T("CF_NAME assign-security-group SECURITY_GROUP ORG SPACE"),
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
	orgName := context.Args()[1]
	spaceName := context.Args()[2]

	cmd.ui.Say(T("Assigning security group {{.security_group}} to space {{.space}} in org {{.organization}} as {{.username}}...",
		map[string]interface{}{
			"security_group": securityGroupName,
			"space":          orgName,
			"organization":   spaceName,
			"username":       cmd.configRepo.Username(),
		}))

	securityGroup, err := cmd.securityGroupRepo.Read(securityGroupName)

	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	org, err := cmd.orgRepo.FindByName(orgName)

	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	space, err := cmd.spaceRepo.FindByNameInOrg(spaceName, org.Guid)

	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	err = cmd.spaceBinder.BindSpace(securityGroup.Guid, space.Guid)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	cmd.ui.Ok()
}
