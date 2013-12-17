package buildpack

import (
	"cf/api"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type UpdateBuildpack struct {
	ui                terminal.UI
	buildpackRepo     api.BuildpackRepository
	buildpackBitsRepo api.BuildpackBitsRepository
	buildpackReq      requirements.BuildpackRequirement
}

func NewUpdateBuildpack(ui terminal.UI, repo api.BuildpackRepository, bitsRepo api.BuildpackBitsRepository) (cmd *UpdateBuildpack) {
	cmd = new(UpdateBuildpack)
	cmd.ui = ui
	cmd.buildpackRepo = repo
	cmd.buildpackBitsRepo = bitsRepo
	return
}

func (cmd *UpdateBuildpack) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "update-buildpack")
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

func (cmd *UpdateBuildpack) Run(c *cli.Context) {
	buildpack := cmd.buildpackReq.GetBuildpack()

	cmd.ui.Say("Updating buildpack %s...", terminal.EntityNameColor(buildpack.Name))

	updateBuildpack := false

	if c.String("i") != "" {
		val := c.Int("i")
		buildpack.Position = &val
		updateBuildpack = true
	}
	
	if c.String("e") != "" {
		val := c.Bool("e")
		buildpack.Enabled = &val
		updateBuildpack = true
	}
	

	if c.String("enabled") != "" {
		val := c.Bool("enabled")
		buildpack.Enabled = &val
		updateBuildpack = true
	}

	if updateBuildpack {
		buildpack, apiResponse := cmd.buildpackRepo.Update(buildpack)
		if apiResponse.IsNotSuccessful() {
			cmd.ui.Failed("Error updating buildpack %s\n%s", terminal.EntityNameColor(buildpack.Name), apiResponse.Message)
			return
		}
	}

	dir := c.String("p")
	if dir != "" {
		apiResponse := cmd.buildpackBitsRepo.UploadBuildpack(buildpack, dir)
		if apiResponse.IsNotSuccessful() {
			cmd.ui.Failed("Error uploading buildpack %s\n%s", terminal.EntityNameColor(buildpack.Name), apiResponse.Message)
			return
		}
	}
	cmd.ui.Ok()
}
