package securitygroup

import (
	"github.com/cloudfoundry/cli/cf/api/organizations"
	"github.com/cloudfoundry/cli/cf/api/securitygroups"
	sgbinder "github.com/cloudfoundry/cli/cf/api/securitygroups/spaces"
	"github.com/cloudfoundry/cli/cf/api/spaces"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
)

type BindSecurityGroup struct {
	ui                terminal.UI
	configRepo        coreconfig.Reader
	orgRepo           organizations.OrganizationRepository
	spaceRepo         spaces.SpaceRepository
	securityGroupRepo security_groups.SecurityGroupRepo
	spaceBinder       sgbinder.SecurityGroupSpaceBinder
}

func init() {
	commandregistry.Register(&BindSecurityGroup{})
}

func (cmd *BindSecurityGroup) MetaData() commandregistry.CommandMetadata {
	primaryUsage := T("CF_NAME bind-security-group SECURITY_GROUP ORG SPACE")
	tipUsage := T("TIP: Changes will not apply to existing running applications until they are restarted.")
	return commandregistry.CommandMetadata{
		Name:        "bind-security-group",
		Description: T("Bind a security group to a space"),
		Usage: []string{
			primaryUsage,
			"\n\n",
			tipUsage,
		},
	}
}

func (cmd *BindSecurityGroup) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	if len(fc.Args()) != 3 {
		cmd.ui.Failed(T("Incorrect Usage. Requires SECURITY_GROUP, ORG and SPACE as arguments\n\n") + commandregistry.Commands.CommandUsage("bind-security-group"))
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

func (cmd *BindSecurityGroup) Execute(context flags.FlagContext) {
	securityGroupName := context.Args()[0]
	orgName := context.Args()[1]
	spaceName := context.Args()[2]

	cmd.ui.Say(T("Assigning security group {{.security_group}} to space {{.space}} in org {{.organization}} as {{.username}}...",
		map[string]interface{}{
			"security_group": securityGroupName,
			"space":          spaceName,
			"organization":   orgName,
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

	space, err := cmd.spaceRepo.FindByNameInOrg(spaceName, org.GUID)

	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	err = cmd.spaceBinder.BindSpace(securityGroup.GUID, space.GUID)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	cmd.ui.Ok()
	cmd.ui.Say("\n\n")
	cmd.ui.Say(T("TIP: Changes will not apply to existing running applications until they are restarted."))
}
