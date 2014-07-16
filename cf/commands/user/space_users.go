package user

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/api/spaces"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

var spaceRoles = []string{models.SPACE_MANAGER, models.SPACE_DEVELOPER, models.SPACE_AUDITOR}

type SpaceUsers struct {
	ui        terminal.UI
	config    configuration.Reader
	spaceRepo spaces.SpaceRepository
	userRepo  api.UserRepository
	orgReq    requirements.OrganizationRequirement
}

func NewSpaceUsers(ui terminal.UI, config configuration.Reader, spaceRepo spaces.SpaceRepository, userRepo api.UserRepository) (cmd *SpaceUsers) {
	cmd = new(SpaceUsers)
	cmd.ui = ui
	cmd.config = config
	cmd.spaceRepo = spaceRepo
	cmd.userRepo = userRepo
	return
}

func (cmd *SpaceUsers) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "space-users",
		Description: T("Show space users by role"),
		Usage:       T("CF_NAME space-users ORG SPACE"),
	}
}

func (cmd *SpaceUsers) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 2 {
		cmd.ui.FailWithUsage(c)
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

	cmd.ui.Say(T("Getting users in org {{.TargetOrg}} / space {{.TargetSpace}} as {{.CurrentUser}}",
		map[string]interface{}{
			"TargetOrg":   terminal.EntityNameColor(org.Name),
			"TargetSpace": terminal.EntityNameColor(space.Name),
			"CurrentUser": terminal.EntityNameColor(cmd.config.Username()),
		}))

	var spaceRoleToDisplayName = map[string]string{
		models.SPACE_MANAGER:   T("SPACE MANAGER"),
		models.SPACE_DEVELOPER: T("SPACE DEVELOPER"),
		models.SPACE_AUDITOR:   T("SPACE AUDITOR"),
	}

	for _, role := range spaceRoles {
		displayName := spaceRoleToDisplayName[role]

		users, apiErr := cmd.userRepo.ListUsersInSpaceForRole(space.Guid, role)

		cmd.ui.Say("")
		cmd.ui.Say("%s", terminal.HeaderColor(displayName))

		for _, user := range users {
			cmd.ui.Say("  %s", user.Username)
		}

		if apiErr != nil {
			cmd.ui.Failed(T("Failed fetching space-users for role {{.SpaceRoleToDisplayName}}.\n{{.Error}}",
				map[string]interface{}{
					"Error":                  apiErr.Error(),
					"SpaceRoleToDisplayName": displayName,
				}))
			return
		}
	}
}
