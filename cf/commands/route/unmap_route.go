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

type UnmapRoute struct {
	ui        terminal.UI
	config    core_config.Reader
	routeRepo api.RouteRepository
	appReq    requirements.ApplicationRequirement
	domainReq requirements.DomainRequirement
}

func init() {
	command_registry.Register(&UnmapRoute{})
}

func (cmd *UnmapRoute) MetaData() command_registry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["hostname"] = &cliFlags.StringFlag{Name: "hostname", ShortName: "n", Usage: T("Hostname used to identify the route")}
	fs["path"] = &cliFlags.StringFlag{Name: "path", Usage: T("Path used to identify the route")}

	return command_registry.CommandMetadata{
		Name:        "unmap-route",
		Description: T("Remove a url route from an app"),
		Usage: T(`CF_NAME unmap-route APP_NAME DOMAIN [--hostname HOSTNAME] [--path PATH]

EXAMPLES:
   CF_NAME unmap-route my-app example.com                              # example.com
   CF_NAME unmap-route my-app example.com --hostname myhost            # myhost.example.com
   CF_NAME unmap-route my-app example.com --hostname myhost --path foo # myhost.example.com/foo`),
		Flags: fs,
	}
}

func (cmd *UnmapRoute) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 2 {
		cmd.ui.Failed(T("Incorrect Usage. Requires app_name, domain_name as arguments\n\n") + command_registry.Commands.CommandUsage("unmap-route"))
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

func (cmd *UnmapRoute) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.routeRepo = deps.RepoLocator.GetRouteRepository()
	return cmd
}

func (cmd *UnmapRoute) Execute(c flags.FlagContext) {
	hostName := c.String("n")
	path := c.String("path")
	domain := cmd.domainReq.GetDomain()
	app := cmd.appReq.GetApplication()

	route, err := cmd.routeRepo.Find(hostName, domain, path)
	if err != nil {
		cmd.ui.Failed(err.Error())
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
		if routeApp.Guid == app.Guid {
			routeFound = true
			err = cmd.routeRepo.Unbind(route.Guid, app.Guid)
			if err != nil {
				cmd.ui.Failed(err.Error())
			}
			break
		}
	}

	cmd.ui.Ok()

	if !routeFound {
		cmd.ui.Warn(T("\nRoute to be unmapped is not currently mapped to the application."))
	}
}
