package route

import (
	"cf/api"
	"cf/configuration"
	"cf/models"
	"cf/requirements"
	"cf/terminal"
	"github.com/codegangsta/cli"
)

type DeleteOrphanedRoutes struct {
	ui        terminal.UI
	routeRepo api.RouteRepository
	config    configuration.Reader
}

func NewDeleteOrphanedRoutes(ui terminal.UI, config configuration.Reader, routeRepo api.RouteRepository) (cmd DeleteOrphanedRoutes) {
	cmd.ui = ui
	cmd.config = config
	cmd.routeRepo = routeRepo
	return
}

func (cmd DeleteOrphanedRoutes) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	reqs = append(reqs, reqFactory.NewLoginRequirement())
	return
}

func (cmd DeleteOrphanedRoutes) Run(c *cli.Context) {

	force := c.Bool("f")
	if !force {
		response := cmd.ui.Confirm(
			"Really delete orphaned routes?%s",
			terminal.PromptColor(">"),
		)

		if !response {
			return
		}
	}

	cmd.ui.Say("Getting routes as %s ...\n",
		terminal.EntityNameColor(cmd.config.Username()),
	)

	apiErr := cmd.routeRepo.ListRoutes(func(route models.Route) bool {

		if len(route.Apps) == 0 {
			cmd.ui.Say("Deleting route %s...", terminal.EntityNameColor(route.Host+"."+route.Domain.Name))
			apiErr := cmd.routeRepo.Delete(route.Guid)
			if apiErr != nil {
				cmd.ui.Failed(apiErr.Error())
				return false
			}
		}
		return true
	})

	if apiErr != nil {
		cmd.ui.Failed("Failed fetching routes.\n%s", apiErr.Error())
		return
	}
	cmd.ui.Ok()
}
