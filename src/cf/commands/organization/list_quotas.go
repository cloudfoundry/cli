package organization

import (
	"cf/api"
	"cf/configuration"
	"cf/formatters"
	"cf/requirements"
	"cf/terminal"
	"github.com/codegangsta/cli"
)

type ListQuotas struct {
	ui        terminal.UI
	config    configuration.Reader
	quotaRepo api.QuotaRepository
}

func NewListQuotas(ui terminal.UI, config configuration.Reader, quotaRepo api.QuotaRepository) (cmd *ListQuotas) {
	cmd = new(ListQuotas)
	cmd.ui = ui
	cmd.config = config
	cmd.quotaRepo = quotaRepo
	return
}

func (cmd *ListQuotas) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
	}
	return
}

func (cmd *ListQuotas) Run(c *cli.Context) {
	cmd.ui.Say("Getting quotas as %s...", terminal.EntityNameColor(cmd.config.Username()))

	quotas, apiErr := cmd.quotaRepo.FindAll()

	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}
	cmd.ui.Ok()
	cmd.ui.Say("")

	table := [][]string{
		[]string{"name", "memory limit"},
	}

	for _, quota := range quotas {
		table = append(table, []string{
			quota.Name,
			formatters.ByteSize(quota.MemoryLimit * formatters.MEGABYTE),
		})
	}

	cmd.ui.DisplayTable(table)
}
