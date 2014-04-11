package quota

import (
	"cf/api"
	"cf/configuration"
	"cf/formatters"
	"cf/models"
	"cf/requirements"
	"cf/terminal"
	"github.com/codegangsta/cli"
)

type CreateQuota struct {
	ui        terminal.UI
	config    configuration.Reader
	quotaRepo api.QuotaRepository
}

func NewCreateQuota(ui terminal.UI, config configuration.Reader, quotaRepo api.QuotaRepository) CreateQuota {
	return CreateQuota{
		ui:        ui,
		config:    config,
		quotaRepo: quotaRepo,
	}
}

func (cmd CreateQuota) GetRequirements(requirementsFactory requirements.Factory, context *cli.Context) ([]requirements.Requirement, error) {
	if len(context.Args()) != 1 {
		cmd.ui.FailWithUsage(context, "create-quota")
	}

	return []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}, nil
}

func (cmd CreateQuota) Run(context *cli.Context) {
	name := context.Args()[0]

	cmd.ui.Say("Creating quota %s as %s", name, cmd.config.Username())
	quota := models.QuotaFields{
		Name: name,
	}

	memoryLimit := context.String("m")
	if memoryLimit != "" {
		parsedMemory, err := formatters.ToMegabytes(memoryLimit)
		if err != nil {
			cmd.ui.Failed("Invalid memory limit: %s\n%s", memoryLimit, err)
		}

		quota.MemoryLimit = parsedMemory
	}

	if context.IsSet("r") {
		quota.RoutesLimit = uint(context.Int("r"))
	}

	if context.IsSet("s") {
		quota.ServicesLimit = uint(context.Int("s"))
	}

	err := cmd.quotaRepo.Create(quota)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	cmd.ui.Ok()
}
