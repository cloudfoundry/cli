package space

import (
	"cf"
	"cf/api"
	"cf/commands/user"
	"cf/configuration"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type CreateSpace struct {
	ui              terminal.UI
	config          *configuration.Configuration
	spaceRepo       api.SpaceRepository
	userRepo        api.UserRepository
	spaceRoleSetter user.SpaceRoleSetter
}

func NewCreateSpace(ui terminal.UI, config *configuration.Configuration, spaceRoleSetter user.SpaceRoleSetter, spaceRepo api.SpaceRepository, userRepo api.UserRepository) (cmd CreateSpace) {
	cmd.ui = ui
	cmd.config = config
	cmd.spaceRoleSetter = spaceRoleSetter
	cmd.spaceRepo = spaceRepo
	cmd.userRepo = userRepo
	return
}

func (cmd CreateSpace) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) == 0 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "create-space")
		return
	}

	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
		reqFactory.NewTargetedOrgRequirement(),
	}
	return
}

func (cmd CreateSpace) Run(c *cli.Context) {
	spaceName := c.Args()[0]
	cmd.ui.Say("Creating space %s in org %s as %s...",
		terminal.EntityNameColor(spaceName),
		terminal.EntityNameColor(cmd.config.OrganizationFields.Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	space, apiResponse := cmd.spaceRepo.Create(spaceName, cmd.config.OrganizationFields.Guid)
	if apiResponse.IsNotSuccessful() {
		if apiResponse.ErrorCode == cf.SPACE_EXISTS {
			cmd.ui.Ok()
			cmd.ui.Warn("Space %s already exists", spaceName)
			return
		}
		cmd.ui.Failed(apiResponse.Message)
		return
	}
	cmd.ui.Ok()

	var err error

	err = cmd.spaceRoleSetter.SetSpaceRole(space, cf.SPACE_MANAGER, cmd.config.UserGuid(), cmd.config.Username())
	if err != nil {
		cmd.ui.Failed(err.Error())
		return
	}

	err = cmd.spaceRoleSetter.SetSpaceRole(space, cf.SPACE_DEVELOPER, cmd.config.UserGuid(), cmd.config.Username())
	if err != nil {
		cmd.ui.Failed(err.Error())
		return
	}

	cmd.ui.Say("\nTIP: Use '%s' to target new space", terminal.CommandColor(cf.Name()+" target -s "+space.Name))
}
