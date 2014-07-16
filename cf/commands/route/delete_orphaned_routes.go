package route

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
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

func (cmd DeleteOrphanedRoutes) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "delete-orphaned-routes",
		Description: T("Delete all orphaned routes (e.g.: those that are not mapped to an app)"),
		Usage:       T("CF_NAME delete-orphaned-routes [-f]"),
		Flags: []cli.Flag{
			cli.BoolFlag{Name: "f", Usage: T("Force deletion without confirmation")},
		},
	}
}

func (cmd DeleteOrphanedRoutes) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	reqs = append(reqs, reqFactory.NewLoginRequirement())
	return
}

func (cmd DeleteOrphanedRoutes) Run(c *cli.Context) {

	force := c.Bool("f")
	if !force {
		response := cmd.ui.Confirm(T("Really delete orphaned routes?{{.Prompt}}",
			map[string]interface{}{"Prompt": terminal.PromptColor(">")}))

		if !response {
			return
		}
	}

	cmd.ui.Say(T("Getting routes as {{.Username}} ...\n",
		map[string]interface{}{"Username": terminal.EntityNameColor(cmd.config.Username())}))

	apiErr := cmd.routeRepo.ListRoutes(func(route models.Route) bool {

		if len(route.Apps) == 0 {
			cmd.ui.Say(T("Deleting route {{.Route}}...",
				map[string]interface{}{"Route": terminal.EntityNameColor(route.Host + "." + route.Domain.Name)}))
			apiErr := cmd.routeRepo.Delete(route.Guid)
			if apiErr != nil {
				cmd.ui.Failed(apiErr.Error())
				return false
			}
		}
		return true
	})

	if apiErr != nil {
		cmd.ui.Failed(T("Failed fetching routes.\n{{.Err}}", map[string]interface{}{"Err": apiErr.Error()}))
		return
	}
	cmd.ui.Ok()
}
