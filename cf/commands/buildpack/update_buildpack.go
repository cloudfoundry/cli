package buildpack

import (
	"errors"
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/flag_helpers"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
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

func (command *UpdateBuildpack) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "update-buildpack",
		Description: "Update a buildpack",
		Usage:       "CF_NAME update-buildpack BUILDPACK [-p PATH] [-i POSITION] [--enable|--disable] [--lock|--unlock]",
		Flags: []cli.Flag{
			flag_helpers.NewIntFlag("i", "Buildpack position among other buildpacks"),
			flag_helpers.NewStringFlag("p", "Path to directory or zip file"),
			cli.BoolFlag{Name: "enable", Usage: "Enable the buildpack"},
			cli.BoolFlag{Name: "disable", Usage: "Disable the buildpack"},
			cli.BoolFlag{Name: "lock", Usage: "Lock the buildpack"},
			cli.BoolFlag{Name: "unlock", Usage: "Unlock the buildpack"},
		},
	}
}

func (cmd *UpdateBuildpack) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "update-buildpack")
		return
	}

	loginReq := requirementsFactory.NewLoginRequirement()
	cmd.buildpackReq = requirementsFactory.NewBuildpackRequirement(c.Args()[0])

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

	if c.IsSet("i") {
		position := c.Int("i")

		buildpack.Position = &position
		updateBuildpack = true
	}

	enabled := c.Bool("enable")
	disabled := c.Bool("disable")
	if enabled && disabled {
		cmd.ui.Failed("Cannot specify both enabled and disabled options.")
		return
	}

	if enabled {
		buildpack.Enabled = &enabled
		updateBuildpack = true
	}
	if disabled {
		disabled = false
		buildpack.Enabled = &disabled
		updateBuildpack = true
	}

	lock := c.Bool("lock")
	unlock := c.Bool("unlock")
	if lock && unlock {
		cmd.ui.Failed("Cannot specify both lock and unlock options.")
		return
	}

	dir := c.String("p")
	if dir != "" && (lock || unlock) {
		cmd.ui.Failed("Cannot specify buildpack bits and lock/unlock.")
	}

	if lock {
		buildpack.Locked = &lock
		updateBuildpack = true
	}
	if unlock {
		unlock = false
		buildpack.Locked = &unlock
		updateBuildpack = true
	}

	if updateBuildpack {
		newBuildpack, apiErr := cmd.buildpackRepo.Update(buildpack)
		if apiErr != nil {
			cmd.ui.Failed("Error updating buildpack %s\n%s", terminal.EntityNameColor(buildpack.Name), apiErr.Error())
		}
		buildpack = newBuildpack
	}

	if dir != "" {
		apiErr := cmd.buildpackBitsRepo.UploadBuildpack(buildpack, dir)
		if apiErr != nil {
			cmd.ui.Failed("Error uploading buildpack %s\n%s", terminal.EntityNameColor(buildpack.Name), apiErr.Error())
		}
	}
	cmd.ui.Ok()
}
