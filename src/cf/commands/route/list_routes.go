package route

import (
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	"cf/terminal"
	"github.com/codegangsta/cli"
)

type ListRoutes struct {
	ui        terminal.UI
	routeRepo api.RouteRepository
	config    *configuration.Configuration
}

func NewListRoutes(ui terminal.UI, config *configuration.Configuration, routeRepo api.RouteRepository) (cmd *ListRoutes) {
	cmd = new(ListRoutes)
	cmd.ui = ui
	cmd.config = config
	cmd.routeRepo = routeRepo
	return
}

func (cmd ListRoutes) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	return
}

func (cmd ListRoutes) Run(c *cli.Context) {
	cmd.ui.Say("Getting routes as %s ...\n",
		terminal.EntityNameColor(cmd.config.Username()),
	)

	stopChan := make(chan bool)
	defer close(stopChan)

	routesChan, statusChan := cmd.routeRepo.ListRoutes(stopChan)

	table := cmd.ui.Table([]string{"host", "domain", "apps"})
	noRoutes := true

	for routes := range routesChan {
		rows := [][]string{}
		for _, route := range routes {
			appNames := ""
			for _, app := range route.Apps {
				appNames = appNames + ", " + app.Name
			}
			rows = append(rows, []string{
				route.Host,
				route.Domain.Name,
				appNames,
			})
		}
		table.Print(rows)
		noRoutes = false
	}

	apiStatus := <-statusChan
	if apiStatus.IsNotSuccessful() {
		cmd.ui.Failed("Failed fetching routes.\n%s", apiStatus.Message)
		return
	}

	if noRoutes {
		cmd.ui.Say("No routes found")
	}
}
