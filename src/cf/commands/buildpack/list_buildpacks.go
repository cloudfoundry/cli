package buildpack

import (
	"cf/api"
	"cf/requirements"
	"cf/terminal"
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

func (cmd ListBuildpacks) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
	}
	return
}

func (cmd ListBuildpacks) Run(c *cli.Context) {
	cmd.ui.Say("Getting buildpacks...")

	buildpacks, apiResponse := cmd.buildpackRepo.FindAll()
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say("")

	if len(buildpacks) == 0 {
		cmd.ui.Say("No buildpacks found")
		return
	}

	table := [][]string{
		{"buildpack", "position"},
	}

	for _, buildpack := range buildpacks {
		position := ""
		if buildpack.Position != nil {
			position = strconv.Itoa(*buildpack.Position)
		}
		table = append(table, []string{
			buildpack.Name,
			position,
		})
	}

	cmd.ui.DisplayTable(table)
}
