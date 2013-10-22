package service

import (
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	"cf/terminal"
	"github.com/codegangsta/cli"
	"strings"
)

type ListServices struct {
	ui                 terminal.UI
	config             *configuration.Configuration
	serviceSummaryRepo api.ServiceSummaryRepository
}

func NewListServices(ui terminal.UI, config *configuration.Configuration, serviceSummaryRepo api.ServiceSummaryRepository) (cmd ListServices) {
	cmd.ui = ui
	cmd.config = config
	cmd.serviceSummaryRepo = serviceSummaryRepo
	return
}

func (cmd ListServices) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	return
}

func (cmd ListServices) Run(c *cli.Context) {
	cmd.ui.Say("Getting services in org %s / space %s as %s...",
		terminal.EntityNameColor(cmd.config.Organization.Name),
		terminal.EntityNameColor(cmd.config.Space.Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	serviceInstances, apiResponse := cmd.serviceSummaryRepo.GetSummariesInCurrentSpace()

	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say("")

	table := [][]string{
		[]string{"name", "service", "plan", "bound apps"},
	}

	for _, instance := range serviceInstances {
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
