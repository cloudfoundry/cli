package route

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
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

func (command ListRoutes) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "routes",
		ShortName:   "r",
		Description: "List all routes in the current space",
		Usage:       "CF_NAME routes",
	}
}

func (cmd ListRoutes) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) ([]requirements.Requirement, error) {
	return []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
	}, nil
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
		table.Add([]string{
			route.Host,
			route.Domain.Name,
			appNames,
		})
		noRoutes = false
		return true
	})
	table.Print()

	if apiErr != nil {
		cmd.ui.Failed("Failed fetching routes.\n%s", apiErr.Error())
		return
	}

	if noRoutes {
		cmd.ui.Say("No routes found")
	}
}
