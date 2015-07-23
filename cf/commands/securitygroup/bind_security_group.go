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

type BindSecurityGroup struct {
	ui                terminal.UI
	configRepo        core_config.Reader
	orgRepo           organizations.OrganizationRepository
	spaceRepo         spaces.SpaceRepository
	securityGroupRepo security_groups.SecurityGroupRepo
	spaceBinder       sgbinder.SecurityGroupSpaceBinder
}

func init() {
	command_registry.Register(&BindSecurityGroup{})
}

func (cmd *BindSecurityGroup) MetaData() command_registry.CommandMetadata {
	primaryUsage := T("CF_NAME bind-security-group SECURITY_GROUP ORG SPACE")
	tipUsage := T("TIP: Changes will not apply to existing running applications until they are restarted.")
	return command_registry.CommandMetadata{
		Name:        "bind-security-group",
		Description: T("Bind a security group to a space"),
		Usage:       strings.Join([]string{primaryUsage, tipUsage}, "\n\n"),
	}
}

func (cmd *BindSecurityGroup) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 3 {
		cmd.ui.Failed(T("Incorrect Usage. Requires SECURITY_GROUP, ORG and SPACE as arguments\n\n") + command_registry.Commands.CommandUsage("bind-security-group"))
	}

	reqs := []requirements.Requirement{}
	reqs = append(reqs, requirementsFactory.NewLoginRequirement())
	return reqs, nil
}

func (cmd *BindSecurityGroup) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
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

	space, err := cmd.spaceRepo.FindByNameInOrg(spaceName, org.Guid)

	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	err = cmd.spaceBinder.BindSpace(securityGroup.Guid, space.Guid)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	cmd.ui.Ok()
	cmd.ui.Say("\n\n")
	cmd.ui.Say(T("TIP: Changes will not apply to existing running applications until they are restarted."))
}
