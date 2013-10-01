package route

import (
	"cf/api"
	"cf/requirements"
	"cf/terminal"
	"github.com/codegangsta/cli"
)

type ListRoutes struct {
	ui        terminal.UI
	routeRepo api.RouteRepository
}

func NewListRoutes(ui terminal.UI, routeRepo api.RouteRepository) (cmd *ListRoutes) {
	cmd = new(ListRoutes)
	cmd.ui = ui
	cmd.routeRepo = routeRepo
	return
}

func (cmd ListRoutes) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	return
}

func (cmd ListRoutes) Run(c *cli.Context) {
	cmd.ui.Say("Getting routes")

	routes, apiStatus := cmd.routeRepo.FindAll()

	if apiStatus.IsError() {
		cmd.ui.Failed(apiStatus.Message)
		return
	}

	cmd.ui.Ok()

	if len(routes) == 0 {
		cmd.ui.Say("No routes found")
		return
	}

	for _, route := range routes {
		cmd.ui.Say(route.URL())
	}
}
