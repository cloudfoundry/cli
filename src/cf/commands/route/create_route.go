package route

import (
	"cf"
	"cf/api"
	"cf/configuration"
	"cf/net"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type RouteCreator interface {
	CreateRoute(hostName string, domain cf.Domain, space cf.Space) (route cf.Route, apiResponse net.ApiResponse)
}

type CreateRoute struct {
	ui        terminal.UI
	config    *configuration.Configuration
	routeRepo api.RouteRepository
	spaceReq  requirements.SpaceRequirement
	domainReq requirements.DomainRequirement
}

func NewCreateRoute(ui terminal.UI, config *configuration.Configuration, routeRepo api.RouteRepository) (cmd *CreateRoute) {
	cmd = new(CreateRoute)
	cmd.ui = ui
	cmd.config = config
	cmd.routeRepo = routeRepo
	return
}

func (cmd *CreateRoute) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {

	if len(c.Args()) != 2 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "create-route")
		return
	}

	spaceName := c.Args()[0]
	domainName := c.Args()[1]

	cmd.spaceReq = reqFactory.NewSpaceRequirement(spaceName)
	cmd.domainReq = reqFactory.NewDomainRequirement(domainName)

	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
		cmd.spaceReq,
		cmd.domainReq,
	}
	return
}

func (cmd *CreateRoute) Run(c *cli.Context) {
	hostName := c.String("n")
	space := cmd.spaceReq.GetSpace()
	domain := cmd.domainReq.GetDomain()

	_, apiResponse := cmd.CreateRoute(hostName, domain, space)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}
}

func (cmd *CreateRoute) CreateRoute(hostName string, domain cf.Domain, space cf.Space) (route cf.Route, apiResponse net.ApiResponse) {
	routeToCreate := cf.Route{Host: hostName, Domain: domain}

	cmd.ui.Say("Creating route %s for org %s / space %s as %s...",
		terminal.EntityNameColor(routeToCreate.URL()),
		terminal.EntityNameColor(cmd.config.Organization.Name),
		terminal.EntityNameColor(space.Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	route, apiResponse = cmd.routeRepo.CreateInSpace(routeToCreate, domain, space)
	if apiResponse.IsNotSuccessful() {
		return
	}

	cmd.ui.Ok()
	return
}
