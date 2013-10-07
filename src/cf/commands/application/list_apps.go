package application

import (
	"cf/api"
	"cf/requirements"
	"cf/terminal"
	"fmt"
	"github.com/codegangsta/cli"
	"strings"
)

type ListApps struct {
	ui        terminal.UI
	spaceRepo api.SpaceRepository
}

func NewListApps(ui terminal.UI, spaceRepo api.SpaceRepository) (cmd ListApps) {
	cmd.ui = ui
	cmd.spaceRepo = spaceRepo
	return
}

func (cmd ListApps) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
		reqFactory.NewTargetedSpaceRequirement(),
	}
	return
}

func (cmd ListApps) Run(c *cli.Context) {
	cmd.ui.Say("Getting apps in %s...", cmd.spaceRepo.GetCurrentSpace().Name)

	space, apiResponse := cmd.spaceRepo.GetSummary()

	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	apps := space.Applications

	cmd.ui.Ok()

	table := [][]string{
		[]string{"name", "status", "usage", "urls"},
	}

	for _, app := range apps {
		table = append(table, []string{
			app.Name,
			app.State,
			fmt.Sprintf("%d x %s", app.Instances, byteSize(app.Memory*MEGABYTE)),
			strings.Join(app.Urls, ", "),
		})
	}

	cmd.ui.DisplayTable(table, cmd.coloringFunc)
}

func (cmd ListApps) coloringFunc(value string, row int, col int) string {
	if row > 0 && col == 1 {
		return coloredState(value)
	}

	return terminal.DefaultColoringFunc(value, row, col)
}
