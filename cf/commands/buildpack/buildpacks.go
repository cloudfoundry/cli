package buildpack

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
	"strconv"
)

type ListBuildpacks struct {
	ui            terminal.UI
	buildpackRepo api.BuildpackRepository
}

func NewListBuildpacks(ui terminal.UI, buildpackRepo api.BuildpackRepository) (cmd ListBuildpacks) {
	cmd.ui = ui
	cmd.buildpackRepo = buildpackRepo
	return
}

func (command ListBuildpacks) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "buildpacks",
		Description: "List all buildpacks",
		Usage:       "CF_NAME buildpacks",
	}
}

func (cmd ListBuildpacks) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}
	return
}

func (cmd ListBuildpacks) Run(c *cli.Context) {
	cmd.ui.Say("Getting buildpacks...\n")

	table := cmd.ui.Table([]string{"buildpack", "position", "enabled", "locked", "filename"})
	noBuildpacks := true

	apiErr := cmd.buildpackRepo.ListBuildpacks(func(buildpack models.Buildpack) bool {
		position := ""
		if buildpack.Position != nil {
			position = strconv.Itoa(*buildpack.Position)
		}
		enabled := ""
		if buildpack.Enabled != nil {
			enabled = strconv.FormatBool(*buildpack.Enabled)
		}
		locked := ""
		if buildpack.Locked != nil {
			locked = strconv.FormatBool(*buildpack.Locked)
		}
		table.Add([]string{
			buildpack.Name,
			position,
			enabled,
			locked,
			buildpack.Filename,
		})
		noBuildpacks = false
		return true
	})
	table.Print()

	if apiErr != nil {
		cmd.ui.Failed("Failed fetching buildpacks.\n%s", apiErr.Error())
		return
	}

	if noBuildpacks {
		cmd.ui.Say("No buildpacks found")
	}
}
