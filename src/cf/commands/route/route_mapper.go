package route

import (
	"cf/api"
	"cf/net"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type RouteMapper struct {
	ui        terminal.UI
	routeRepo api.RouteRepository
	appReq    requirements.ApplicationRequirement
	routeReq  requirements.RouteRequirement
	bind      bool
}

func NewRouteMapper(ui terminal.UI, routeRepo api.RouteRepository, bind bool) (cmd *RouteMapper) {
	cmd = new(RouteMapper)
	cmd.ui = ui
	cmd.routeRepo = routeRepo
	cmd.bind = bind
	return
}

func (cmd *RouteMapper) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 2 {
		err = errors.New("Incorrect Usage")
		if cmd.bind {
			cmd.ui.FailWithUsage(c, "map-route")
		} else {
			cmd.ui.FailWithUsage(c, "unmap-route")
		}
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

func (cmd *RouteMapper) Run(c *cli.Context) {
	route := cmd.routeReq.GetRoute()
	app := cmd.appReq.GetApplication()

	var apiResponse net.ApiResponse

	if cmd.bind {
		cmd.ui.Say("Adding url route %s to app %s...",
			terminal.EntityNameColor(route.URL()),
			terminal.EntityNameColor(app.Name))

		apiResponse = cmd.routeRepo.Bind(route, app)
	} else {
		cmd.ui.Say("Removing url route %s from app %s...",
			terminal.EntityNameColor(route.URL()),
			terminal.EntityNameColor(app.Name))

		apiResponse = cmd.routeRepo.Unbind(route, app)
	}

	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()
}
