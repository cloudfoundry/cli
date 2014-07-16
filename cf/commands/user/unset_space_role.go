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

type UnsetSpaceRole struct {
	ui        terminal.UI
	config    configuration.Reader
	spaceRepo spaces.SpaceRepository
	userRepo  api.UserRepository
	userReq   requirements.UserRequirement
	orgReq    requirements.OrganizationRequirement
}

func NewUnsetSpaceRole(ui terminal.UI, config configuration.Reader, spaceRepo spaces.SpaceRepository, userRepo api.UserRepository) (cmd *UnsetSpaceRole) {
	cmd = new(UnsetSpaceRole)
	cmd.ui = ui
	cmd.config = config
	cmd.spaceRepo = spaceRepo
	cmd.userRepo = userRepo
	return
}

func (cmd *UnsetSpaceRole) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "unset-space-role",
		Description: T("Remove a space role from a user"),
		Usage: T("CF_NAME unset-space-role USERNAME ORG SPACE ROLE\n\n") +
			T("ROLES:\n") +
			T("   SpaceManager - Invite and manage users, and enable features for a given space\n") +
			T("   SpaceDeveloper - Create and manage apps and services, and see logs and reports\n") +
			T("   SpaceAuditor - View logs, reports, and settings on this space\n"),
	}
}

func (cmd *UnsetSpaceRole) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
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

func (cmd *UnsetSpaceRole) Run(c *cli.Context) {
	spaceName := c.Args()[2]
	role := models.UserInputToSpaceRole[c.Args()[3]]

	user := cmd.userReq.GetUser()
	org := cmd.orgReq.GetOrganization()
	space, apiErr := cmd.spaceRepo.FindByNameInOrg(spaceName, org.Guid)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Say(T("Removing role {{.Role}} from user {{.TargetUser}} in org {{.TargetOrg}} / space {{.TargetSpace}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"Role":        terminal.EntityNameColor(role),
			"TargetUser":  terminal.EntityNameColor(user.Username),
			"TargetOrg":   terminal.EntityNameColor(org.Name),
			"TargetSpace": terminal.EntityNameColor(space.Name),
			"CurrentUser": terminal.EntityNameColor(cmd.config.Username()),
		}))

	apiErr = cmd.userRepo.UnsetSpaceRole(user.Guid, space.Guid, role)

	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Ok()
}
