package user

import (
	"errors"
	"github.com/cloudfoundry/cli/cf/api"
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
	spaceRepo api.SpaceRepository
	userRepo  api.UserRepository
	userReq   requirements.UserRequirement
	orgReq    requirements.OrganizationRequirement
}

func NewSetSpaceRole(ui terminal.UI, config configuration.Reader, spaceRepo api.SpaceRepository, userRepo api.UserRepository) (cmd *SetSpaceRole) {
	cmd = new(SetSpaceRole)
	cmd.ui = ui
	cmd.config = config
	cmd.spaceRepo = spaceRepo
	cmd.userRepo = userRepo
	return
}

func (command *SetSpaceRole) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "set-space-role",
		Description: "Assign a space role to a user",
		Usage: "CF_NAME set-space-role USERNAME ORG SPACE ROLE\n\n" +
			"ROLES:\n" +
			"   SpaceManager - Invite and manage users, and enable features for a given space\n" +
			"   SpaceDeveloper - Create and manage apps and services, and see logs and reports\n" +
			"   SpaceAuditor - View logs, reports, and settings on this space\n",
	}
}

func (cmd *SetSpaceRole) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 4 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "set-space-role")
		return
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
	cmd.ui.Say("Assigning role %s to user %s in org %s / space %s as %s...",
		terminal.EntityNameColor(role),
		terminal.EntityNameColor(userName),
		terminal.EntityNameColor(space.Organization.Name),
		terminal.EntityNameColor(space.Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	apiErr := cmd.userRepo.SetSpaceRole(userGuid, space.Guid, space.Organization.Guid, role)
	if apiErr != nil {
		err = errors.New(apiErr.Error())
		return
	}

	cmd.ui.Ok()
	return
}
