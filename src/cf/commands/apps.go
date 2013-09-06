package commands

import (
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	term "cf/terminal"
	"fmt"
	"github.com/codegangsta/cli"
	"strings"
)

type Apps struct {
	ui        term.UI
	config    *configuration.Configuration
	spaceRepo api.SpaceRepository
}

func NewApps(ui term.UI, config *configuration.Configuration, spaceRepo api.SpaceRepository) (a Apps) {
	a.ui = ui
	a.config = config
	a.spaceRepo = spaceRepo
	return
}

func (a Apps) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
		reqFactory.NewSpaceRequirement(),
	}
	return
}

func (a Apps) Run(c *cli.Context) {
	a.ui.Say("Getting applications in %s", a.config.Space.Name)

	space, err := a.spaceRepo.GetSummary()

	if err != nil {
		a.ui.Failed("Error loading applications", err)
		return
	}

	apps := space.Applications

	a.ui.Ok()

	table := [][]string{
		[]string{"name", "status", "usage", "url"},
	}

	for _, app := range apps {
		table = append(table, []string{
			app.Name,
			app.State,
			fmt.Sprintf("%d x %dM", app.Instances, app.Memory),
			strings.Join(app.Urls, ", "),
		})
	}

	a.ui.DisplayTable(table, a.coloringFunc)
}

func (a Apps) coloringFunc(value string, row int, col int) string {
	if row > 0 && col == 1 {
		return coloredState(value)
	}

	return term.DefaultColoringFunc(value, row, col)
}

func coloredState(state string) (colored string) {
	switch state {
	case "started":
		colored = term.Green("running")
	case "stopped":
		colored = term.Yellow("stopped")
	default:
		colored = term.Red(state)
	}

	return
}
