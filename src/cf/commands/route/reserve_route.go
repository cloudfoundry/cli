package route

import (
	"cf"
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type ReserveRoute struct {
	ui        terminal.UI
	config    *configuration.Configuration
	routeRepo api.RouteRepository
	spaceReq  requirements.SpaceRequirement
	domainReq requirements.DomainRequirement
}

func NewReserveRoute(ui terminal.UI, config *configuration.Configuration, routeRepo api.RouteRepository) (cmd *ReserveRoute) {
	cmd = new(ReserveRoute)
	cmd.ui = ui
	cmd.config = config
	cmd.routeRepo = routeRepo
	return
}

func (cmd *ReserveRoute) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {

	if len(c.Args()) != 2 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "reserve-route")
		return
	}

	cmd.spaceReq = reqFactory.NewSpaceRequirement(c.Args()[0])
	cmd.domainReq = reqFactory.NewDomainRequirement(c.Args()[1])
	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
		cmd.spaceReq,
		cmd.domainReq,
	}
	return
}

func (cmd *ReserveRoute) Run(c *cli.Context) {
	space := cmd.spaceReq.GetSpace()
	domain := cmd.domainReq.GetDomain()
	route := cf.Route{Host: c.String("n"), Domain: domain}

	cmd.ui.Say("Reserving route %s for org %s / space %s as %s...",
		terminal.EntityNameColor(route.URL()),
		terminal.EntityNameColor(cmd.config.Organization.Name),
		terminal.EntityNameColor(space.Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	_, apiResponse := cmd.routeRepo.CreateInSpace(route, domain, space)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()
}
