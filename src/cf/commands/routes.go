package commands

import (
	"cf/api"
	"cf/requirements"
	"cf/terminal"
	"github.com/codegangsta/cli"
)

type Routes struct {
	ui        terminal.UI
	routeRepo api.RouteRepository
}

func NewRoutes(ui terminal.UI, routeRepo api.RouteRepository) (cmd *Routes) {
	cmd = new(Routes)
	cmd.ui = ui
	cmd.routeRepo = routeRepo
	return
}

func (cmd Routes) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	return
}

func (cmd Routes) Run(c *cli.Context) {
	cmd.ui.Say("Getting routes")

	routes, err := cmd.routeRepo.FindAll()

	if err != nil {
		cmd.ui.Failed(err.Error())
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
