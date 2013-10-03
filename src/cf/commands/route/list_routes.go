package route

import (
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	"cf/terminal"
	"github.com/codegangsta/cli"
	"strings"
)

type ListRoutes struct {
	ui        terminal.UI
	routeRepo api.RouteRepository
	config    *configuration.Configuration
}

func NewListRoutes(ui terminal.UI, config *configuration.Configuration, routeRepo api.RouteRepository) (cmd *ListRoutes) {
	cmd = new(ListRoutes)
	cmd.ui = ui
	cmd.config = config
	cmd.routeRepo = routeRepo
	return
}

func (cmd ListRoutes) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	return
}

func (cmd ListRoutes) Run(c *cli.Context) {
	cmd.ui.Say("Getting routes in space %s...", terminal.EntityNameColor(cmd.config.Space.Name))

	routes, apiStatus := cmd.routeRepo.FindAll()

	if apiStatus.NotSuccessful() {
		cmd.ui.Failed(apiStatus.Message)
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say("")

	if len(routes) == 0 {
		cmd.ui.Say("No routes found")
		return
	}

	table := [][]string{
		{"host", "domain", "apps"},
	}

	for _, route := range routes {
		table = append(table, []string{
			route.Host,
			route.Domain.Name,
			strings.Join(route.AppNames, ", "),
		})
	}

	cmd.ui.DisplayTable(table, nil)
}
