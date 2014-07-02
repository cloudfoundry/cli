package securitygroup

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/api/security_groups"
	sgbinder "github.com/cloudfoundry/cli/cf/api/security_groups/spaces"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type UnassignSecurityGroup struct {
	ui                terminal.UI
	configRepo        configuration.Reader
	securityGroupRepo security_groups.SecurityGroupRepo
	orgRepo           api.OrganizationRepository
	spaceRepo         api.SpaceRepository
	secBinder         sgbinder.SecurityGroupSpaceBinder
}

func NewUnassignSecurityGroup(ui terminal.UI, configRepo configuration.Reader, securityGroupRepo security_groups.SecurityGroupRepo, orgRepo api.OrganizationRepository, spaceRepo api.SpaceRepository, secBinder sgbinder.SecurityGroupSpaceBinder) UnassignSecurityGroup {
	return UnassignSecurityGroup{
		ui:                ui,
		configRepo:        configRepo,
		securityGroupRepo: securityGroupRepo,
		orgRepo:           orgRepo,
		spaceRepo:         spaceRepo,
		secBinder:         secBinder,
	}
}

func (cmd UnassignSecurityGroup) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "unassign-security-group",
		Description: "Unassigns a security group from a given space",
		Usage:       "CF_NAME unassign-security-group SECURITY_GROUP ORG SPACE",
	}
}

func (cmd UnassignSecurityGroup) GetRequirements(requirementsFactory requirements.Factory, context *cli.Context) ([]requirements.Requirement, error) {
	argLength := len(context.Args())
	if argLength == 0 || argLength == 2 || argLength >= 4 {
		cmd.ui.FailWithUsage(context)
	}

	requirements := []requirements.Requirement{requirementsFactory.NewLoginRequirement()}
	return requirements, nil
}

func (cmd UnassignSecurityGroup) Run(context *cli.Context) {
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

func (cmd UnassignSecurityGroup) flavorText(secName string, orgName string, spaceName string) {
	cmd.ui.Say("Removing security group %s from %s/%s as %s",
		terminal.EntityNameColor(secName),
		terminal.EntityNameColor(orgName),
		terminal.EntityNameColor(spaceName),
		terminal.EntityNameColor(cmd.configRepo.Username()))
}

func (cmd UnassignSecurityGroup) lookupSpaceGuid(orgName string, spaceName string) string {
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
