package service

import (
	"strings"

	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
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

func (cmd ListServices) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "services",
		ShortName:   "s",
		Description: T("List all service instances in the target space"),
		Usage:       "CF_NAME services",
	}
}

func (cmd ListServices) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	reqs = append(reqs,
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
	)
	return
}

func (cmd ListServices) Run(c *cli.Context) {
	cmd.ui.Say(T("Getting services in org {{.OrgName}} / space {{.SpaceName}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"OrgName":     terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			"SpaceName":   terminal.EntityNameColor(cmd.config.SpaceFields().Name),
			"CurrentUser": terminal.EntityNameColor(cmd.config.Username()),
		}))

	serviceInstances, apiErr := cmd.serviceSummaryRepo.GetSummariesInCurrentSpace()

	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say("")

	if len(serviceInstances) == 0 {
		cmd.ui.Say(T("No services found"))
		return
	}

	table := terminal.NewTable(cmd.ui, []string{T("name"), T("service"), T("plan"), T("bound apps")})

	for _, instance := range serviceInstances {
		var serviceColumn string

		if instance.IsUserProvided() {
			serviceColumn = T("user-provided")
		} else {
			serviceColumn = instance.ServiceOffering.Label
		}

		table.Add(
			instance.Name,
			serviceColumn,
			instance.ServicePlan.Name,
			strings.Join(instance.ApplicationNames, ", "),
		)
	}

	table.Print()
}
