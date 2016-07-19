package route

import (
	"fmt"

	"code.cloudfoundry.org/cli/cf"
	"code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/flags"
	. "code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
)

//go:generate counterfeiter . Creator

type Creator interface {
	CreateRoute(hostName string, path string, port int, randomPort bool, domain models.DomainFields, space models.SpaceFields) (route models.Route, apiErr error)
}

type CreateRoute struct {
	ui        terminal.UI
	config    coreconfig.Reader
	routeRepo api.RouteRepository
	spaceReq  requirements.SpaceRequirement
	domainReq requirements.DomainRequirement
}

func init() {
	commandregistry.Register(&CreateRoute{})
}

func (cmd *CreateRoute) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["hostname"] = &flags.StringFlag{Name: "hostname", ShortName: "n", Usage: T("Hostname for the HTTP route (required for shared domains)")}
	fs["path"] = &flags.StringFlag{Name: "path", Usage: T("Path for the HTTP route")}
	fs["port"] = &flags.IntFlag{Name: "port", Usage: T("Port for the TCP route")}
	fs["random-port"] = &flags.BoolFlag{Name: "random-port", Usage: T("Create a random port for the TCP route")}

	return commandregistry.CommandMetadata{
		Name:        "create-route",
		Description: T("Create a url route in a space for later use"),
		Usage: []string{
			fmt.Sprintf("%s:\n", T("Create an HTTP route")),
			"      CF_NAME create-route ",
			fmt.Sprintf("%s ", T("SPACE")),
			fmt.Sprintf("%s ", T("DOMAIN")),
			fmt.Sprintf("[--hostname %s] ", T("HOSTNAME")),
			fmt.Sprintf("[--path %s]\n\n", T("PATH")),
			fmt.Sprintf("   %s:\n", T("Create a TCP route")),
			"      CF_NAME create-route ",
			fmt.Sprintf("%s ", T("SPACE")),
			fmt.Sprintf("%s ", T("DOMAIN")),
			fmt.Sprintf("(--port %s | --random-port)", T("PORT")),
		},
		Examples: []string{
			"CF_NAME create-route my-space example.com                             # example.com",
			"CF_NAME create-route my-space example.com --hostname myapp            # myapp.example.com",
			"CF_NAME create-route my-space example.com --hostname myapp --path foo # myapp.example.com/foo",
			"CF_NAME create-route my-space example.com --port 50000                # example.com:50000",
		},
		Flags: fs,
	}
}

func (cmd *CreateRoute) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 2 {
		cmd.ui.Failed(T("Incorrect Usage. Requires SPACE and DOMAIN as arguments\n\n") + commandregistry.Commands.CommandUsage("create-route"))
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

	domainName := fc.Args()[1]

	cmd.spaceReq = requirementsFactory.NewSpaceRequirement(fc.Args()[0])
	cmd.domainReq = requirementsFactory.NewDomainRequirement(domainName)

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedOrgRequirement(),
		cmd.spaceReq,
		cmd.domainReq,
	}

	if fc.IsSet("path") {
		reqs = append(reqs, requirementsFactory.NewMinAPIVersionRequirement("Option '--path'", cf.RoutePathMinimumAPIVersion))
	}

	if fc.IsSet("port") {
		reqs = append(reqs, requirementsFactory.NewMinAPIVersionRequirement("Option '--port'", cf.TCPRoutingMinimumAPIVersion))
	}

	if fc.IsSet("random-port") {
		reqs = append(reqs, requirementsFactory.NewMinAPIVersionRequirement("Option '--random-port'", cf.TCPRoutingMinimumAPIVersion))
	}

	return reqs, nil
}

func (cmd *CreateRoute) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.routeRepo = deps.RepoLocator.GetRouteRepository()
	return cmd
}

func (cmd *CreateRoute) Execute(c flags.FlagContext) error {
	hostName := c.String("n")
	space := cmd.spaceReq.GetSpace()
	domain := cmd.domainReq.GetDomain()
	path := c.String("path")
	port := c.Int("port")
	randomPort := c.Bool("random-port")

	_, err := cmd.CreateRoute(hostName, path, port, randomPort, domain, space.SpaceFields)
	if err != nil {
		return err
	}

	return nil
}

func (cmd *CreateRoute) CreateRoute(hostName string, path string, port int, randomPort bool, domain models.DomainFields, space models.SpaceFields) (models.Route, error) {
	cmd.ui.Say(T("Creating route {{.URL}} for org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...",
		map[string]interface{}{
			"URL":       terminal.EntityNameColor(domain.URLForHostAndPath(hostName, path, port)),
			"OrgName":   terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			"SpaceName": terminal.EntityNameColor(space.Name),
			"Username":  terminal.EntityNameColor(cmd.config.Username())}))

	route, err := cmd.routeRepo.CreateInSpace(hostName, path, domain.GUID, space.GUID, port, randomPort)
	if err != nil {
		var findErr error
		route, findErr = cmd.routeRepo.Find(hostName, domain, path, port)
		if findErr != nil {
			return models.Route{}, err
		}

		if route.Space.GUID != space.GUID || route.Domain.GUID != domain.GUID {
			return models.Route{}, err
		}

		cmd.ui.Ok()
		cmd.ui.Warn(T("Route {{.URL}} already exists",
			map[string]interface{}{"URL": route.URL()}))

		return route, nil
	}

	cmd.ui.Ok()
	if randomPort {
		cmd.ui.Say("Route %s:%d has been created", route.Domain.Name, route.Port)
	}

	return route, nil
}
