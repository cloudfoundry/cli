package securitygroup

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/api/security_groups"
	sgbinder "github.com/cloudfoundry/cli/cf/api/security_groups/spaces"
	"github.com/cloudfoundry/cli/cf/api/spaces"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type BindSecurityGroup struct {
	ui                terminal.UI
	configRepo        configuration.Reader
	orgRepo           api.OrganizationRepository
	spaceRepo         spaces.SpaceRepository
	securityGroupRepo security_groups.SecurityGroupRepo
	spaceBinder       sgbinder.SecurityGroupSpaceBinder
}

func NewBindSecurityGroup(
	ui terminal.UI,
	configRepo configuration.Reader,
	securityGroupRepo security_groups.SecurityGroupRepo,
	spaceRepo spaces.SpaceRepository,
	orgRepo api.OrganizationRepository,
	spaceBinder sgbinder.SecurityGroupSpaceBinder,
) BindSecurityGroup {
	return BindSecurityGroup{
		ui:                ui,
		configRepo:        configRepo,
		spaceRepo:         spaceRepo,
		orgRepo:           orgRepo,
		securityGroupRepo: securityGroupRepo,
		spaceBinder:       spaceBinder,
	}
}

func (cmd BindSecurityGroup) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "bind-security-group",
		Description: T("Bind a security group to a space"),
		Usage:       T("CF_NAME bind-security-group SECURITY_GROUP ORG SPACE"),
	}
}

func (cmd BindSecurityGroup) GetRequirements(requirementsFactory requirements.Factory, context *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(context.Args()) != 3 {
		cmd.ui.FailWithUsage(context)
	}

	reqs = append(reqs, requirementsFactory.NewLoginRequirement())
	return
}

func (cmd BindSecurityGroup) Run(context *cli.Context) {
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
