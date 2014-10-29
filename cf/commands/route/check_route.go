package route

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type CheckRoute struct {
	ui         terminal.UI
	config     core_config.Reader
	routeRepo  api.RouteRepository
	domainRepo api.DomainRepository
}

func NewCheckRoute(ui terminal.UI, config core_config.Reader, routeRepo api.RouteRepository, domainRepo api.DomainRepository) (cmd *CheckRoute) {
	cmd = new(CheckRoute)
	cmd.ui = ui
	cmd.config = config
	cmd.routeRepo = routeRepo
	cmd.domainRepo = domainRepo
	return
}

func (cmd *CheckRoute) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "check-route",
		Description: T("Perform a simple check to determine whether a route currently exists or not."),
		Usage:       T("CF_NAME check-route HOST DOMAIN"),
	}
}

func (cmd *CheckRoute) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 2 {
		cmd.ui.FailWithUsage(c)
		return
	}
	reqs = []requirements.Requirement{
		requirementsFactory.NewTargetedOrgRequirement(),
		requirementsFactory.NewLoginRequirement(),
	}
	return
}

func (cmd *CheckRoute) Run(c *cli.Context) {
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
