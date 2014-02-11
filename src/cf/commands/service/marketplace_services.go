package service

import (
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	"cf/terminal"
	"github.com/codegangsta/cli"
	"sort"
	"strings"
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

func (cmd MarketplaceServices) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	return
}

func (cmd MarketplaceServices) Run(c *cli.Context) {
	if cmd.config.HasSpace() {
		cmd.ui.Say("Getting services from marketplace in org %s / space %s as %s...",
			terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			terminal.EntityNameColor(cmd.config.SpaceFields().Name),
			terminal.EntityNameColor(cmd.config.Username()),
		)
	} else {
		cmd.ui.Say("Getting services from marketplace...")
	}

	serviceOfferings, apiResponse := cmd.serviceRepo.GetServiceOfferings()

	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say("")

	if len(serviceOfferings) == 0 {
		cmd.ui.Say("No service offerings found")
		return
	}

	table := [][]string{
		[]string{"service", "plans", "description"},
	}

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

		table = append(table, []string{
			offering.Label,
			planNames,
			offering.Description,
		})
	}

	cmd.ui.DisplayTable(table)
	return
}
