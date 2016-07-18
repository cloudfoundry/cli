package securitygroup

import (
	"github.com/cloudfoundry/cli/cf/api/organizations"
	"github.com/cloudfoundry/cli/cf/api/securitygroups"
	sgbinder "github.com/cloudfoundry/cli/cf/api/securitygroups/spaces"
	"github.com/cloudfoundry/cli/cf/api/spaces"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/flags"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type UnbindSecurityGroup struct {
	ui                terminal.UI
	configRepo        coreconfig.Reader
	securityGroupRepo securitygroups.SecurityGroupRepo
	orgRepo           organizations.OrganizationRepository
	spaceRepo         spaces.SpaceRepository
	secBinder         sgbinder.SecurityGroupSpaceBinder
}

func init() {
	commandregistry.Register(&UnbindSecurityGroup{})
}

func (cmd *UnbindSecurityGroup) MetaData() commandregistry.CommandMetadata {
	primaryUsage := T("CF_NAME unbind-security-group SECURITY_GROUP ORG SPACE")
	tipUsage := T("TIP: Changes will not apply to existing running applications until they are restarted.")
	return commandregistry.CommandMetadata{
		Name:        "unbind-security-group",
		Description: T("Unbind a security group from a space"),
		Usage: []string{
			primaryUsage,
			"\n\n",
			tipUsage,
		},
	}
}

func (cmd *UnbindSecurityGroup) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	argLength := len(fc.Args())
	if argLength == 0 || argLength == 2 || argLength >= 4 {
		cmd.ui.Failed(T("Incorrect Usage. Requires SECURITY_GROUP, ORG and SPACE as arguments\n\n") + commandregistry.Commands.CommandUsage("unbind-security-group"))
	}

	reqs := []requirements.Requirement{requirementsFactory.NewLoginRequirement()}
	return reqs
}

func (cmd *UnbindSecurityGroup) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.configRepo = deps.Config
	cmd.securityGroupRepo = deps.RepoLocator.GetSecurityGroupRepository()
	cmd.spaceRepo = deps.RepoLocator.GetSpaceRepository()
	cmd.orgRepo = deps.RepoLocator.GetOrganizationRepository()
	cmd.secBinder = deps.RepoLocator.GetSecurityGroupSpaceBinder()
	return cmd
}

func (cmd *UnbindSecurityGroup) Execute(context flags.FlagContext) error {
	var spaceGUID string
	var err error

	secName := context.Args()[0]

	if len(context.Args()) == 1 {
		spaceGUID = cmd.configRepo.SpaceFields().GUID
		spaceName := cmd.configRepo.SpaceFields().Name
		orgName := cmd.configRepo.OrganizationFields().Name

		cmd.flavorText(secName, orgName, spaceName)
	} else {
		orgName := context.Args()[1]
		spaceName := context.Args()[2]

		cmd.flavorText(secName, orgName, spaceName)

		spaceGUID, err = cmd.lookupSpaceGUID(orgName, spaceName)
		if err != nil {
			return err
		}
	}

	securityGroup, err := cmd.securityGroupRepo.Read(secName)
	if err != nil {
		return err
	}

	secGUID := securityGroup.GUID

	err = cmd.secBinder.UnbindSpace(secGUID, spaceGUID)
	if err != nil {
		return err
	}
	cmd.ui.Ok()
	cmd.ui.Say("\n\n")
	cmd.ui.Say(T("TIP: Changes will not apply to existing running applications until they are restarted."))
	return nil
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

func (cmd UnbindSecurityGroup) lookupSpaceGUID(orgName string, spaceName string) (string, error) {
	organization, err := cmd.orgRepo.FindByName(orgName)
	if err != nil {
		return "", err
	}
	orgGUID := organization.GUID

	space, err := cmd.spaceRepo.FindByNameInOrg(spaceName, orgGUID)
	if err != nil {
		return "", err
	}
	return space.GUID, nil
}
