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
	fs["n"] = &cliFlags.StringFlag{Name: "n", Usage: T("Hostname")}

	return command_registry.CommandMetadata{
		Name:        "unmap-route",
		Description: T("Remove a url route from an app"),
		Usage:       T("CF_NAME unmap-route APP_NAME DOMAIN [-n HOSTNAME]"),
		Flags:       fs,
	}
}

func (cmd *UnmapRoute) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(fc.Args()) != 2 {
		cmd.ui.Failed(T("Incorrect Usage. Requires app_name, domain_name as arguments\n\n") + command_registry.Commands.CommandUsage("unmap-route"))
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

func (cmd *UnmapRoute) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.routeRepo = deps.RepoLocator.GetRouteRepository()
	return cmd
}

func (cmd *UnmapRoute) Execute(c flags.FlagContext) {
	hostName := c.String("n")
	domain := cmd.domainReq.GetDomain()
	app := cmd.appReq.GetApplication()

	route, apiErr := cmd.routeRepo.FindByHostAndDomain(hostName, domain)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
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
			apiErr = cmd.routeRepo.Unbind(route.Guid, app.Guid)
			if apiErr != nil {
				cmd.ui.Failed(apiErr.Error())
				return
			}
		}
	}
	cmd.ui.Ok()

	if !routeFound {
		cmd.ui.Warn(T("\nRoute to be unmapped is not currently mapped to the application."))
	}

}
