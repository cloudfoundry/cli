package route

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/flag_helpers"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type MapRoute struct {
	ui           terminal.UI
	config       configuration.Reader
	routeRepo    api.RouteRepository
	appReq       requirements.ApplicationRequirement
	domainReq    requirements.DomainRequirement
	routeCreator RouteCreator
}

func NewMapRoute(ui terminal.UI, config configuration.Reader, routeRepo api.RouteRepository, routeCreator RouteCreator) (cmd *MapRoute) {
	cmd = new(MapRoute)
	cmd.ui = ui
	cmd.config = config
	cmd.routeRepo = routeRepo
	cmd.routeCreator = routeCreator
	return
}

func (cmd *MapRoute) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "map-route",
		Description: T("Add a url route to an app"),
		Usage:       T("CF_NAME map-route APP DOMAIN [-n HOSTNAME]"),
		Flags: []cli.Flag{
			flag_helpers.NewStringFlag("n", T("Hostname")),
		},
	}
}

func (cmd *MapRoute) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 2 {
		cmd.ui.FailWithUsage(c)
	}

	appName := c.Args()[0]
	domainName := c.Args()[1]

	cmd.appReq = requirementsFactory.NewApplicationRequirement(appName)
	cmd.domainReq = requirementsFactory.NewDomainRequirement(domainName)

	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		cmd.appReq,
		cmd.domainReq,
	}
	return
}

func (cmd *MapRoute) Run(c *cli.Context) {
	hostName := c.String("n")
	domain := cmd.domainReq.GetDomain()
	app := cmd.appReq.GetApplication()

	route, apiErr := cmd.routeCreator.CreateRoute(hostName, domain, cmd.config.SpaceFields())
	if apiErr != nil {
		cmd.ui.Failed(T("Error resolving route:\n{{.Err}}", map[string]interface{}{"Err": apiErr.Error()}))
	}
	cmd.ui.Say(T("Adding route {{.URL}} to app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...",
		map[string]interface{}{
			"URL":       terminal.EntityNameColor(route.URL()),
			"AppName":   terminal.EntityNameColor(app.Name),
			"OrgName":   terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			"SpaceName": terminal.EntityNameColor(cmd.config.SpaceFields().Name),
			"Username":  terminal.EntityNameColor(cmd.config.Username())}))

	apiErr = cmd.routeRepo.Bind(route.Guid, app.Guid)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Ok()
}
