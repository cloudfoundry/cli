package buildpack

import (
	"cf/api"
	"cf/errors"
	"cf/requirements"
	"cf/terminal"
	"github.com/codegangsta/cli"
)

type RenameBuildpack struct {
	ui            terminal.UI
	buildpackRepo api.BuildpackRepository
}

func NewRenameBuildpack(ui terminal.UI, repo api.BuildpackRepository) (cmd *RenameBuildpack) {
	cmd = new(RenameBuildpack)
	cmd.ui = ui
	cmd.buildpackRepo = repo
	return
}

func (cmd *RenameBuildpack) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 2 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "rename-buildpack")
		return
	}

	reqs = []requirements.Requirement{reqFactory.NewLoginRequirement()}
	return
}

func (cmd *RenameBuildpack) Run(c *cli.Context) {
	buildpackName := c.Args()[0]
	newBuildpackName := c.Args()[1]

	cmd.ui.Say("Renaming buildpack %s to %s...", terminal.EntityNameColor(buildpackName), terminal.EntityNameColor(newBuildpackName))

	buildpack, apiErr := cmd.buildpackRepo.FindByName(buildpackName)

	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
	}

	buildpack.Name = newBuildpackName
	buildpack, apiErr = cmd.buildpackRepo.Update(buildpack)
	if apiErr != nil {
		cmd.ui.Failed("Error renaming buildpack %s\n%s", terminal.EntityNameColor(buildpackName), apiErr.Error())
		return
	}

	cmd.ui.Ok()
}
