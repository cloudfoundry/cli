package service

import (
	"cf/api"
	"cf/requirements"
	"cf/terminal"
	"github.com/codegangsta/cli"
	"strings"
)

type ListServices struct {
	ui        terminal.UI
	spaceRepo api.SpaceRepository
}

func NewListServices(ui terminal.UI, spaceRepo api.SpaceRepository) (cmd ListServices) {
	cmd.ui = ui
	cmd.spaceRepo = spaceRepo
	return
}

func (cmd ListServices) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	return
}

func (cmd ListServices) Run(c *cli.Context) {
	cmd.ui.Say("Getting services in %s...", cmd.spaceRepo.GetCurrentSpace().Name)

	space, apiResponse := cmd.spaceRepo.GetSummary()

	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()

	table := [][]string{
		[]string{"name", "service", "plan", "bound apps"},
	}

	for _, instance := range space.ServiceInstances {
		var serviceColumn string

		if instance.IsUserProvided() {
			serviceColumn = "user-provided"
		} else {
			serviceColumn = instance.ServiceOffering().Label
		}

		table = append(table, []string{
			instance.Name,
			serviceColumn,
			instance.ServicePlan.Name,
			strings.Join(instance.ApplicationNames, ", "),
		})
	}

	cmd.ui.DisplayTable(table)
}
