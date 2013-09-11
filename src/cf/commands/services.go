package commands

import (
	"cf/api"
	"cf/requirements"
	term "cf/terminal"
	"github.com/codegangsta/cli"
	"strings"
)

type Services struct {
	ui        term.UI
	spaceRepo api.SpaceRepository
}

func NewServices(ui term.UI, spaceRepo api.SpaceRepository) (cmd Services) {
	cmd.ui = ui
	cmd.spaceRepo = spaceRepo
	return
}

func (cmd Services) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	return
}

func (cmd Services) Run(c *cli.Context) {
	cmd.ui.Say("Getting services in %s", cmd.spaceRepo.GetCurrentSpace().Name)

	space, err := cmd.spaceRepo.GetSummary()

	if err != nil {
		cmd.ui.Failed(err.Error())
		return
	}

	cmd.ui.Ok()

	table := [][]string{
		[]string{"name", "service", "provider", "version", "plan", "bound apps"},
	}

	for _, instance := range space.ServiceInstances {
		table = append(table, []string{
			instance.Name,
			instance.ServicePlan.ServiceOffering.Label,
			instance.ServicePlan.ServiceOffering.Provider,
			instance.ServicePlan.ServiceOffering.Version,
			instance.ServicePlan.Name,
			strings.Join(instance.ApplicationNames, ", "),
		})
	}

	cmd.ui.DisplayTable(table, nil)
}
