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
	CreateRoute(hostName string, domain cf.DomainFields, space cf.SpaceFields) (route cf.Route, apiResponse net.ApiResponse)
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
		reqFactory.NewTargetedOrgRequirement(),
		cmd.spaceReq,
		cmd.domainReq,
	}
	return
}

func (cmd *CreateRoute) Run(c *cli.Context) {
	hostName := c.String("n")
	space := cmd.spaceReq.GetSpace()
	domain := cmd.domainReq.GetDomain()

	_, apiResponse := cmd.CreateRoute(hostName, domain.DomainFields, space.SpaceFields)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}
}

func (cmd *CreateRoute) CreateRoute(hostName string, domain cf.DomainFields, space cf.SpaceFields) (route cf.Route, apiResponse net.ApiResponse) {
	cmd.ui.Say("Creating route %s for org %s / space %s as %s...",
		terminal.EntityNameColor(domain.UrlForHost(hostName)),
		terminal.EntityNameColor(cmd.config.OrganizationFields.Name),
		terminal.EntityNameColor(space.Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	_, apiResponse = cmd.routeRepo.CreateInSpace(hostName, domain.Guid, space.Guid)
	if apiResponse.IsNotSuccessful() {
		var findApiResponse net.ApiResponse
		route, findApiResponse = cmd.routeRepo.FindByHostAndDomain(hostName, domain.Name)

		if findApiResponse.IsNotSuccessful() ||
			route.Space.Guid != space.Guid ||
			route.Domain.Guid != domain.Guid ||
			route.Host != hostName {
			return
		}

		apiResponse = net.NewSuccessfulApiResponse()
		cmd.ui.Ok()
		cmd.ui.Warn("Route %s already exists", route.URL())
		return
	}

	cmd.ui.Ok()
	return
}
