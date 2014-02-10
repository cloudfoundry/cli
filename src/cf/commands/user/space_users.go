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

var spaceRoles = []string{models.SPACE_MANAGER, models.SPACE_DEVELOPER, models.SPACE_AUDITOR}

var spaceRoleToDisplayName = map[string]string{
	models.SPACE_MANAGER:   "SPACE MANAGER",
	models.SPACE_DEVELOPER: "SPACE DEVELOPER",
	models.SPACE_AUDITOR:   "SPACE AUDITOR",
}

type SpaceUsers struct {
	ui        terminal.UI
	config    *configuration.Configuration
	spaceRepo api.SpaceRepository
	userRepo  api.UserRepository
	orgReq    requirements.OrganizationRequirement
}

func NewSpaceUsers(ui terminal.UI, config *configuration.Configuration, spaceRepo api.SpaceRepository, userRepo api.UserRepository) (cmd *SpaceUsers) {
	cmd = new(SpaceUsers)
	cmd.ui = ui
	cmd.config = config
	cmd.spaceRepo = spaceRepo
	cmd.userRepo = userRepo
	return
}

func (cmd *SpaceUsers) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 2 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "space-users")
		return
	}

	orgName := c.Args()[0]
	cmd.orgReq = reqFactory.NewOrganizationRequirement(orgName)
	reqs = append(reqs, reqFactory.NewLoginRequirement(), cmd.orgReq)

	return
}

func (cmd *SpaceUsers) Run(c *cli.Context) {
	spaceName := c.Args()[1]
	org := cmd.orgReq.GetOrganization()

	space, apiResponse := cmd.spaceRepo.FindByNameInOrg(spaceName, org.Guid)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
	}

	cmd.ui.Say("Getting users in org %s / space %s as %s",
		terminal.EntityNameColor(org.Name),
		terminal.EntityNameColor(space.Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	for _, role := range spaceRoles {
		displayName := spaceRoleToDisplayName[role]

		users, apiResponse := cmd.userRepo.ListUsersInSpaceForRole(space.Guid, role)

		cmd.ui.Say("")
		cmd.ui.Say("%s", terminal.HeaderColor(displayName))

		for _, user := range users {
			cmd.ui.Say("  %s", user.Username)
		}

		if apiResponse.IsNotSuccessful() {
			cmd.ui.Failed("Failed fetching space-users for role %s.\n%s", apiResponse.Message, displayName)
			return
		}
	}
}
