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
	config             configuration.Reader
	serviceSummaryRepo api.ServiceSummaryRepository
}

func NewListServices(ui terminal.UI, config configuration.Reader, serviceSummaryRepo api.ServiceSummaryRepository) (cmd ListServices) {
	cmd.ui = ui
	cmd.config = config
	cmd.serviceSummaryRepo = serviceSummaryRepo
	return
}

func (cmd ListServices) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	reqs = append(reqs,
		reqFactory.NewLoginRequirement(),
		reqFactory.NewTargetedSpaceRequirement(),
	)
	return
}

func (cmd ListServices) Run(c *cli.Context) {
	cmd.ui.Say("Getting services in org %s / space %s as %s...",
		terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
		terminal.EntityNameColor(cmd.config.SpaceFields().Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	serviceInstances, apiResponse := cmd.serviceSummaryRepo.GetSummariesInCurrentSpace()

	if apiResponse != nil {
		cmd.ui.Failed(apiResponse.Error())
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say("")

	if len(serviceInstances) == 0 {
		cmd.ui.Say("No services found")
		return
	}

	table := [][]string{
		[]string{"name", "service", "plan", "bound apps"},
	}

	for _, instance := range serviceInstances {
		var serviceColumn string

		if instance.IsUserProvided() {
			serviceColumn = "user-provided"
		} else {
			serviceColumn = instance.ServiceOffering.Label
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
