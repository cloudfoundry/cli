package buildpack

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type DeleteBuildpack struct {
	ui            terminal.UI
	buildpackRepo api.BuildpackRepository
}

func NewDeleteBuildpack(ui terminal.UI, repo api.BuildpackRepository) (cmd *DeleteBuildpack) {
	cmd = new(DeleteBuildpack)
	cmd.ui = ui
	cmd.buildpackRepo = repo
	return
}

func (command *DeleteBuildpack) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "delete-buildpack",
		Description: "Delete a buildpack",
		Usage:       "CF_NAME delete-buildpack BUILDPACK [-f]",
		Flags: []cli.Flag{
			cli.BoolFlag{Name: "f", Usage: "Force deletion without confirmation"},
		},
	}
}

func (cmd *DeleteBuildpack) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "delete-buildpack")
		return
	}

	loginReq := requirementsFactory.NewLoginRequirement()

	reqs = []requirements.Requirement{
		loginReq,
	}

	return
}

func (cmd *DeleteBuildpack) Run(c *cli.Context) {
	buildpackName := c.Args()[0]

	force := c.Bool("f")

	if !force {
		answer := cmd.ui.ConfirmDelete("buildpack", buildpackName)
		if !answer {
			return
		}
	}

	cmd.ui.Say("Deleting buildpack %s...", terminal.EntityNameColor(buildpackName))

	buildpack, apiErr := cmd.buildpackRepo.FindByName(buildpackName)

	switch apiErr.(type) {
	case nil: //do nothing
	case *errors.ModelNotFoundError:
		cmd.ui.Ok()
		cmd.ui.Warn("Buildpack %s does not exist.", buildpackName)
		return
	default:
		cmd.ui.Failed(apiErr.Error())
		return

	}

	apiErr = cmd.buildpackRepo.Delete(buildpack.Guid)
	if apiErr != nil {
		cmd.ui.Failed("Error deleting buildpack %s\n%s", terminal.EntityNameColor(buildpack.Name), apiErr.Error())
		return
	}

	cmd.ui.Ok()
}
