package quota

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/flag_helpers"
	"github.com/cloudfoundry/cli/cf/formatters"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
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

func (command CreateQuota) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "create-quota",
		Description: "Define a new resource quota",
		Usage:       "CF_NAME create-quota QUOTA [-m MEMORY] [-r ROUTES] [-s SERVICE_INSTANCES] [--allow-paid-service-plans]",
		Flags: []cli.Flag{
			flag_helpers.NewStringFlag("m", "Total amount of memory (e.g. 1024M, 1G, 10G)"),
			flag_helpers.NewIntFlag("r", "Total number of routes"),
			flag_helpers.NewIntFlag("s", "Total number of service instances"),
			cli.BoolFlag{Name: "allow-paid-service-plans", Usage: "Can provision instances of paid service plans"},
		},
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

	cmd.ui.Say("Creating quota %s as %s...",
		terminal.EntityNameColor(name),
		terminal.EntityNameColor(cmd.config.Username()))

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
		quota.RoutesLimit = context.Int("r")
	}

	if context.IsSet("s") {
		quota.ServicesLimit = context.Int("s")
	}

	if context.IsSet("allow-paid-service-plans") {
		quota.NonBasicServicesAllowed = true
	}

	err := cmd.quotaRepo.Create(quota)

	httpErr, ok := err.(errors.HttpError)
	if ok && httpErr.ErrorCode() == errors.QUOTA_EXISTS {
		cmd.ui.Ok()
		cmd.ui.Warn("Quota Definition %s already exists", quota.Name)
		return
	}

	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	cmd.ui.Ok()
}
