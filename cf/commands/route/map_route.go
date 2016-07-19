package route

import (
	"errors"
	"fmt"

	"code.cloudfoundry.org/cli/cf"
	"code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/flags"
	. "code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
)

type MapRoute struct {
	ui           terminal.UI
	config       coreconfig.Reader
	routeRepo    api.RouteRepository
	appReq       requirements.ApplicationRequirement
	domainReq    requirements.DomainRequirement
	routeCreator Creator
}

func init() {
	commandregistry.Register(&MapRoute{})
}

func (cmd *MapRoute) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["hostname"] = &flags.StringFlag{Name: "hostname", ShortName: "n", Usage: T("Hostname for the HTTP route (required for shared domains)")}
	fs["path"] = &flags.StringFlag{Name: "path", Usage: T("Path for the HTTP route")}
	fs["port"] = &flags.IntFlag{Name: "port", Usage: T("Port for the TCP route")}
	fs["random-port"] = &flags.BoolFlag{Name: "random-port", Usage: T("Create a random port for the TCP route")}

	return commandregistry.CommandMetadata{
		Name:        "map-route",
		Description: T("Add a url route to an app"),
		Usage: []string{
			fmt.Sprintf("%s:\n", T("Map an HTTP route")),
			"      CF_NAME map-route ",
			fmt.Sprintf("%s ", T("APP_NAME")),
			fmt.Sprintf("%s ", T("DOMAIN")),
			fmt.Sprintf("[--hostname %s] ", T("HOSTNAME")),
			fmt.Sprintf("[--path %s]\n\n", T("PATH")),
			fmt.Sprintf("   %s:\n", T("Map a TCP route")),
			"      CF_NAME map-route ",
			fmt.Sprintf("%s ", T("APP_NAME")),
			fmt.Sprintf("%s ", T("DOMAIN")),
			fmt.Sprintf("(--port %s | --random-port)", T("PORT")),
		},
		Examples: []string{
			"CF_NAME map-route my-app example.com                              # example.com",
			"CF_NAME map-route my-app example.com --hostname myhost            # myhost.example.com",
			"CF_NAME map-route my-app example.com --hostname myhost --path foo # myhost.example.com/foo",
			"CF_NAME map-route my-app example.com --port 50000                 # example.com:50000",
		},
		Flags: fs,
	}
}

func (cmd *MapRoute) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 2 {
		cmd.ui.Failed(T("Incorrect Usage. Requires APP_NAME and DOMAIN as arguments\n\n") + commandregistry.Commands.CommandUsage("map-route"))
		return nil, fmt.Errorf("Incorrect usage: %d arguments of %d required", len(fc.Args()), 2)
	}

	if fc.IsSet("port") && (fc.IsSet("hostname") || fc.IsSet("path")) {
		cmd.ui.Failed(T("Cannot specify port together with hostname and/or path."))
		return nil, fmt.Errorf("Cannot specify port together with hostname and/or path.")
	}

	if fc.IsSet("random-port") && (fc.IsSet("port") || fc.IsSet("hostname") || fc.IsSet("path")) {
		cmd.ui.Failed(T("Cannot specify random-port together with port, hostname and/or path."))
		return nil, fmt.Errorf("Cannot specify random-port together with port, hostname and/or path.")
	}

	appName := fc.Args()[0]
	domainName := fc.Args()[1]

	requirement := requirementsFactory.NewApplicationRequirement(appName)
	cmd.appReq = requirement

	cmd.domainReq = requirementsFactory.NewDomainRequirement(domainName)

	var reqs []requirements.Requirement

	if fc.String("path") != "" {
		reqs = append(reqs, requirementsFactory.NewMinAPIVersionRequirement("Option '--path'", cf.RoutePathMinimumAPIVersion))
	}

	var flag string
	switch {
	case fc.IsSet("port"):
		flag = "port"
	case fc.IsSet("random-port"):
		flag = "random-port"
	}

	if flag != "" {
		reqs = append(reqs, requirementsFactory.NewMinAPIVersionRequirement(fmt.Sprintf("Option '--%s'", flag), cf.TCPRoutingMinimumAPIVersion))
		reqs = append(reqs, requirementsFactory.NewDiegoApplicationRequirement(appName))
	}

	reqs = append(reqs, []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		cmd.appReq,
		cmd.domainReq,
	}...)

	return reqs, nil
}

func (cmd *MapRoute) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.routeRepo = deps.RepoLocator.GetRouteRepository()

	//get create-route for dependency
	createRoute := commandregistry.Commands.FindCommand("create-route")
	createRoute = createRoute.SetDependency(deps, false)
	cmd.routeCreator = createRoute.(Creator)

	return cmd
}

func (cmd *MapRoute) Execute(c flags.FlagContext) error {
	hostName := c.String("n")
	path := c.String("path")
	domain := cmd.domainReq.GetDomain()
	app := cmd.appReq.GetApplication()

	port := c.Int("port")
	randomPort := c.Bool("random-port")
	route, err := cmd.routeCreator.CreateRoute(hostName, path, port, randomPort, domain, cmd.config.SpaceFields())
	if err != nil {
		return errors.New(T("Error resolving route:\n{{.Err}}", map[string]interface{}{"Err": err.Error()}))
	}
	cmd.ui.Say(T("Adding route {{.URL}} to app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...",
		map[string]interface{}{
			"URL":       terminal.EntityNameColor(route.URL()),
			"AppName":   terminal.EntityNameColor(app.Name),
			"OrgName":   terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			"SpaceName": terminal.EntityNameColor(cmd.config.SpaceFields().Name),
			"Username":  terminal.EntityNameColor(cmd.config.Username())}))

	err = cmd.routeRepo.Bind(route.GUID, app.GUID)
	if err != nil {
		return err
	}

	cmd.ui.Ok()
	return nil
}
