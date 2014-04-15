package quota

import (
	"cf/api"
	"cf/configuration"
	"cf/formatters"
	"cf/requirements"
	"cf/terminal"
	"github.com/codegangsta/cli"
)

type updateQuota struct {
	ui        terminal.UI
	config    configuration.Reader
	quotaRepo api.QuotaRepository
}

func NewUpdateQuota(ui terminal.UI, config configuration.Reader, quotaRepo api.QuotaRepository) *updateQuota {
	return &updateQuota{
		ui:        ui,
		config:    config,
		quotaRepo: quotaRepo,
	}
}

func (cmd *updateQuota) GetRequirements(requirementsFactory requirements.Factory, context *cli.Context) ([]requirements.Requirement, error) {
	if len(context.Args()) != 1 {
		cmd.ui.FailWithUsage(context, "update-quota")
	}

	return []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}, nil
}

func (cmd *updateQuota) Run(c *cli.Context) {
	oldQuotaName := c.Args()[0]
	quota, err := cmd.quotaRepo.FindByName(oldQuotaName)

	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	if c.String("m") != "" {
		memory, err := formatters.ToMegabytes(c.String("m"))

		if err != nil {
			cmd.ui.FailWithUsage(c, "update-quota")
		}

		quota.MemoryLimit = memory
	}

	if c.String("n") != "" {
		quota.Name = c.String("n")
	}

	if c.IsSet("s") {
		quota.ServicesLimit = c.Int("s")
	}

	if c.IsSet("r") {
		quota.RoutesLimit = c.Int("r")
	}

	cmd.ui.Say("Updating quota %s as %s...",
		terminal.EntityNameColor(oldQuotaName),
		terminal.EntityNameColor(cmd.config.Username()))

	err = cmd.quotaRepo.Update(quota)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}
	cmd.ui.Ok()
}
