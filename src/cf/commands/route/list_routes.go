package route

import (
	"cf/api"
	"cf/configuration"
	"cf/models"
	"cf/requirements"
	"cf/terminal"
	"github.com/codegangsta/cli"
	"strings"
)

type ListRoutes struct {
	ui        terminal.UI
	routeRepo api.RouteRepository
	config    configuration.Reader
}

func NewListRoutes(ui terminal.UI, config configuration.Reader, routeRepo api.RouteRepository) (cmd ListRoutes) {
	cmd.ui = ui
	cmd.config = config
	cmd.routeRepo = routeRepo
	return
}

func (cmd ListRoutes) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	reqs = append(reqs, reqFactory.NewLoginRequirement())
	return
}

func (cmd ListRoutes) Run(c *cli.Context) {
	cmd.ui.Say("Getting routes as %s ...\n",
		terminal.EntityNameColor(cmd.config.Username()),
	)

	table := cmd.ui.Table([]string{"host", "domain", "apps"})

	noRoutes := true
	apiErr := cmd.routeRepo.ListRoutes(func(route models.Route) bool {
		appNames := ""
		for _, app := range route.Apps {
			appNames = appNames + ", " + app.Name
		}
		appNames = strings.TrimPrefix(appNames, ", ")
		table.Print([][]string{{
			route.Host,
			route.Domain.Name,
			appNames,
		}})
		noRoutes = false
		return true
	})

	if apiErr != nil {
		cmd.ui.Failed("Failed fetching routes.\n%s", apiErr.Error())
		return
	}

	if noRoutes {
		cmd.ui.Say("No routes found")
	}
}
