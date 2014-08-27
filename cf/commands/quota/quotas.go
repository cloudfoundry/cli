package quota

import (
	"fmt"
	. "github.com/cloudfoundry/cli/cf/i18n"

	"github.com/cloudfoundry/cli/cf/api/quotas"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/formatters"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type ListQuotas struct {
	ui        terminal.UI
	config    configuration.Reader
	quotaRepo quotas.QuotaRepository
}

func NewListQuotas(ui terminal.UI, config configuration.Reader, quotaRepo quotas.QuotaRepository) (cmd *ListQuotas) {
	return &ListQuotas{
		ui:        ui,
		config:    config,
		quotaRepo: quotaRepo,
	}
}

func (cmd *ListQuotas) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "quotas",
		Description: T("List available usage quotas"),
		Usage:       T("CF_NAME quotas"),
	}
}

func (cmd *ListQuotas) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}
	return
}

func (cmd *ListQuotas) Run(c *cli.Context) {
	cmd.ui.Say(T("Getting quotas as {{.Username}}...", map[string]interface{}{"Username": terminal.EntityNameColor(cmd.config.Username())}))

	quotas, apiErr := cmd.quotaRepo.FindAll()

	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}
	cmd.ui.Ok()
	cmd.ui.Say("")

	table := terminal.NewTable(cmd.ui, []string{T("name"), T("total memory limit"), T("instance memory limit"), T("routes"), T("service instances"), T("paid service plans")})

	var megabytes string
	for _, quota := range quotas {
		if quota.InstanceMemoryLimit == -1 {
			megabytes = "-1"
		} else {
			megabytes = formatters.ByteSize(quota.InstanceMemoryLimit * formatters.MEGABYTE)
		}

		table.Add(
			quota.Name,
			formatters.ByteSize(quota.MemoryLimit*formatters.MEGABYTE),
			megabytes,
			fmt.Sprintf("%d", quota.RoutesLimit),
			fmt.Sprintf("%d", quota.ServicesLimit),
			formatters.Allowed(quota.NonBasicServicesAllowed),
		)
	}

	table.Print()
}
