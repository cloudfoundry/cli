package commands

import (
	"cf/api"
	"cf/requirements"
	term "cf/terminal"
	"github.com/codegangsta/cli"
	"strings"
)

type MarketplaceServices struct {
	ui          term.UI
	serviceRepo api.ServiceRepository
}

func NewMarketplaceServices(ui term.UI, serviceRepo api.ServiceRepository) (cmd MarketplaceServices) {
	cmd.ui = ui
	cmd.serviceRepo = serviceRepo
	return
}

func (cmd MarketplaceServices) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	return
}

func (cmd MarketplaceServices) Run(c *cli.Context) {
	cmd.ui.Say("Getting services from marketplace...")

	serviceOfferings, err := cmd.serviceRepo.GetServiceOfferings()

	if err != nil {
		cmd.ui.Failed(err.Error())
		return
	}

	cmd.ui.Ok()

	table := [][]string{
		[]string{"service", "version", "provider", "plans", "description"},
	}

	for _, offering := range serviceOfferings {
		var planNames []string
		for _, plan := range offering.Plans {
			planNames = append(planNames, plan.Name)
		}

		table = append(table, []string{
			offering.Label,
			offering.Version,
			offering.Provider,
			strings.Join(planNames, ", "),
			offering.Description,
		})
	}

	cmd.ui.DisplayTable(table, nil)
	return
}
