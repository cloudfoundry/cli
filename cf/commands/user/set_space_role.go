package user

import (
	"fmt"

	"code.cloudfoundry.org/cli/cf/flags"
	. "code.cloudfoundry.org/cli/cf/i18n"

	"code.cloudfoundry.org/cli/cf"
	"code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/api/featureflags"
	"code.cloudfoundry.org/cli/cf/api/spaces"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
)

//go:generate counterfeiter . SpaceRoleSetter

type SpaceRoleSetter interface {
	commandregistry.Command
	SetSpaceRole(space models.Space, orgGUID, orgName string, role models.Role, userGUID, userName string) (err error)
}

type SetSpaceRole struct {
	ui        terminal.UI
	config    coreconfig.Reader
	spaceRepo spaces.SpaceRepository
	flagRepo  featureflags.FeatureFlagRepository
	userRepo  api.UserRepository
	userReq   requirements.UserRequirement
	orgReq    requirements.OrganizationRequirement
}

func init() {
	commandregistry.Register(&SetSpaceRole{})
}

func (cmd *SetSpaceRole) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "set-space-role",
		Description: T("Assign a space role to a user"),
		Usage: []string{
			T("CF_NAME set-space-role USERNAME ORG SPACE ROLE\n\n"),
			T("ROLES:\n"),
			fmt.Sprintf("   'SpaceManager' - %s", T("Invite and manage users, and enable features for a given space\n")),
			fmt.Sprintf("   'SpaceDeveloper' - %s", T("Create and manage apps and services, and see logs and reports\n")),
			fmt.Sprintf("   'SpaceAuditor' - %s", T("View logs, reports, and settings on this space\n")),
		},
	}
}

func (cmd *SetSpaceRole) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 4 {
		cmd.ui.Failed(T("Incorrect Usage. Requires USERNAME, ORG, SPACE, ROLE as arguments\n\n") + commandregistry.Commands.CommandUsage("set-space-role"))
		return nil, fmt.Errorf("Incorrect usage: %d arguments of %d required", len(fc.Args()), 4)
	}

	var wantGUID bool
	if cmd.config.IsMinAPIVersion(cf.SetRolesByUsernameMinimumAPIVersion) {
		setRolesByUsernameFlag, err := cmd.flagRepo.FindByName("set_roles_by_username")
		wantGUID = (err != nil || !setRolesByUsernameFlag.Enabled)
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

func (cmd *SetSpaceRole) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.spaceRepo = deps.RepoLocator.GetSpaceRepository()
	cmd.userRepo = deps.RepoLocator.GetUserRepository()
	cmd.flagRepo = deps.RepoLocator.GetFeatureFlagRepository()
	return cmd
}

func (cmd *SetSpaceRole) Execute(c flags.FlagContext) error {
	spaceName := c.Args()[2]
	roleStr := c.Args()[3]
	role, err := models.RoleFromString(roleStr)
	if err != nil {
		return err
	}

	userFields := cmd.userReq.GetUser()
	org := cmd.orgReq.GetOrganization()

	space, err := cmd.spaceRepo.FindByNameInOrg(spaceName, org.GUID)
	if err != nil {
		return err
	}

	err = cmd.SetSpaceRole(space, org.GUID, org.Name, role, userFields.GUID, userFields.Username)
	if err != nil {
		return err
	}
	return nil
}

func (cmd *SetSpaceRole) SetSpaceRole(space models.Space, orgGUID, orgName string, role models.Role, userGUID, username string) error {
	var err error

	cmd.ui.Say(T("Assigning role {{.Role}} to user {{.TargetUser}} in org {{.TargetOrg}} / space {{.TargetSpace}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"Role":        terminal.EntityNameColor(role.ToString()),
			"TargetUser":  terminal.EntityNameColor(username),
			"TargetOrg":   terminal.EntityNameColor(orgName),
			"TargetSpace": terminal.EntityNameColor(space.Name),
			"CurrentUser": terminal.EntityNameColor(cmd.config.Username()),
		}))

	if len(userGUID) > 0 {
		err = cmd.userRepo.SetSpaceRoleByGUID(userGUID, space.GUID, orgGUID, role)
	} else {
		err = cmd.userRepo.SetSpaceRoleByUsername(username, space.GUID, orgGUID, role)
	}
	if err != nil {
		return err
	}

	cmd.ui.Ok()
	return nil
}
