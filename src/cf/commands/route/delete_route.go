package route

import (
	"cf/api"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type DeleteRoute struct {
	ui        terminal.UI
	routeRepo api.RouteRepository
}

func NewDeleteRoute(ui terminal.UI, routeRepo api.RouteRepository) (cmd *DeleteRoute) {
	cmd = &DeleteRoute{
		ui:        ui,
		routeRepo: routeRepo,
	}
	return
}

func (cmd *DeleteRoute) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {

	if len(c.Args()) != 1 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "delete-route")
		return
	}

	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
	}
	return
}

func (cmd *DeleteRoute) Run(c *cli.Context) {
	host := c.String("n")
	domainName := c.Args()[0]

	url := domainName
	if host != "" {
		url = host + "." + domainName
	}
	force := c.Bool("f")
	if !force {
		response := cmd.ui.Confirm(
			"Really delete route %s?%s",
			terminal.EntityNameColor(url),
			terminal.PromptColor(">"),
		)

		if !response {
			return
		}
	}

	cmd.ui.Say("Deleting route %s...", terminal.EntityNameColor(url))

	route, apiResponse := cmd.routeRepo.FindByHostAndDomain(host, domainName)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	apiResponse = cmd.routeRepo.Delete(route)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()
}
