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

func (r Routes) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	return
}

func (r Routes) Run(c *cli.Context) {
	r.ui.Say("Getting routes")

	routes, err := r.routeRepo.FindAll()

	if err != nil {
		r.ui.Failed(err.Error())
	}

	r.ui.Ok()

	if len(routes) == 0 {
		r.ui.Say("No routes found")
		return
	}

	for _, route := range routes {
		r.ui.Say(route.URL())
	}
}
