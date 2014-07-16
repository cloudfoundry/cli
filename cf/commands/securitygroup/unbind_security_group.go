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

type UnbindSecurityGroup struct {
	ui                terminal.UI
	configRepo        configuration.Reader
	securityGroupRepo security_groups.SecurityGroupRepo
	orgRepo           api.OrganizationRepository
	spaceRepo         spaces.SpaceRepository
	secBinder         sgbinder.SecurityGroupSpaceBinder
}

func NewUnbindSecurityGroup(ui terminal.UI, configRepo configuration.Reader, securityGroupRepo security_groups.SecurityGroupRepo, orgRepo api.OrganizationRepository, spaceRepo spaces.SpaceRepository, secBinder sgbinder.SecurityGroupSpaceBinder) UnbindSecurityGroup {
	return UnbindSecurityGroup{
		ui:                ui,
		configRepo:        configRepo,
		securityGroupRepo: securityGroupRepo,
		orgRepo:           orgRepo,
		spaceRepo:         spaceRepo,
		secBinder:         secBinder,
	}
}

func (cmd UnbindSecurityGroup) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "unbind-security-group",
		Description: T("Unbind a security group from a space"),
		Usage:       T("CF_NAME unbind-security-group SECURITY_GROUP ORG SPACE"),
	}
}

func (cmd UnbindSecurityGroup) GetRequirements(requirementsFactory requirements.Factory, context *cli.Context) ([]requirements.Requirement, error) {
	argLength := len(context.Args())
	if argLength == 0 || argLength == 2 || argLength >= 4 {
		cmd.ui.FailWithUsage(context)
	}

	requirements := []requirements.Requirement{requirementsFactory.NewLoginRequirement()}
	return requirements, nil
}

func (cmd UnbindSecurityGroup) Run(context *cli.Context) {
	var spaceGuid string
	secName := context.Args()[0]

	if len(context.Args()) == 1 {
		spaceGuid = cmd.configRepo.SpaceFields().Guid
		spaceName := cmd.configRepo.SpaceFields().Name
		orgName := cmd.configRepo.OrganizationFields().Name

		cmd.flavorText(secName, orgName, spaceName)
	} else {
		orgName := context.Args()[1]
		spaceName := context.Args()[2]

		cmd.flavorText(secName, orgName, spaceName)

		spaceGuid = cmd.lookupSpaceGuid(orgName, spaceName)
	}

	securityGroup, err := cmd.securityGroupRepo.Read(secName)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	secGuid := securityGroup.Guid

	err = cmd.secBinder.UnbindSpace(secGuid, spaceGuid)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}
	cmd.ui.Ok()
}

func (cmd UnbindSecurityGroup) flavorText(secName string, orgName string, spaceName string) {
	cmd.ui.Say(T("Unbinding security group {{.security_group}} from {{.organization}}/{{.space}} as {{.username}}",
		map[string]interface{}{
			"security_group": terminal.EntityNameColor(secName),
			"organization":   terminal.EntityNameColor(orgName),
			"space":          terminal.EntityNameColor(spaceName),
			"username":       terminal.EntityNameColor(cmd.configRepo.Username()),
		}))
}

func (cmd UnbindSecurityGroup) lookupSpaceGuid(orgName string, spaceName string) string {
	organization, err := cmd.orgRepo.FindByName(orgName)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}
	orgGuid := organization.Guid

	space, err := cmd.spaceRepo.FindByNameInOrg(spaceName, orgGuid)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}
	return space.Guid
}
