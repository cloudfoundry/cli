package quota

import (
	"fmt"
	. "github.com/cloudfoundry/cli/cf/i18n"

	"github.com/cloudfoundry/cli/cf/api"
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
	quotaRepo api.QuotaRepository
}

func NewListQuotas(ui terminal.UI, config configuration.Reader, quotaRepo api.QuotaRepository) (cmd *ListQuotas) {
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

	table := terminal.NewTable(cmd.ui, []string{T("name"), T("memory limit"), T("routes"), T("service instances"), T("paid service plans")})

	for _, quota := range quotas {
		table.Add(
			quota.Name,
			formatters.ByteSize(quota.MemoryLimit*formatters.MEGABYTE),
			fmt.Sprintf("%d", quota.RoutesLimit),
			fmt.Sprintf("%d", quota.ServicesLimit),
			formatters.Allowed(quota.NonBasicServicesAllowed),
		)
	}

	table.Print()
}
