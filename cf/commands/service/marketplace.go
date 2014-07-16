package service

import (
	. "github.com/cloudfoundry/cli/cf/i18n"
	"sort"
	"strings"

	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type MarketplaceServices struct {
	ui          terminal.UI
	config      configuration.Reader
	serviceRepo api.ServiceRepository
}

func NewMarketplaceServices(ui terminal.UI, config configuration.Reader, serviceRepo api.ServiceRepository) (cmd MarketplaceServices) {
	cmd.ui = ui
	cmd.config = config
	cmd.serviceRepo = serviceRepo
	return
}

func (cmd MarketplaceServices) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "marketplace",
		ShortName:   "m",
		Description: T("List available offerings in the marketplace"),
		Usage:       "CF_NAME marketplace",
	}
}

func (cmd MarketplaceServices) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	reqs = append(reqs, requirementsFactory.NewApiEndpointRequirement())
	return
}

func (cmd MarketplaceServices) Run(c *cli.Context) {
	var (
		serviceOfferings models.ServiceOfferings
		apiErr           error
	)

	if cmd.config.HasSpace() {
		cmd.ui.Say(T("Getting services from marketplace in org {{.OrgName}} / space {{.SpaceName}} as {{.CurrentUser}}...",
			map[string]interface{}{
				"OrgName":     terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
				"SpaceName":   terminal.EntityNameColor(cmd.config.SpaceFields().Name),
				"CurrentUser": terminal.EntityNameColor(cmd.config.Username()),
			}))
		serviceOfferings, apiErr = cmd.serviceRepo.GetServiceOfferingsForSpace(cmd.config.SpaceFields().Guid)
	} else if !cmd.config.IsLoggedIn() {
		cmd.ui.Say(T("Getting all services from marketplace..."))
		serviceOfferings, apiErr = cmd.serviceRepo.GetAllServiceOfferings()
	} else {
		cmd.ui.Failed(T("Cannot list marketplace services without a targeted space"))
	}

	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say("")

	if len(serviceOfferings) == 0 {
		cmd.ui.Say(T("No service offerings found"))
		return
	}

	table := terminal.NewTable(cmd.ui, []string{T("service"), T("plans"), T("description")})

	sort.Sort(serviceOfferings)
	for _, offering := range serviceOfferings {
		planNames := ""

		for _, plan := range offering.Plans {
			if plan.Name == "" {
				continue
			}
			planNames = planNames + ", " + plan.Name
		}

		planNames = strings.TrimPrefix(planNames, ", ")

		table.Add(offering.Label, planNames, offering.Description)
	}

	table.Print()
	return
}
