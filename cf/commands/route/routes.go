package route

import (
	. "github.com/cloudfoundry/cli/cf/i18n"
	"strings"

	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
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

func (cmd ListRoutes) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "routes",
		ShortName:   "r",
		Description: T("List all routes in the current space"),
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
	cmd.ui.Say(T("Getting routes as {{.Username}} ...\n",
		map[string]interface{}{"Username": terminal.EntityNameColor(cmd.config.Username())}))

	table := cmd.ui.Table([]string{T("host"), T("domain"), T("apps")})

	noRoutes := true
	apiErr := cmd.routeRepo.ListRoutes(func(route models.Route) bool {
		noRoutes = false
		appNames := []string{}
		for _, app := range route.Apps {
			appNames = append(appNames, app.Name)
		}

		table.Add(route.Host, route.Domain.Name, strings.Join(appNames, ","))
		return true
	})
	table.Print()

	if apiErr != nil {
		cmd.ui.Failed(T("Failed fetching routes.\n{{.Err}}", map[string]interface{}{"Err": apiErr.Error()}))
		return
	}

	if noRoutes {
		cmd.ui.Say(T("No routes found"))
	}
}
