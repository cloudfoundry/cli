package route

import (
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type MapRoute struct {
	ui           terminal.UI
	config       configuration.Reader
	routeRepo    api.RouteRepository
	appReq       requirements.ApplicationRequirement
	domainReq    requirements.DomainRequirement
	routeCreator RouteCreator
}

func NewMapRoute(ui terminal.UI, config configuration.Reader, routeRepo api.RouteRepository, routeCreator RouteCreator) (cmd *MapRoute) {
	cmd = new(MapRoute)
	cmd.ui = ui
	cmd.config = config
	cmd.routeRepo = routeRepo
	cmd.routeCreator = routeCreator
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

	cmd.appReq = reqFactory.NewApplicationRequirement(appName)
	cmd.domainReq = reqFactory.NewDomainRequirement(domainName)

	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
		cmd.appReq,
		cmd.domainReq,
	}
	return
}

func (cmd *MapRoute) Run(c *cli.Context) {
	hostName := c.String("n")
	domain := cmd.domainReq.GetDomain()
	app := cmd.appReq.GetApplication()

	route, apiResponse := cmd.routeCreator.CreateRoute(hostName, domain, cmd.config.SpaceFields())
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed("Error resolving route:\n%s", apiResponse.Message)
	}
	cmd.ui.Say("Adding route %s to app %s in org %s / space %s as %s...",
		terminal.EntityNameColor(route.URL()),
		terminal.EntityNameColor(app.Name),
		terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
		terminal.EntityNameColor(cmd.config.SpaceFields().Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	apiResponse = cmd.routeRepo.Bind(route.Guid, app.Guid)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()
}
