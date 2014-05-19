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

	"github.com/cloudfoundry/cli/cf/i18n"

	goi18n "github.com/nicksnyder/go-i18n/i18n"
)

type CreateQuota struct {
	ui        terminal.UI
	config    configuration.Reader
	quotaRepo api.QuotaRepository
	T         goi18n.TranslateFunc
}

func NewCreateQuota(ui terminal.UI, config configuration.Reader, quotaRepo api.QuotaRepository) CreateQuota {
	t, err := i18n.Init("quota", i18n.RESOURCES_PATH)
	if err != nil {
		ui.Failed(err.Error())
	}

	return CreateQuota{
		ui:        ui,
		config:    config,
		quotaRepo: quotaRepo,
		T:         t,
	}
}

func (cmd CreateQuota) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "create-quota",
		Description: cmd.T("Define a new resource quota"),
		Usage:       cmd.T("CF_NAME create-quota QUOTA [-m MEMORY] [-r ROUTES] [-s SERVICE_INSTANCES] [--allow-paid-service-plans]"),
		Flags: []cli.Flag{
			flag_helpers.NewStringFlag("m", cmd.T("Total amount of memory (e.g. 1024M, 1G, 10G)")),
			flag_helpers.NewIntFlag("r", cmd.T("Total number of routes")),
			flag_helpers.NewIntFlag("s", cmd.T("Total number of service instances")),
			cli.BoolFlag{Name: "allow-paid-service-plans", Usage: cmd.T("Can provision instances of paid service plans")},
		},
	}
}

func (cmd CreateQuota) GetRequirements(requirementsFactory requirements.Factory, context *cli.Context) ([]requirements.Requirement, error) {
	if len(context.Args()) != 1 {
		cmd.ui.FailWithUsage(context)
	}

	return []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}, nil
}

func (cmd CreateQuota) Run(context *cli.Context) {
	name := context.Args()[0]

	cmd.ui.Say(cmd.T("Creating quota {{.QuotaName}} as {{.Username}}...", map[string]interface{}{
		"QuotaName": terminal.EntityNameColor(name),
		"Username":  terminal.EntityNameColor(cmd.config.Username()),
	}))

	quota := models.QuotaFields{
		Name: name,
	}

	memoryLimit := context.String("m")
	if memoryLimit != "" {
		parsedMemory, err := formatters.ToMegabytes(memoryLimit)
		if err != nil {
			cmd.ui.Failed(cmd.T("Invalid memory limit: {{.MemoryLimit}}\n{{.Err}}", map[string]interface{}{"MemoryLimit": memoryLimit, "Err": err}))
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
		cmd.ui.Warn(cmd.T("Quota Definition {{.QuotaName}} already exists", map[string]interface{}{"QuotaName": quota.Name}))
		return
	}

	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	cmd.ui.Ok()
}
