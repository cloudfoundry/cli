package user

import (
	"cf/api"
	"cf/configuration"
	"cf/models"
	"cf/requirements"
	"cf/terminal"
	"errors"
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

func (cmd *SetSpaceRole) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 4 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "set-space-role")
		return
	}

	cmd.userReq = reqFactory.NewUserRequirement(c.Args()[0])
	cmd.orgReq = reqFactory.NewOrganizationRequirement(c.Args()[1])

	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
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

	space, apiResponse := cmd.spaceRepo.FindByNameInOrg(spaceName, org.Guid)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
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

	apiResponse := cmd.userRepo.SetSpaceRole(userGuid, space.Guid, space.Organization.Guid, role)
	if apiResponse.IsNotSuccessful() {
		err = errors.New(apiResponse.Message)
		return
	}

	cmd.ui.Ok()
	return
}
