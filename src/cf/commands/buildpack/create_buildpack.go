package buildpack

import (
	"cf"
	"cf/api"
	"cf/net"
	"cf/requirements"
	"cf/terminal"
	"github.com/codegangsta/cli"
	"strconv"
)

type CreateBuildpack struct {
	ui                terminal.UI
	buildpackRepo     api.BuildpackRepository
	buildpackBitsRepo api.BuildpackBitsRepository
}

func NewCreateBuildpack(ui terminal.UI, buildpackRepo api.BuildpackRepository, buildpackBitsRepo api.BuildpackBitsRepository) (cmd CreateBuildpack) {
	cmd.ui = ui
	cmd.buildpackRepo = buildpackRepo
	cmd.buildpackBitsRepo = buildpackBitsRepo
	return
}

func (cmd CreateBuildpack) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
	}
	return
}

func (cmd CreateBuildpack) Run(c *cli.Context) {
	if len(c.Args()) != 3 {
		cmd.ui.FailWithUsage(c, "create-buildpack")
		return
	}

	buildpackName := c.Args()[0]

	cmd.ui.Say("Creating buildpack %s...", terminal.EntityNameColor(buildpackName))

	buildpack, apiResponse := cmd.createBuildpack(buildpackName, c)
	if apiResponse.IsNotSuccessful() {
		if apiResponse.ErrorCode == cf.BUILDPACK_EXISTS {
			cmd.ui.Ok()
			cmd.ui.Warn("Buildpack %s already exists", buildpackName)
		} else {
			cmd.ui.Failed(apiResponse.Message)
		}
		return
	}
	cmd.ui.Ok()
	cmd.ui.Say("")

	cmd.ui.Say("Uploading buildpack %s...", terminal.EntityNameColor(buildpackName))

	dir := c.Args()[1]

	apiResponse = cmd.buildpackBitsRepo.UploadBuildpack(buildpack, dir)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()
}

func (cmd CreateBuildpack) createBuildpack(buildpackName string, c *cli.Context) (buildpack cf.Buildpack, apiResponse net.ApiResponse) {
	position, err := strconv.Atoi(c.Args()[2])
	if err != nil {
		apiResponse = net.NewApiResponseWithMessage("Invalid position. %s", err.Error())
		return
	}

	enabled := c.Bool("enable")
	disabled := c.Bool("disable")
	if enabled && disabled {
		apiResponse = net.NewApiResponseWithMessage("Cannot specify both enabled and disabled.")
		return
	}

	var enableOption *bool = nil
	if enabled {
		enableOption = &enabled
	}
	if disabled {
		disabled = false
		enableOption = &disabled
	}

	buildpack, apiResponse = cmd.buildpackRepo.Create(buildpackName, &position, enableOption)

	return
}
