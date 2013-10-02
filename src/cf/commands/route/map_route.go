package route

import (
	"cf/api"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type MapRoute struct {
	ui        terminal.UI
	routeRepo api.RouteRepository
	appReq    requirements.ApplicationRequirement
	routeReq  requirements.RouteRequirement
}

func NewMapRoute(ui terminal.UI, routeRepo api.RouteRepository) (cmd *MapRoute) {
	cmd = new(MapRoute)
	cmd.ui = ui
	cmd.routeRepo = routeRepo
	return
}

func (cmd *MapRoute) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 2 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "map-route")
		return
	}

	appName := c.Args()[0]
	domainName := c.Args()[1]
	hostName := c.String("n")

	cmd.appReq = reqFactory.NewApplicationRequirement(appName)
	cmd.routeReq = reqFactory.NewRouteRequirement(hostName, domainName)

	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
		cmd.appReq,
		cmd.routeReq,
	}
	return
}

func (cmd *MapRoute) Run(c *cli.Context) {
	route := cmd.routeReq.GetRoute()
	app := cmd.appReq.GetApplication()

	cmd.ui.Say("Adding url route %s to app %s",
		terminal.EntityNameColor(route.URL()),
		terminal.EntityNameColor(app.Name))

	apiStatus := cmd.routeRepo.Bind(route, app)
	if apiStatus.IsError() {
		cmd.ui.Failed(apiStatus.Message)
	}

	cmd.ui.Ok()
}
