package route

import (
	"strings"

	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/flags"
	"github.com/cloudfoundry/cli/flags/flag"

	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type ListRoutes struct {
	ui        terminal.UI
	routeRepo api.RouteRepository
	config    core_config.Reader
}

func init() {
	command_registry.Register(&ListRoutes{})
}

func (cmd *ListRoutes) MetaData() command_registry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["orglevel"] = &cliFlags.BoolFlag{Name: "orglevel", Usage: T("List all the routes for all spaces of current organization")}

	return command_registry.CommandMetadata{
		Name:        "routes",
		ShortName:   "r",
		Description: T("List all routes in the current space or the current organization"),
		Usage:       "CF_NAME routes",
		Flags:       fs,
	}
}

func (cmd *ListRoutes) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 0 {
		cmd.ui.Failed(T("Incorrect Usage. No argument required\n\n") + command_registry.Commands.CommandUsage("routes"))
	}

	return []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
	}, nil
}

func (cmd *ListRoutes) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.routeRepo = deps.RepoLocator.GetRouteRepository()
	return cmd
}

func (cmd *ListRoutes) Execute(c flags.FlagContext) {
	cmd.ui.Say(T("Getting routes as {{.Username}} ...\n",
		map[string]interface{}{"Username": terminal.EntityNameColor(cmd.config.Username())}))

	table := cmd.ui.Table([]string{T("space"), T("host"), T("domain"), T("path"), T("apps")})

	noRoutes := true
	var apiErr error
	flag := c.Bool("orglevel")

	if flag {
		apiErr = cmd.routeRepo.ListAllRoutes(func(route models.Route) bool {
			noRoutes = false
			appNames := []string{}
			for _, app := range route.Apps {
				appNames = append(appNames, app.Name)
			}

			table.Add(route.Space.Name, route.Host, route.Domain.Name, route.Path, strings.Join(appNames, ","))
			return true
		})

	} else {

		apiErr = cmd.routeRepo.ListRoutes(func(route models.Route) bool {
			noRoutes = false
			appNames := []string{}
			for _, app := range route.Apps {
				appNames = append(appNames, app.Name)
			}

			table.Add(route.Space.Name, route.Host, route.Domain.Name, route.Path, strings.Join(appNames, ","))
			return true
		})
	}

	table.Print()

	if apiErr != nil {
		cmd.ui.Failed(T("Failed fetching routes.\n{{.Err}}", map[string]interface{}{"Err": apiErr.Error()}))
		return
	}

	if noRoutes {
		cmd.ui.Say(T("No routes found"))
	}
}
