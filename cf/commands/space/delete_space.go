package space

import (
	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/api/spaces"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type DeleteSpace struct {
	ui        terminal.UI
	config    configuration.ReadWriter
	spaceRepo spaces.SpaceRepository
	spaceReq  requirements.SpaceRequirement
}

func NewDeleteSpace(ui terminal.UI, config configuration.ReadWriter, spaceRepo spaces.SpaceRepository) (cmd *DeleteSpace) {
	cmd = new(DeleteSpace)
	cmd.ui = ui
	cmd.config = config
	cmd.spaceRepo = spaceRepo
	return
}

func (cmd *DeleteSpace) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "delete-space",
		Description: T("Delete a space"),
		Usage:       T("CF_NAME delete-space SPACE [-f]"),
		Flags: []cli.Flag{
			cli.BoolFlag{Name: "f", Usage: T("Force deletion without confirmation")},
		},
	}
}

func (cmd *DeleteSpace) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		cmd.ui.FailWithUsage(c)
	}

	cmd.spaceReq = requirementsFactory.NewSpaceRequirement(c.Args()[0])
	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedOrgRequirement(),
		cmd.spaceReq,
	}
	return
}

func (cmd *DeleteSpace) Run(c *cli.Context) {
	spaceName := c.Args()[0]

	if !c.Bool("f") {
		if !cmd.ui.ConfirmDelete(T("space"), spaceName) {
			return
		}
	}

	cmd.ui.Say(T("Deleting space {{.TargetSpace}} in org {{.TargetOrg}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"TargetSpace": terminal.EntityNameColor(spaceName),
			"TargetOrg":   terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			"CurrentUser": terminal.EntityNameColor(cmd.config.Username()),
		}))

	space := cmd.spaceReq.GetSpace()

	apiErr := cmd.spaceRepo.Delete(space.Guid)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Ok()

	if cmd.config.SpaceFields().Name == spaceName {
		cmd.config.SetSpaceFields(models.SpaceFields{})
		cmd.ui.Say(T("TIP: No space targeted, use '{{.CfTargetCommand}}' to target a space",
			map[string]interface{}{"CfTargetCommand": cf.Name() + " target -s"}))
	}

	return
}
