package route

import (
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
	CreateRoute(hostName string, domain models.DomainFields, space models.SpaceFields) (route models.Route, apiErr error)
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
	fs["n"] = &cliFlags.StringFlag{Name: "n", Usage: T("Hostname")}

	return command_registry.CommandMetadata{
		Name:        "create-route",
		Description: T("Create a url route in a space for later use"),
		Usage:       T("CF_NAME create-route SPACE DOMAIN [-n HOSTNAME]"),
		Flags:       fs,
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

	_, apiErr := cmd.CreateRoute(hostName, domain, space.SpaceFields)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}
}

func (cmd *CreateRoute) CreateRoute(hostName string, domain models.DomainFields, space models.SpaceFields) (route models.Route, apiErr error) {
	cmd.ui.Say(T("Creating route {{.Hostname}} for org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...",
		map[string]interface{}{
			"Hostname":  terminal.EntityNameColor(domain.UrlForHost(hostName)),
			"OrgName":   terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			"SpaceName": terminal.EntityNameColor(space.Name),
			"Username":  terminal.EntityNameColor(cmd.config.Username())}))

	route, apiErr = cmd.routeRepo.CreateInSpace(hostName, domain.Guid, space.Guid)
	if apiErr != nil {
		var findApiResponse error
		route, findApiResponse = cmd.routeRepo.FindByHostAndDomain(hostName, domain)

		if findApiResponse != nil ||
			route.Space.Guid != space.Guid ||
			route.Domain.Guid != domain.Guid {
			return
		}

		apiErr = nil
		cmd.ui.Ok()
		cmd.ui.Warn(T("Route {{.URL}} already exists",
			map[string]interface{}{"URL": route.URL()}))
		return
	}

	cmd.ui.Ok()
	return
}
