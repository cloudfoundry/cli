package route

import (
	"strings"

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
	fs := make(map[string]flags.FlagSet)
	fs["path"] = &cliFlags.StringFlag{Name: "path", Usage: T("Path for the route")}

	return command_registry.CommandMetadata{
		Name:        "check-route",
		Description: T("Perform a simple check to determine whether a route currently exists or not."),
		Usage: T(`CF_NAME check-route HOST DOMAIN [--path PATH]

EXAMPLES:
   CF_NAME check-route myhost example.com            # example.com
   CF_NAME check-route myhost example.com --path foo # myhost.example.com/foo`),
		Flags: fs,
	}
}

func (cmd *CheckRoute) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 2 {
		cmd.ui.Failed(T("Incorrect Usage. Requires host and domain as arguments\n\n") + command_registry.Commands.CommandUsage("check-route"))
	}

	var reqs []requirements.Requirement

	if fc.String("path") != "" {
		requiredVersion, err := semver.Make("2.36.0")
		if err != nil {
			panic(err.Error())
		}
		reqs = append(reqs, requirementsFactory.NewMinAPIVersionRequirement("Option '--path'", requiredVersion))
	}

	reqs = append(reqs, []requirements.Requirement{
		requirementsFactory.NewTargetedOrgRequirement(),
		requirementsFactory.NewLoginRequirement(),
	}...)

	return reqs, nil
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
	path := c.String("path")

	cmd.ui.Say(T("Checking for route..."))

	exists, err := cmd.CheckRoute(hostName, domainName, path)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	cmd.ui.Ok()

	var existence string
	if exists {
		existence = "does exist"
	} else {
		existence = "does not exist"
	}

	if path != "" {
		cmd.ui.Say(T("Route {{.HostName}}.{{.DomainName}}/{{.Path}} {{.Existence}}",
			map[string]interface{}{
				"HostName":   hostName,
				"DomainName": domainName,
				"Existence":  existence,
				"Path":       strings.TrimPrefix(path, `/`),
			},
		))
	} else {
		cmd.ui.Say(T("Route {{.HostName}}.{{.DomainName}} {{.Existence}}",
			map[string]interface{}{
				"HostName":   hostName,
				"DomainName": domainName,
				"Existence":  existence,
			},
		))
	}
}

func (cmd *CheckRoute) CheckRoute(hostName, domainName, path string) (bool, error) {
	orgGuid := cmd.config.OrganizationFields().Guid
	domain, err := cmd.domainRepo.FindByNameInOrg(domainName, orgGuid)
	if err != nil {
		return false, err
	}

	found, err := cmd.routeRepo.CheckIfExists(hostName, domain, path)
	if err != nil {
		return false, err
	}

	return found, nil
}
