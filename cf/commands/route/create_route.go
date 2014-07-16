package route

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/flag_helpers"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type RouteCreator interface {
	CreateRoute(hostName string, domain models.DomainFields, space models.SpaceFields) (route models.Route, apiErr error)
}

type CreateRoute struct {
	ui        terminal.UI
	config    configuration.Reader
	routeRepo api.RouteRepository
	spaceReq  requirements.SpaceRequirement
	domainReq requirements.DomainRequirement
}

func NewCreateRoute(ui terminal.UI, config configuration.Reader, routeRepo api.RouteRepository) (cmd *CreateRoute) {
	cmd = new(CreateRoute)
	cmd.ui = ui
	cmd.config = config
	cmd.routeRepo = routeRepo
	return
}

func (cmd *CreateRoute) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "create-route",
		Description: T("Create a url route in a space for later use"),
		Usage:       T("CF_NAME create-route SPACE DOMAIN [-n HOSTNAME]"),
		Flags: []cli.Flag{
			flag_helpers.NewStringFlag("n", T("Hostname")),
		},
	}
}

func (cmd *CreateRoute) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {

	if len(c.Args()) != 2 {
		cmd.ui.FailWithUsage(c)
	}

	spaceName := c.Args()[0]
	domainName := c.Args()[1]

	cmd.spaceReq = requirementsFactory.NewSpaceRequirement(spaceName)
	cmd.domainReq = requirementsFactory.NewDomainRequirement(domainName)

	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedOrgRequirement(),
		cmd.spaceReq,
		cmd.domainReq,
	}
	return
}

func (cmd *CreateRoute) Run(c *cli.Context) {
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
