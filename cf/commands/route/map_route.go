package route

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/simonleung8/flags"
	"github.com/simonleung8/flags/flag"
)

type MapRoute struct {
	ui           terminal.UI
	config       core_config.Reader
	routeRepo    api.RouteRepository
	appReq       requirements.ApplicationRequirement
	domainReq    requirements.DomainRequirement
	routeCreator RouteCreator
}

func init() {
	command_registry.Register(&MapRoute{})
}

func (cmd *MapRoute) MetaData() command_registry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["n"] = &cliFlags.StringFlag{Name: "n", Usage: T("Hostname")}

	return command_registry.CommandMetadata{
		Name:        "map-route",
		Description: T("Add a url route to an app"),
		Usage:       T("CF_NAME map-route APP_NAME DOMAIN [-n HOSTNAME]"),
		Flags:       fs,
	}
}

func (cmd *MapRoute) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(fc.Args()) != 2 {
		cmd.ui.Failed(T("Incorrect Usage. Requires APP and DOMAIN as arguments\n\n") + command_registry.Commands.CommandUsage("map-route"))
	}

	domainName := fc.Args()[1]

	cmd.appReq = requirementsFactory.NewApplicationRequirement(fc.Args()[0])

	cmd.domainReq = requirementsFactory.NewDomainRequirement(domainName)

	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		cmd.appReq,
		cmd.domainReq,
	}
	return
}

func (cmd *MapRoute) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.routeRepo = deps.RepoLocator.GetRouteRepository()

	//get create-route for dependency
	createRoute := command_registry.Commands.FindCommand("create-route")
	createRoute = createRoute.SetDependency(deps, false)
	cmd.routeCreator = createRoute.(RouteCreator)

	return cmd
}

func (cmd *MapRoute) Execute(c flags.FlagContext) {
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
