package route

import (
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

type UnmapRoute struct {
	ui        terminal.UI
	config    coreconfig.Reader
	routeRepo api.RouteRepository
	appReq    requirements.ApplicationRequirement
	domainReq requirements.DomainRequirement
}

func init() {
	commandregistry.Register(&UnmapRoute{})
}

func (cmd *UnmapRoute) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["hostname"] = &flags.StringFlag{Name: "hostname", ShortName: "n", Usage: T("Hostname used to identify the HTTP route")}
	fs["path"] = &flags.StringFlag{Name: "path", Usage: T("Path used to identify the HTTP route")}
	fs["port"] = &flags.IntFlag{Name: "port", Usage: T("Port used to identify the TCP route")}

	return commandregistry.CommandMetadata{
		Name:        "unmap-route",
		Description: T("Remove a url route from an app"),
		Usage: []string{
			fmt.Sprintf("%s:\n", T("Unmap an HTTP route")),
			"      CF_NAME unmap-route ",
			fmt.Sprintf("%s ", T("APP_NAME")),
			fmt.Sprintf("%s ", T("DOMAIN")),
			fmt.Sprintf("[--hostname %s] ", T("HOSTNAME")),
			fmt.Sprintf("[--path %s]\n\n", T("PATH")),
			fmt.Sprintf("   %s:\n", T("Unmap a TCP route")),
			"      CF_NAME unmap-route ",
			fmt.Sprintf("%s ", T("APP_NAME")),
			fmt.Sprintf("%s ", T("DOMAIN")),
			fmt.Sprintf("--port %s", T("PORT")),
		},
		Examples: []string{
			"CF_NAME unmap-route my-app example.com                              # example.com",
			"CF_NAME unmap-route my-app example.com --hostname myhost            # myhost.example.com",
			"CF_NAME unmap-route my-app example.com --hostname myhost --path foo # myhost.example.com/foo",
			"CF_NAME unmap-route my-app example.com --port 5000                  # example.com:5000",
		},
		Flags: fs,
	}
}

func (cmd *UnmapRoute) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 2 {
		cmd.ui.Failed(T("Incorrect Usage. Requires app_name, domain_name as arguments\n\n") + commandregistry.Commands.CommandUsage("unmap-route"))
		return nil, fmt.Errorf("Incorrect usage: %d arguments of %d required", len(fc.Args()), 2)
	}

	if fc.IsSet("port") && (fc.IsSet("hostname") || fc.IsSet("path")) {
		cmd.ui.Failed(T("Cannot specify port together with hostname and/or path."))
		return nil, fmt.Errorf("Cannot specify port together with hostname and/or path.")
	}

	domainName := fc.Args()[1]

	cmd.appReq = requirementsFactory.NewApplicationRequirement(fc.Args()[0])
	cmd.domainReq = requirementsFactory.NewDomainRequirement(domainName)

	var reqs []requirements.Requirement

	if fc.String("path") != "" {
		reqs = append(reqs, requirementsFactory.NewMinAPIVersionRequirement("Option '--path'", cf.RoutePathMinimumAPIVersion))
	}

	if fc.IsSet("port") {
		reqs = append(reqs, requirementsFactory.NewMinAPIVersionRequirement("Option '--port'", cf.TCPRoutingMinimumAPIVersion))
	}

	reqs = append(reqs, []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		cmd.appReq,
		cmd.domainReq,
	}...)

	return reqs, nil
}

func (cmd *UnmapRoute) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.routeRepo = deps.RepoLocator.GetRouteRepository()
	return cmd
}

func (cmd *UnmapRoute) Execute(c flags.FlagContext) error {
	hostName := c.String("n")
	path := c.String("path")
	port := c.Int("port")
	domain := cmd.domainReq.GetDomain()
	app := cmd.appReq.GetApplication()

	route, err := cmd.routeRepo.Find(hostName, domain, path, port)
	if err != nil {
		return err
	}

	cmd.ui.Say(T("Removing route {{.URL}} from app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...",
		map[string]interface{}{
			"URL":       terminal.EntityNameColor(route.URL()),
			"AppName":   terminal.EntityNameColor(app.Name),
			"OrgName":   terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			"SpaceName": terminal.EntityNameColor(cmd.config.SpaceFields().Name),
			"Username":  terminal.EntityNameColor(cmd.config.Username())}))

	var routeFound bool
	for _, routeApp := range route.Apps {
		if routeApp.GUID == app.GUID {
			routeFound = true
			err = cmd.routeRepo.Unbind(route.GUID, app.GUID)
			if err != nil {
				return err
			}
			break
		}
	}

	cmd.ui.Ok()

	if !routeFound {
		cmd.ui.Warn(T("\nRoute to be unmapped is not currently mapped to the application."))
	}
	return nil
}
