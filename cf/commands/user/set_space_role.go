package user

import (
	"errors"
	. "github.com/cloudfoundry/cli/cf/i18n"

	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/api/spaces"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type SpaceRoleSetter interface {
	SetSpaceRole(space models.Space, role, userGuid, userName string) (err error)
}

type SetSpaceRole struct {
	ui        terminal.UI
	config    configuration.Reader
	spaceRepo spaces.SpaceRepository
	userRepo  api.UserRepository
	userReq   requirements.UserRequirement
	orgReq    requirements.OrganizationRequirement
}

func NewSetSpaceRole(ui terminal.UI, config configuration.Reader, spaceRepo spaces.SpaceRepository, userRepo api.UserRepository) (cmd *SetSpaceRole) {
	cmd = new(SetSpaceRole)
	cmd.ui = ui
	cmd.config = config
	cmd.spaceRepo = spaceRepo
	cmd.userRepo = userRepo
	return
}

func (cmd *SetSpaceRole) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "set-space-role",
		Description: T("Assign a space role to a user"),
		Usage: T("CF_NAME set-space-role USERNAME ORG SPACE ROLE\n\n") +
			T("ROLES:\n") +
			T("   SpaceManager - Invite and manage users, and enable features for a given space\n") +
			T("   SpaceDeveloper - Create and manage apps and services, and see logs and reports\n") +
			T("   SpaceAuditor - View logs, reports, and settings on this space\n"),
	}
}

func (cmd *SetSpaceRole) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 4 {
		cmd.ui.FailWithUsage(c)
	}

	cmd.userReq = requirementsFactory.NewUserRequirement(c.Args()[0])
	cmd.orgReq = requirementsFactory.NewOrganizationRequirement(c.Args()[1])

	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		cmd.userReq,
		cmd.orgReq,
	}
	return
}

func (cmd *SetSpaceRole) Run(c *cli.Context) {
	spaceName := c.Args()[2]
	role := models.UserInputToSpaceRole[c.Args()[3]]
	user := cmd.userReq.GetUser()
	org := cmd.orgReq.GetOrganization()

	space, apiErr := cmd.spaceRepo.FindByNameInOrg(spaceName, org.Guid)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	err := cmd.SetSpaceRole(space, role, user.Guid, user.Username)
	if err != nil {
		cmd.ui.Failed(err.Error())
		return
	}
}

func (cmd *SetSpaceRole) SetSpaceRole(space models.Space, role, userGuid, userName string) (err error) {
	cmd.ui.Say(T("Assigning role {{.Role}} to user {{.TargetUser}} in org {{.TargetOrg}} / space {{.TargetSpace}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"Role":        terminal.EntityNameColor(role),
			"TargetUser":  terminal.EntityNameColor(userName),
			"TargetOrg":   terminal.EntityNameColor(space.Organization.Name),
			"TargetSpace": terminal.EntityNameColor(space.Name),
			"CurrentUser": terminal.EntityNameColor(cmd.config.Username()),
		}))

	apiErr := cmd.userRepo.SetSpaceRole(userGuid, space.Guid, space.Organization.Guid, role)
	if apiErr != nil {
		err = errors.New(apiErr.Error())
		return
	}

	cmd.ui.Ok()
	return
}
