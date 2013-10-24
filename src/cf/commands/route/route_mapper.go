package route

import (
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type RouteMapper struct {
	ui           terminal.UI
	config       *configuration.Configuration
	routeRepo    api.RouteRepository
	appReq       requirements.ApplicationRequirement
	domainReq    requirements.DomainRequirement
	routeCreator RouteCreator
	bind         bool
}

func NewRouteMapper(ui terminal.UI, config *configuration.Configuration, routeRepo api.RouteRepository, routeCreator RouteCreator, bind bool) (cmd *RouteMapper) {
	cmd = new(RouteMapper)
	cmd.ui = ui
	cmd.config = config
	cmd.routeRepo = routeRepo
	cmd.routeCreator = routeCreator
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

	cmd.appReq = reqFactory.NewApplicationRequirement(appName)
	cmd.domainReq = reqFactory.NewDomainRequirement(domainName)

	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
		cmd.appReq,
		cmd.domainReq,
	}
	return
}

func (cmd *RouteMapper) Run(c *cli.Context) {

	// resolve the route we will bind to
	hostName := c.String("n")
	domain := cmd.domainReq.GetDomain()

	route, apiResponse := cmd.routeCreator.CreateRoute(hostName, domain, cmd.config.Space)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed("Error resolving route:\n%s", apiResponse.Message)
	}

	app := cmd.appReq.GetApplication()

	if cmd.bind {
		cmd.ui.Say("Adding route %s to app %s in org %s / space %s as %s...",
			terminal.EntityNameColor(route.URL()),
			terminal.EntityNameColor(app.Name),
			terminal.EntityNameColor(cmd.config.Organization.Name),
			terminal.EntityNameColor(cmd.config.Space.Name),
			terminal.EntityNameColor(cmd.config.Username()),
		)

		apiResponse = cmd.routeRepo.Bind(route, app)
	} else {
		cmd.ui.Say("Removing route %s from app %s in org %s / space %s as %s...",
			terminal.EntityNameColor(route.URL()),
			terminal.EntityNameColor(app.Name),
			terminal.EntityNameColor(cmd.config.Organization.Name),
			terminal.EntityNameColor(cmd.config.Space.Name),
			terminal.EntityNameColor(cmd.config.Username()),
		)

		apiResponse = cmd.routeRepo.Unbind(route, app)
	}

	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()
}
