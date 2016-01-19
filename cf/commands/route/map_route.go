package route

import (
	"github.com/blang/semver"
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
	"github.com/cloudfoundry/cli/flags/flag"
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
	fs["hostname"] = &cliFlags.StringFlag{Name: "hostname", ShortName: "n", Usage: T("Hostname for the route (required for shared domains)")}
	fs["path"] = &cliFlags.StringFlag{Name: "path", Usage: T("Path for the route")}

	return command_registry.CommandMetadata{
		Name:        "map-route",
		Description: T("Add a url route to an app"),
		Usage: T(`CF_NAME map-route APP_NAME DOMAIN [--hostname HOSTNAME] [--path PATH]

EXAMPLES:
   CF_NAME map-route my-app example.com                              # example.com
   CF_NAME map-route my-app example.com --hostname myhost            # myhost.example.com
   CF_NAME map-route my-app example.com --hostname myhost --path foo # myhost.example.com/foo`),
		Flags: fs,
	}
}

func (cmd *MapRoute) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 2 {
		cmd.ui.Failed(T("Incorrect Usage. Requires APP_NAME and DOMAIN as arguments\n\n") + command_registry.Commands.CommandUsage("map-route"))
	}

	domainName := fc.Args()[1]

	cmd.appReq = requirementsFactory.NewApplicationRequirement(fc.Args()[0])

	cmd.domainReq = requirementsFactory.NewDomainRequirement(domainName)

	requiredVersion, err := semver.Make("2.36.0")
	if err != nil {
		panic(err.Error())
	}

	var reqs []requirements.Requirement

	if fc.String("path") != "" {
		reqs = append(reqs, requirementsFactory.NewMinAPIVersionRequirement("Option '--path'", requiredVersion))
	}

	reqs = append(reqs, []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		cmd.appReq,
		cmd.domainReq,
	}...)

	return reqs, nil
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
	path := c.String("path")
	domain := cmd.domainReq.GetDomain()
	app := cmd.appReq.GetApplication()

	route, apiErr := cmd.routeCreator.CreateRoute(hostName, path, domain, cmd.config.SpaceFields())
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
