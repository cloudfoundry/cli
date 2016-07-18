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
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type BindSecurityGroup struct {
	ui                terminal.UI
	configRepo        coreconfig.Reader
	orgRepo           organizations.OrganizationRepository
	spaceRepo         spaces.SpaceRepository
	securityGroupRepo securitygroups.SecurityGroupRepo
	spaceBinder       sgbinder.SecurityGroupSpaceBinder
}

func init() {
	commandregistry.Register(&BindSecurityGroup{})
}

func (cmd *BindSecurityGroup) MetaData() commandregistry.CommandMetadata {
	primaryUsage := T("CF_NAME bind-security-group SECURITY_GROUP ORG [SPACE]")
	tipUsage := T("TIP: Changes will not apply to existing running applications until they are restarted.")
	return commandregistry.CommandMetadata{
		Name:        "bind-security-group",
		Description: T("Bind a security group to a particular space, or all existing spaces of an org"),
		Usage: []string{
			primaryUsage,
			"\n\n",
			tipUsage,
		},
	}
}

func (cmd *BindSecurityGroup) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	if len(fc.Args()) != 2 && len(fc.Args()) != 3 {
		cmd.ui.Failed(T("Incorrect Usage. Requires SECURITY_GROUP and ORG, optional SPACE as arguments\n\n") + commandregistry.Commands.CommandUsage("bind-security-group"))
	}

	reqs := []requirements.Requirement{}
	reqs = append(reqs, requirementsFactory.NewLoginRequirement())
	return reqs
}

func (cmd *BindSecurityGroup) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.configRepo = deps.Config
	cmd.spaceRepo = deps.RepoLocator.GetSpaceRepository()
	cmd.orgRepo = deps.RepoLocator.GetOrganizationRepository()
	cmd.securityGroupRepo = deps.RepoLocator.GetSecurityGroupRepository()
	cmd.spaceBinder = deps.RepoLocator.GetSecurityGroupSpaceBinder()
	return cmd
}

func (cmd *BindSecurityGroup) Execute(context flags.FlagContext) error {
	securityGroupName := context.Args()[0]
	orgName := context.Args()[1]

	securityGroup, err := cmd.securityGroupRepo.Read(securityGroupName)
	if err != nil {
		return err
	}

	org, err := cmd.orgRepo.FindByName(orgName)
	if err != nil {
		return err
	}

	spaces := []models.Space{}
	if len(context.Args()) > 2 {
		var space models.Space
		space, err = cmd.spaceRepo.FindByNameInOrg(context.Args()[2], org.GUID)
		if err != nil {
			return err
		}

		spaces = append(spaces, space)
	} else {
		err = cmd.spaceRepo.ListSpacesFromOrg(org.GUID, func(space models.Space) bool {
			spaces = append(spaces, space)
			return true
		})
		if err != nil {
			return err
		}
	}

	for _, space := range spaces {
		cmd.ui.Say(T("Assigning security group {{.security_group}} to space {{.space}} in org {{.organization}} as {{.username}}...",
			map[string]interface{}{
				"security_group": terminal.EntityNameColor(securityGroupName),
				"space":          terminal.EntityNameColor(space.Name),
				"organization":   terminal.EntityNameColor(orgName),
				"username":       terminal.EntityNameColor(cmd.configRepo.Username()),
			}))
		err = cmd.spaceBinder.BindSpace(securityGroup.GUID, space.GUID)
		if err != nil {
			return err
		}
		cmd.ui.Ok()
		cmd.ui.Say("")
	}

	cmd.ui.Say(T("TIP: Changes will not apply to existing running applications until they are restarted."))
	return nil
}
