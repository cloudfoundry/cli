package route

import (
	"github.com/blang/semver"
	"strconv"
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
	"github.com/cloudfoundry/cli/flags/flag"
)

//go:generate counterfeiter -o fakes/fake_route_creator.go . RouteCreator
type RouteCreator interface {
	CreateRoute(hostName string, port string, path string, domain models.DomainFields, space models.SpaceFields) (route models.Route, apiErr error)
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
	fs["hostname"] = &cliFlags.StringFlag{Name: "hostname", ShortName: "n", Usage: T("Hostname for the route (required for shared domains)")}
	fs["path"] = &cliFlags.StringFlag{Name: "path", Usage: T("Path for the route")}
	fs["port"] = &cliFlags.IntFlag{Name: "port", Usage: T("Port used to create a TCP route. E.g. where domain is example.com and port is 50000, the resulting route is example.com:50000.")}

	return command_registry.CommandMetadata{
		Name:        "create-route",
		Description: T("Create a url route in a space for later use"),
		Usage: T(`cf create-route SPACE DOMAIN [[--hostname HOSTNAME] [--path PATH] | --port PORT]

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

	domainName := fc.Args()[1]

	cmd.spaceReq = requirementsFactory.NewSpaceRequirement(fc.Args()[0])
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
		requirementsFactory.NewTargetedOrgRequirement(),
		cmd.spaceReq,
		cmd.domainReq,
	}...)

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

	portStr := ""
	if c.IsSet("port") {
		portStr = strconv.Itoa(c.Int("port"))
	}

	space := cmd.spaceReq.GetSpace()
	domain := cmd.domainReq.GetDomain()
	path := c.String("path")

	_, apiErr := cmd.CreateRoute(hostName, portStr, path, domain, space.SpaceFields)

	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}
}

func (cmd *CreateRoute) CreateRoute(hostName string, port string, path string, domain models.DomainFields, space models.SpaceFields) (models.Route, error) {
	var (
		portInt int
		err     error
	)

	if port != "" {
		portInt, err = strconv.Atoi(port)
		if err != nil {
			return models.Route{}, err
		}
	}

	uiArgMap :=
		map[string]interface{}{
			"URL":       terminal.EntityNameColor(domain.UrlForHostAndPath(hostName, portInt, path)),
			"OrgName":   terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			"SpaceName": terminal.EntityNameColor(space.Name),
			"Username":  terminal.EntityNameColor(cmd.config.Username()),
		}

	cmd.ui.Say(T("Creating route {{.URL}} for org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", uiArgMap))

	route, err := cmd.routeRepo.CreateInSpace(hostName, port, path, domain.Guid, space.Guid)
	if err != nil {
		var findErr error
		route, findErr = cmd.routeRepo.Find(hostName, port, domain, path)
		httpErr, isHttpError := err.(errors.HttpError)
		if !isHttpError || (httpErr.ErrorCode() != errors.PORT_TAKEN && httpErr.ErrorCode() != errors.HOST_TAKEN) {
			return models.Route{}, err
		}

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
