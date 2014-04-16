package quota

import (
	"cf/api"
	"cf/command_metadata"
	"cf/configuration"
	"cf/formatters"
	"cf/requirements"
	"cf/terminal"
	"fmt"
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

func (command *ListQuotas) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "quotas",
		Description: "List available usage quotas",
		Usage:       "CF_NAME quotas",
	}
}

func (cmd *ListQuotas) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
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

	table := terminal.NewTable(cmd.ui, []string{"name", "memory limit", "routes", "service instances"})
	rows := [][]string{}

	for _, quota := range quotas {
		rows = append(rows, []string{
			quota.Name,
			formatters.ByteSize(quota.MemoryLimit * formatters.MEGABYTE),
			fmt.Sprintf("%d", quota.RoutesLimit),
			fmt.Sprintf("%d", quota.ServicesLimit),
		})
	}

	table.Print(rows)
}
