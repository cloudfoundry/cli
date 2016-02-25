package route

import (
	"github.com/blang/semver"
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
)

//go:generate counterfeiter -o fakes/fake_route_creator.go . RouteCreator
type RouteCreator interface {
	CreateRoute(hostName string, path string, port int, domain models.DomainFields, space models.SpaceFields) (route models.Route, apiErr error)
}

type CreateRoute struct {
	ui        terminal.UI
	config    core_config.Reader
	routeRepo api.RouteRepository
	spaceReq  requirements.SpaceRequirement
	domainReq requirements.DomainRequirement
}

func init() {
	command_registry.Register(&CreateRoute{})
}

func (cmd *CreateRoute) MetaData() command_registry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["hostname"] = &flags.StringFlag{Name: "hostname", ShortName: "n", Usage: T("Hostname for the HTTP route (required for shared domains)")}
	fs["path"] = &flags.StringFlag{Name: "path", Usage: T("Path for the HTTP route")}
	fs["port"] = &flags.IntFlag{Name: "port", Usage: T("Port for the TCP route")}

	return command_registry.CommandMetadata{
		Name:        "create-route",
		Description: T("Create a url route in a space for later use"),
		Usage: T(`CF_NAME create-route SPACE DOMAIN [--hostname HOSTNAME] [--path PATH] [--port PORT]

EXAMPLES:
   CF_NAME create-route my-space example.com                             # example.com
   CF_NAME create-route my-space example.com --hostname myapp            # myapp.example.com
   CF_NAME create-route my-space example.com --hostname myapp --path foo # myapp.example.com/foo
   CF_NAME create-route my-space example.com --port 50000                # example.com:50000`),
		Flags: fs,
	}
}

func (cmd *CreateRoute) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 2 {
		cmd.ui.Failed(T("Incorrect Usage. Requires SPACE and DOMAIN as arguments\n\n") + command_registry.Commands.CommandUsage("create-route"))
	}

	if fc.IsSet("port") && (fc.IsSet("hostname") || fc.IsSet("path")) {
		cmd.ui.Failed(T("Cannot specify port together with hostname and/or path."))
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
		requiredVersion, err := semver.Make("2.36.0")
		if err != nil {
			panic(err.Error())
		}

		reqs = append(reqs, requirementsFactory.NewMinAPIVersionRequirement("Option '--path'", requiredVersion))
	}

	if fc.IsSet("port") {
		requiredVersion, err := semver.Make("2.51.0")
		if err != nil {
			panic(err.Error())
		}

		reqs = append(reqs, requirementsFactory.NewMinAPIVersionRequirement("Option '--port'", requiredVersion))
	}

	return reqs, nil
}

func (cmd *CreateRoute) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.routeRepo = deps.RepoLocator.GetRouteRepository()
	return cmd
}

func (cmd *CreateRoute) Execute(c flags.FlagContext) {
	hostName := c.String("n")
	space := cmd.spaceReq.GetSpace()
	domain := cmd.domainReq.GetDomain()
	path := c.String("path")
	port := c.Int("port")

	_, apiErr := cmd.CreateRoute(hostName, path, port, domain, space.SpaceFields)

	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}
}

func (cmd *CreateRoute) CreateRoute(hostName string, path string, port int, domain models.DomainFields, space models.SpaceFields) (models.Route, error) {
	cmd.ui.Say(T("Creating route {{.URL}} for org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...",
		map[string]interface{}{
			"URL":       terminal.EntityNameColor(domain.UrlForHostAndPath(hostName, path)),
			"OrgName":   terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			"SpaceName": terminal.EntityNameColor(space.Name),
			"Username":  terminal.EntityNameColor(cmd.config.Username())}))

	route, err := cmd.routeRepo.CreateInSpace(hostName, path, domain.Guid, space.Guid, port)
	if err != nil {
		var findErr error
		route, findErr = cmd.routeRepo.Find(hostName, domain, path)
		if findErr != nil {
			return models.Route{}, err
		}

		if route.Space.Guid != space.Guid || route.Domain.Guid != domain.Guid {
			return models.Route{}, err
		}

		cmd.ui.Ok()
		cmd.ui.Warn(T("Route {{.URL}} already exists",
			map[string]interface{}{"URL": route.URL()}))

		return route, nil
	}

	cmd.ui.Ok()

	return route, nil
}
