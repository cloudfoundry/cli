package buildpack

import (
	"strconv"

	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
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

func (cmd CreateBuildpack) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "create-buildpack",
		Description: "Create a buildpack",
		Usage: "CF_NAME create-buildpack BUILDPACK PATH POSITION [--enable|--disable]" +
			"\n\nTIP:\n" +
			"   Path should be a zip file, a url to a zip file, or a local directory. Position is an integer, sets priority, and is sorted from lowest to highest.",
		Flags: []cli.Flag{
			cli.BoolFlag{Name: "enable", Usage: "Enable the buildpack"},
			cli.BoolFlag{Name: "disable", Usage: "Disable the buildpack"},
		},
	}
}

func (cmd CreateBuildpack) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}
	return
}

func (cmd CreateBuildpack) Run(c *cli.Context) {
	if len(c.Args()) != 3 {
		cmd.ui.FailWithUsage(c)
		return
	}

	buildpackName := c.Args()[0]

	cmd.ui.Say("Creating buildpack %s...", terminal.EntityNameColor(buildpackName))

	buildpack, err := cmd.createBuildpack(buildpackName, c)

	if err != nil {
		if err, ok := err.(errors.HttpError); ok && err.ErrorCode() == errors.BUILDPACK_EXISTS {
			cmd.ui.Ok()
			cmd.ui.Warn("Buildpack %s already exists", buildpackName)
			cmd.ui.Say("TIP: use '%s' to update this buildpack", terminal.CommandColor(cf.Name()+" "+"update-buildpack"))
		} else {
			cmd.ui.Failed(err.Error())
		}
		return
	}
	cmd.ui.Ok()
	cmd.ui.Say("")

	cmd.ui.Say("Uploading buildpack %s...", terminal.EntityNameColor(buildpackName))

	dir := c.Args()[1]

	err = cmd.buildpackBitsRepo.UploadBuildpack(buildpack, dir)
	if err != nil {
		cmd.ui.Failed(err.Error())
		return
	}

	cmd.ui.Ok()
}

func (cmd CreateBuildpack) createBuildpack(buildpackName string, c *cli.Context) (buildpack models.Buildpack, apiErr error) {
	position, err := strconv.Atoi(c.Args()[2])
	if err != nil {
		apiErr = errors.NewWithFmt("Invalid position. %s", err.Error())
		return
	}

	enabled := c.Bool("enable")
	disabled := c.Bool("disable")
	if enabled && disabled {
		apiErr = errors.New("Cannot specify both enabled and disabled.")
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

	buildpack, apiErr = cmd.buildpackRepo.Create(buildpackName, &position, enableOption, nil)

	return
}
