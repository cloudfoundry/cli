package user

import (
	"fmt"

	"code.cloudfoundry.org/cli/cf"
	"code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/api/featureflags"
	"code.cloudfoundry.org/cli/cf/api/spaces"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/flags"
	. "code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
)

type UnsetSpaceRole struct {
	ui        terminal.UI
	config    coreconfig.Reader
	spaceRepo spaces.SpaceRepository
	userRepo  api.UserRepository
	flagRepo  featureflags.FeatureFlagRepository
	userReq   requirements.UserRequirement
	orgReq    requirements.OrganizationRequirement
}

func init() {
	commandregistry.Register(&UnsetSpaceRole{})
}

func (cmd *UnsetSpaceRole) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "unset-space-role",
		Description: T("Remove a space role from a user"),
		Usage: []string{
			T("CF_NAME unset-space-role USERNAME ORG SPACE ROLE\n\n"),
			T("ROLES:\n"),
			fmt.Sprintf("   'SpaceManager' - %s", T("Invite and manage users, and enable features for a given space\n")),
			fmt.Sprintf("   'SpaceDeveloper' - %s", T("Create and manage apps and services, and see logs and reports\n")),
			fmt.Sprintf("   'SpaceAuditor' - %s", T("View logs, reports, and settings on this space\n")),
		},
	}
}

func (cmd *UnsetSpaceRole) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 4 {
		cmd.ui.Failed(T("Incorrect Usage. Requires USERNAME, ORG, SPACE, ROLE as arguments\n\n") + commandregistry.Commands.CommandUsage("unset-space-role"))
		return nil, fmt.Errorf("Incorrect usage: %d arguments of %d required", len(fc.Args()), 4)
	}

	var wantGUID bool
	if cmd.config.IsMinAPIVersion(cf.SetRolesByUsernameMinimumAPIVersion) {
		unsetRolesByUsernameFlag, err := cmd.flagRepo.FindByName("unset_roles_by_username")
		wantGUID = (err != nil || !unsetRolesByUsernameFlag.Enabled)
	} else {
		wantGUID = true
	}

	cmd.userReq = requirementsFactory.NewUserRequirement(fc.Args()[0], wantGUID)
	cmd.orgReq = requirementsFactory.NewOrganizationRequirement(fc.Args()[1])

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		cmd.userReq,
		cmd.orgReq,
	}

	return reqs, nil
}

func (cmd *UnsetSpaceRole) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.spaceRepo = deps.RepoLocator.GetSpaceRepository()
	cmd.userRepo = deps.RepoLocator.GetUserRepository()
	cmd.flagRepo = deps.RepoLocator.GetFeatureFlagRepository()
	return cmd
}

func (cmd *UnsetSpaceRole) Execute(c flags.FlagContext) error {
	spaceName := c.Args()[2]
	roleStr := c.Args()[3]
	role, err := models.RoleFromString(roleStr)
	if err != nil {
		return err
	}
	user := cmd.userReq.GetUser()
	org := cmd.orgReq.GetOrganization()
	space, err := cmd.spaceRepo.FindByNameInOrg(spaceName, org.GUID)
	if err != nil {
		return err
	}

	cmd.ui.Say(T("Removing role {{.Role}} from user {{.TargetUser}} in org {{.TargetOrg}} / space {{.TargetSpace}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"Role":        terminal.EntityNameColor(roleStr),
			"TargetUser":  terminal.EntityNameColor(user.Username),
			"TargetOrg":   terminal.EntityNameColor(org.Name),
			"TargetSpace": terminal.EntityNameColor(space.Name),
			"CurrentUser": terminal.EntityNameColor(cmd.config.Username()),
		}))

	if len(user.GUID) > 0 {
		err = cmd.userRepo.UnsetSpaceRoleByGUID(user.GUID, space.GUID, role)
	} else {
		err = cmd.userRepo.UnsetSpaceRoleByUsername(user.Username, space.GUID, role)
	}
	if err != nil {
		return err
	}

	cmd.ui.Ok()
	return nil
}
