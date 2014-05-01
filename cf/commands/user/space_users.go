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

var spaceRoles = []string{models.SPACE_MANAGER, models.SPACE_DEVELOPER, models.SPACE_AUDITOR}

var spaceRoleToDisplayName = map[string]string{
	models.SPACE_MANAGER:   "SPACE MANAGER",
	models.SPACE_DEVELOPER: "SPACE DEVELOPER",
	models.SPACE_AUDITOR:   "SPACE AUDITOR",
}

type SpaceUsers struct {
	ui        terminal.UI
	config    configuration.Reader
	spaceRepo api.SpaceRepository
	userRepo  api.UserRepository
	orgReq    requirements.OrganizationRequirement
}

func NewSpaceUsers(ui terminal.UI, config configuration.Reader, spaceRepo api.SpaceRepository, userRepo api.UserRepository) (cmd *SpaceUsers) {
	cmd = new(SpaceUsers)
	cmd.ui = ui
	cmd.config = config
	cmd.spaceRepo = spaceRepo
	cmd.userRepo = userRepo
	return
}

func (command *SpaceUsers) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "space-users",
		Description: "Show space users by role",
		Usage:       "CF_NAME space-users ORG SPACE",
	}
}

func (cmd *SpaceUsers) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 2 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "space-users")
		return
	}

	orgName := c.Args()[0]
	cmd.orgReq = requirementsFactory.NewOrganizationRequirement(orgName)
	reqs = append(reqs, requirementsFactory.NewLoginRequirement(), cmd.orgReq)

	return
}

func (cmd *SpaceUsers) Run(c *cli.Context) {
	spaceName := c.Args()[1]
	org := cmd.orgReq.GetOrganization()

	space, apiErr := cmd.spaceRepo.FindByNameInOrg(spaceName, org.Guid)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
	}

	cmd.ui.Say("Getting users in org %s / space %s as %s",
		terminal.EntityNameColor(org.Name),
		terminal.EntityNameColor(space.Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	for _, role := range spaceRoles {
		displayName := spaceRoleToDisplayName[role]

		users, apiErr := cmd.userRepo.ListUsersInSpaceForRole(space.Guid, role)

		cmd.ui.Say("")
		cmd.ui.Say("%s", terminal.HeaderColor(displayName))

		for _, user := range users {
			cmd.ui.Say("  %s", user.Username)
		}

		if apiErr != nil {
			cmd.ui.Failed("Failed fetching space-users for role %s.\n%s", apiErr.Error(), displayName)
			return
		}
	}
}
