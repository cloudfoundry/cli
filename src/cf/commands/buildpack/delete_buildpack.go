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
}

func NewDeleteBuildpack(ui terminal.UI, repo api.BuildpackRepository) (cmd *DeleteBuildpack) {
	cmd = new(DeleteBuildpack)
	cmd.ui = ui
	cmd.buildpackRepo = repo
	return
}

func (cmd *DeleteBuildpack) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "delete-buildpack")
		return
	}

	loginReq := reqFactory.NewLoginRequirement()

	reqs = []requirements.Requirement{
		loginReq,
	}

	return
}

func (cmd *DeleteBuildpack) Run(c *cli.Context) {
	buildpackName := c.Args()[0]

	force := c.Bool("f")

	if !force {
		answer := cmd.ui.Confirm("Are you sure you want to delete the buildpack %s ?", terminal.EntityNameColor(buildpackName))
		if !answer {
			return
		}
	}

	cmd.ui.Say("Deleting buildpack %s...", terminal.EntityNameColor(buildpackName))

	buildpack, apiResponse := cmd.buildpackRepo.FindByName(buildpackName)

	if apiResponse.IsNotFound() {
		cmd.ui.Ok()
		cmd.ui.Warn("Buildpack %s does not exist.", buildpackName)
		return
	}

	if apiResponse != nil {
		cmd.ui.Failed(apiResponse.Error())
		return
	}

	apiResponse = cmd.buildpackRepo.Delete(buildpack.Guid)
	if apiResponse != nil {
		cmd.ui.Failed("Error deleting buildpack %s\n%s", terminal.EntityNameColor(buildpack.Name), apiResponse.Error())
		return
	}

	cmd.ui.Ok()
}
