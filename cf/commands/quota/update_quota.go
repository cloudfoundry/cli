package quota

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/flag_helpers"
	"github.com/cloudfoundry/cli/cf/formatters"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
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

func (cmd *updateQuota) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "update-quota",
		Description: T("Update an existing resource quota"),
		Usage:       T("CF_NAME update-quota QUOTA [-m MEMORY] [-n NEW_NAME] [-r ROUTES] [-s SERVICE_INSTANCES] [--allow-paid-service-plans | --disallow-paid-service-plans]"),
		Flags: []cli.Flag{
			flag_helpers.NewStringFlag("m", T("Total amount of memory (e.g. 1024M, 1G, 10G)")),
			flag_helpers.NewStringFlag("n", T("New name")),
			flag_helpers.NewIntFlag("r", T("Total number of routes")),
			flag_helpers.NewIntFlag("s", T("Total number of service instances")),
			cli.BoolFlag{Name: "allow-paid-service-plans", Usage: T("Can provision instances of paid service plans")},
			cli.BoolFlag{Name: "disallow-paid-service-plans", Usage: T("Cannot provision instances of paid service plans")},
		},
	}
}

func (cmd *updateQuota) GetRequirements(requirementsFactory requirements.Factory, context *cli.Context) ([]requirements.Requirement, error) {
	if len(context.Args()) != 1 {
		cmd.ui.FailWithUsage(context)
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

	allowPaidServices := c.Bool("allow-paid-service-plans")
	disallowPaidServices := c.Bool("disallow-paid-service-plans")
	if allowPaidServices && disallowPaidServices {
		cmd.ui.Failed(T("Please choose either allow or disallow. Both flags are not permitted to be passed in the same command."))
	}

	if allowPaidServices {
		quota.NonBasicServicesAllowed = true
	}

	if disallowPaidServices {
		quota.NonBasicServicesAllowed = false
	}

	if c.String("m") != "" {
		memory, formatError := formatters.ToMegabytes(c.String("m"))

		if formatError != nil {
			cmd.ui.FailWithUsage(c)
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

	cmd.ui.Say(T("Updating quota {{.QuotaName}} as {{.Username}}...", map[string]interface{}{
		"QuotaName": terminal.EntityNameColor(oldQuotaName),
		"Username":  terminal.EntityNameColor(cmd.config.Username()),
	}))

	err = cmd.quotaRepo.Update(quota)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}
	cmd.ui.Ok()
}
