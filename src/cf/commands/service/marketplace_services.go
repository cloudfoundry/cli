package service

import (
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	"cf/terminal"
	"github.com/codegangsta/cli"
	"strings"
)

type MarketplaceServices struct {
	ui          terminal.UI
	config      *configuration.Configuration
	serviceRepo api.ServiceRepository
}

func NewMarketplaceServices(ui terminal.UI, config *configuration.Configuration, serviceRepo api.ServiceRepository) (cmd MarketplaceServices) {
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
			terminal.EntityNameColor(cmd.config.Organization.Name),
			terminal.EntityNameColor(cmd.config.Space.Name),
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

	table := [][]string{
		[]string{"service", "plans", "description"},
	}

	for _, offering := range serviceOfferings {
		var planNames []string
		for _, plan := range offering.Plans {
			planNames = append(planNames, plan.Name)
		}

		table = append(table, []string{
			offering.Label,
			strings.Join(planNames, ", "),
			offering.Description,
		})
	}

	cmd.ui.DisplayTable(table)
	return
}
