package commands

import (
	"cf/api"
	"cf/requirements"
	"cf/terminal"
	"fmt"
	"github.com/codegangsta/cli"
	"strings"
)

type Apps struct {
	ui        terminal.UI
	spaceRepo api.SpaceRepository
}

func NewApps(ui terminal.UI, spaceRepo api.SpaceRepository) (cmd Apps) {
	cmd.ui = ui
	cmd.spaceRepo = spaceRepo
	return
}

func (cmd Apps) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
		reqFactory.NewTargetedSpaceRequirement(),
	}
	return
}

func (cmd Apps) Run(c *cli.Context) {
	cmd.ui.Say("Getting applications in %s...", cmd.spaceRepo.GetCurrentSpace().Name)

	space, err := cmd.spaceRepo.GetSummary()

	if err != nil {
		cmd.ui.Failed(err.Error())
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

func (cmd Apps) coloringFunc(value string, row int, col int) string {
	if row > 0 && col == 1 {
		return coloredState(value)
	}

	return terminal.DefaultColoringFunc(value, row, col)
}
