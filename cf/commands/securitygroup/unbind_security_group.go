package securitygroup

import (
	"strings"

	"github.com/cloudfoundry/cli/cf/api/organizations"
	"github.com/cloudfoundry/cli/cf/api/security_groups"
	sgbinder "github.com/cloudfoundry/cli/cf/api/security_groups/spaces"
	"github.com/cloudfoundry/cli/cf/api/spaces"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
)

type UnbindSecurityGroup struct {
	ui                terminal.UI
	configRepo        core_config.Reader
	securityGroupRepo security_groups.SecurityGroupRepo
	orgRepo           organizations.OrganizationRepository
	spaceRepo         spaces.SpaceRepository
	secBinder         sgbinder.SecurityGroupSpaceBinder
}

func init() {
	command_registry.Register(&UnbindSecurityGroup{})
}

func (cmd *UnbindSecurityGroup) MetaData() command_registry.CommandMetadata {
	primaryUsage := T("CF_NAME unbind-security-group SECURITY_GROUP ORG SPACE")
	tipUsage := T("TIP: Changes will not apply to existing running applications until they are restarted.")
	return command_registry.CommandMetadata{
		Name:        "unbind-security-group",
		Description: T("Unbind a security group from a space"),
		Usage:       strings.Join([]string{primaryUsage, tipUsage}, "\n\n"),
	}
}

func (cmd *UnbindSecurityGroup) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	argLength := len(fc.Args())
	if argLength == 0 || argLength == 2 || argLength >= 4 {
		cmd.ui.Failed(T("Incorrect Usage. Requires SECURITY_GROUP, ORG and SPACE as arguments\n\n") + command_registry.Commands.CommandUsage("unbind-security-group"))
	}

	requirements := []requirements.Requirement{requirementsFactory.NewLoginRequirement()}
	return requirements, nil
}

func (cmd *UnbindSecurityGroup) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.configRepo = deps.Config
	cmd.securityGroupRepo = deps.RepoLocator.GetSecurityGroupRepository()
	cmd.spaceRepo = deps.RepoLocator.GetSpaceRepository()
	cmd.orgRepo = deps.RepoLocator.GetOrganizationRepository()
	cmd.secBinder = deps.RepoLocator.GetSecurityGroupSpaceBinder()
	return cmd
}

func (cmd *UnbindSecurityGroup) Execute(context flags.FlagContext) {
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
	cmd.ui.Say("\n\n")
	cmd.ui.Say(T("TIP: Changes will not apply to existing running applications until they are restarted."))
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
