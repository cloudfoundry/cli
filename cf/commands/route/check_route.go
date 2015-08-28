package route

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/simonleung8/flags"
)

type CheckRoute struct {
	ui         terminal.UI
	config     core_config.Reader
	routeRepo  api.RouteRepository
	domainRepo api.DomainRepository
}

func init() {
	command_registry.Register(&CheckRoute{})
}

func (cmd *CheckRoute) MetaData() command_registry.CommandMetadata {
	return command_registry.CommandMetadata{
		Name:        "check-route",
		Description: T("Perform a simple check to determine whether a route currently exists or not."),
		Usage:       T("CF_NAME check-route HOST DOMAIN"),
	}
}

func (cmd *CheckRoute) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(fc.Args()) != 2 {
		cmd.ui.Failed(T("Incorrect Usage. Requires host and domain as arguments\n\n") + command_registry.Commands.CommandUsage("check-route"))
	}

	reqs = []requirements.Requirement{
		requirementsFactory.NewTargetedOrgRequirement(),
		requirementsFactory.NewLoginRequirement(),
	}
	return
}

func (cmd *CheckRoute) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.routeRepo = deps.RepoLocator.GetRouteRepository()
	cmd.domainRepo = deps.RepoLocator.GetDomainRepository()
	return cmd
}

func (cmd *CheckRoute) Execute(c flags.FlagContext) {
	hostName := c.Args()[0]
	domainName := c.Args()[1]

	cmd.ui.Say(T("Checking for route..."))

	exists, err := cmd.CheckRoute(hostName, domainName)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	cmd.ui.Ok()

	if exists {
		cmd.ui.Say(T("Route {{.HostName}}.{{.DomainName}} does exist",
			map[string]interface{}{"HostName": hostName, "DomainName": domainName},
		))
	} else {
		cmd.ui.Say(T("Route {{.HostName}}.{{.DomainName}} does not exist",
			map[string]interface{}{"HostName": hostName, "DomainName": domainName},
		))
	}
}

func (cmd *CheckRoute) CheckRoute(hostName, domainName string) (bool, error) {
	orgGuid := cmd.config.OrganizationFields().Guid
	domain, err := cmd.domainRepo.FindByNameInOrg(domainName, orgGuid)
	if err != nil {
		return false, err
	}

	found, err := cmd.routeRepo.CheckIfExists(hostName, domain)
	if err != nil {
		return false, err
	}

	return found, nil
}
