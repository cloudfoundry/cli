package service

import (
	"fmt"

	. "github.com/cloudfoundry/cli/cf/i18n"

	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type ListUserProvidedServices struct {
	ui      terminal.UI
	config  core_config.Reader
	upsRepo api.UserProvidedServiceInstanceRepository
}

func NewListUserProvidedServices(ui terminal.UI, config core_config.Reader, upsSummaryRepo api.UserProvidedServiceInstanceRepository) ListUserProvidedServices {
	cmd := ListUserProvidedServices{}
	cmd.ui = ui
	cmd.config = config
	cmd.upsRepo = upsSummaryRepo
	return cmd
}

func (cmd ListUserProvidedServices) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "user-provided-services",
		ShortName:   "ups",
		Description: T("List all user provided service instances in the target space"),
		Usage:       "CF_NAME user-provided-services",
	}
}

func (cmd ListUserProvidedServices) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 0 {
		cmd.ui.FailWithUsage(c)
	}
	reqs = append(reqs,
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
	)
	return
}

func (cmd ListUserProvidedServices) Run(c *cli.Context) {
	cmd.ui.Say(T("Getting user provided services in org {{.OrgName}} / space {{.SpaceName}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"OrgName":     terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			"SpaceName":   terminal.EntityNameColor(cmd.config.SpaceFields().Name),
			"CurrentUser": terminal.EntityNameColor(cmd.config.Username()),
		}))

	summaryModel, apiErr := cmd.upsRepo.GetSummaries()

	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say("")

	if len(summaryModel.Resources) == 0 {
		cmd.ui.Say(T("No user provided services found"))
		return
	}

	table := terminal.NewTable(cmd.ui, []string{T("name"), T("key"), T("value"), T("syslog drain")})

	for _, instance := range summaryModel.Resources {
		name := instance.Name
		if len(instance.Credentials) > 0 {
			for k, v := range instance.Credentials {
				table.Add(name, k, fmt.Sprintf("%v", v), instance.SysLogDrainUrl)
				name = ""
			}
		} else {
			table.Add(name, "n/a", "n/a")
		}
		table.Add()
	}

	table.Print()
}
