package buildpack

import (
	"cf/api"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type DeleteBuildpack struct {
	ui            terminal.UI
	buildpackRepo api.BuildpackRepository
	buildpackReq  requirements.BuildpackRequirement
}

func NewDeleteBuildpack(ui terminal.UI, repo api.BuildpackRepository) (cmd *DeleteBuildpack) {
	cmd = &DeleteBuildpack{ui: ui, buildpackRepo: repo}
	return
}

func (cmd *DeleteBuildpack) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "delete-buildpack")
		return
	}

	loginReq := reqFactory.NewLoginRequirement()
	cmd.buildpackReq = reqFactory.NewBuildpackRequirement(c.Args()[0])

	reqs = []requirements.Requirement{
		loginReq,
		cmd.buildpackReq,
	}

	return
}

func (cmd *DeleteBuildpack) Run(c *cli.Context) {
	buildpack := cmd.buildpackReq.GetBuildpack()
	force := c.Bool("f")

	cmd.ui.Say("Deleting buildpack %s...", terminal.EntityNameColor(buildpack.Name))

	if !force {
		answer := cmd.ui.Confirm("Are you sure you want to delete the buildpack %s ?", terminal.EntityNameColor(buildpack.Name))
		if !answer {
			return
		}
	}

	apiResponse := cmd.buildpackRepo.Delete(buildpack)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed("Error deleting buildpack %s\n%s", buildpack.Name, apiResponse.Message)
		return
	}

	cmd.ui.Ok()
}
