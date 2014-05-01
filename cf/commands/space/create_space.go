package space

import (
	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/commands/user"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/flag_helpers"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type CreateSpace struct {
	ui              terminal.UI
	config          configuration.Reader
	spaceRepo       api.SpaceRepository
	orgRepo         api.OrganizationRepository
	userRepo        api.UserRepository
	spaceRoleSetter user.SpaceRoleSetter
}

func NewCreateSpace(ui terminal.UI, config configuration.Reader, spaceRoleSetter user.SpaceRoleSetter, spaceRepo api.SpaceRepository, orgRepo api.OrganizationRepository, userRepo api.UserRepository) (cmd CreateSpace) {
	cmd.ui = ui
	cmd.config = config
	cmd.spaceRoleSetter = spaceRoleSetter
	cmd.spaceRepo = spaceRepo
	cmd.orgRepo = orgRepo
	cmd.userRepo = userRepo
	return
}

func (command CreateSpace) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "create-space",
		Description: "Create a space",
		Usage:       "CF_NAME create-space SPACE [-o ORG]",
		Flags: []cli.Flag{
			flag_helpers.NewStringFlag("o", "Organization"),
		},
	}
}

func (cmd CreateSpace) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) == 0 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "create-space")
		return
	}

	reqs = []requirements.Requirement{requirementsFactory.NewLoginRequirement()}
	if c.String("o") == "" {
		reqs = append(reqs, requirementsFactory.NewTargetedOrgRequirement())
	}

	return
}

func (cmd CreateSpace) Run(c *cli.Context) {
	spaceName := c.Args()[0]
	orgName := c.String("o")
	orgGuid := ""
	if orgName == "" {
		orgName = cmd.config.OrganizationFields().Name
		orgGuid = cmd.config.OrganizationFields().Guid
	}

	cmd.ui.Say("Creating space %s in org %s as %s...",
		terminal.EntityNameColor(spaceName),
		terminal.EntityNameColor(orgName),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	if orgGuid == "" {
		org, apiErr := cmd.orgRepo.FindByName(orgName)
		switch apiErr.(type) {
		case nil:
		case *errors.ModelNotFoundError:
			cmd.ui.Failed("Org %s does not exist or is not accessible", orgName)
			return
		default:
			cmd.ui.Failed("Error finding org %s\n%s", orgName, apiErr.Error())
			return
		}

		orgGuid = org.Guid
	}

	space, err := cmd.spaceRepo.Create(spaceName, orgGuid)
	if err != nil {
		if httpErr, ok := err.(errors.HttpError); ok && httpErr.ErrorCode() == errors.SPACE_EXISTS {
			cmd.ui.Ok()
			cmd.ui.Warn("Space %s already exists", spaceName)
			return
		}
		cmd.ui.Failed(err.Error())
		return
	}
	cmd.ui.Ok()

	err = cmd.spaceRoleSetter.SetSpaceRole(space, models.SPACE_MANAGER, cmd.config.UserGuid(), cmd.config.Username())
	if err != nil {
		cmd.ui.Failed(err.Error())
		return
	}

	err = cmd.spaceRoleSetter.SetSpaceRole(space, models.SPACE_DEVELOPER, cmd.config.UserGuid(), cmd.config.Username())
	if err != nil {
		cmd.ui.Failed(err.Error())
		return
	}

	cmd.ui.Say("\nTIP: Use '%s' to target new space", terminal.CommandColor(cf.Name()+" target -o "+orgName+" -s "+space.Name))
}
