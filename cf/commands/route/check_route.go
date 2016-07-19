package route

import (
	"fmt"
	"strings"

	"code.cloudfoundry.org/cli/cf"
	"code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/flags"
	. "code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
)

type CheckRoute struct {
	ui         terminal.UI
	config     coreconfig.Reader
	routeRepo  api.RouteRepository
	domainRepo api.DomainRepository
}

func init() {
	commandregistry.Register(&CheckRoute{})
}

func (cmd *CheckRoute) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["path"] = &flags.StringFlag{Name: "path", Usage: T("Path for the route")}

	return commandregistry.CommandMetadata{
		Name:        "check-route",
		Description: T("Perform a simple check to determine whether a route currently exists or not"),
		Usage: []string{
			T("CF_NAME check-route HOST DOMAIN [--path PATH]"),
		},
		Examples: []string{
			"CF_NAME check-route myhost example.com            # example.com",
			"CF_NAME check-route myhost example.com --path foo # myhost.example.com/foo",
		},
		Flags: fs,
	}
}

func (cmd *CheckRoute) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 2 {
		cmd.ui.Failed(T("Incorrect Usage. Requires host and domain as arguments\n\n") + commandregistry.Commands.CommandUsage("check-route"))
		return nil, fmt.Errorf("Incorrect usage: %d arguments of %d required", len(fc.Args()), 2)
	}

	var reqs []requirements.Requirement

	if fc.String("path") != "" {
		reqs = append(reqs, requirementsFactory.NewMinAPIVersionRequirement("Option '--path'", cf.RoutePathMinimumAPIVersion))
	}

	reqs = append(reqs, []requirements.Requirement{
		requirementsFactory.NewTargetedOrgRequirement(),
		requirementsFactory.NewLoginRequirement(),
	}...)

	return reqs, nil
}

func (cmd *CheckRoute) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.routeRepo = deps.RepoLocator.GetRouteRepository()
	cmd.domainRepo = deps.RepoLocator.GetDomainRepository()
	return cmd
}

func (cmd *CheckRoute) Execute(c flags.FlagContext) error {
	hostName := c.Args()[0]
	domainName := c.Args()[1]
	path := c.String("path")

	cmd.ui.Say(T("Checking for route..."))

	exists, err := cmd.CheckRoute(hostName, domainName, path)
	if err != nil {
		return err
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
	return nil
}

func (cmd *CheckRoute) CheckRoute(hostName, domainName, path string) (bool, error) {
	orgGUID := cmd.config.OrganizationFields().GUID
	domain, err := cmd.domainRepo.FindByNameInOrg(domainName, orgGUID)
	if err != nil {
		return false, err
	}

	found, err := cmd.routeRepo.CheckIfExists(hostName, domain, path)
	if err != nil {
		return false, err
	}

	return found, nil
}
