package route

import (
	"strings"

	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/simonleung8/flags"
	"github.com/simonleung8/flags/flag"
)

type RouteCreator interface {
	CreateRoute(hostName, path string, domain models.DomainFields, space models.SpaceFields) (route models.Route, apiErr error)
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
	fs["n"] = &cliFlags.StringFlag{Name: "n", Usage: T("Hostname for the route")}
	fs["path"] = &cliFlags.StringFlag{Name: "path", Usage: T("Path for the route")}

	return command_registry.CommandMetadata{
		Name:        "create-route",
		Description: T("Create a url route in a space for later use"),
		Usage: T(`CF_NAME create-route SPACE DOMAIN [-n HOSTNAME] [--path PATH]

EXAMPLES:
   CF_NAME create-route my-space example.com                     # example.com
   CF_NAME create-route my-space example.com -n myapp            # myapp.example.com
   CF_NAME create-route my-space example.com -n myapp --path foo # myapp.example.com/foo`),
		Flags: fs,
	}
}

func (cmd *CreateRoute) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(fc.Args()) != 2 {
		cmd.ui.Failed(T("Incorrect Usage. Requires SPACE and DOMAIN as arguments\n\n") + command_registry.Commands.CommandUsage("create-route"))
	}

	domainName := fc.Args()[1]

	cmd.spaceReq = requirementsFactory.NewSpaceRequirement(fc.Args()[0])
	cmd.domainReq = requirementsFactory.NewDomainRequirement(domainName)

	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedOrgRequirement(),
		cmd.spaceReq,
		cmd.domainReq,
	}
	return
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

	_, apiErr := cmd.CreateRoute(hostName, path, domain, space.SpaceFields)

	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}
}

func (cmd *CreateRoute) CreateRoute(hostName string, path string, domain models.DomainFields, space models.SpaceFields) (models.Route, error) {
	if path != "" && !strings.HasPrefix(path, `/`) {
		path = `/` + path
	}

	cmd.ui.Say(T("Creating route {{.URL}} for org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...",
		map[string]interface{}{
			"URL":       terminal.EntityNameColor(domain.UrlForHostAndPath(hostName, path)),
			"OrgName":   terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			"SpaceName": terminal.EntityNameColor(space.Name),
			"Username":  terminal.EntityNameColor(cmd.config.Username())}))

	route, err := cmd.routeRepo.CreateInSpace(hostName, path, domain.Guid, space.Guid)
	if err != nil {
		var findErr error
		route, findErr = cmd.routeRepo.FindByHostAndDomain(hostName, domain)
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
