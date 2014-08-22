package spacequota

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/api/space_quotas"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/flag_helpers"
	"github.com/cloudfoundry/cli/cf/formatters"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type CreateSpaceQuota struct {
	ui        terminal.UI
	config    configuration.Reader
	quotaRepo space_quotas.SpaceQuotaRepository
	orgRepo   api.OrganizationRepository
}

func NewCreateSpaceQuota(ui terminal.UI, config configuration.Reader, quotaRepo space_quotas.SpaceQuotaRepository, orgRepo api.OrganizationRepository) CreateSpaceQuota {
	return CreateSpaceQuota{
		ui:        ui,
		config:    config,
		quotaRepo: quotaRepo,
		orgRepo:   orgRepo,
	}
}

func (cmd CreateSpaceQuota) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "create-space-quota",
		Description: T("Define a new space resource quota"),
		Usage:       T("CF_NAME create-space-quota QUOTA [-m MEMORY] [-r ROUTES] [-s SERVICE_INSTANCES] [--allow-paid-service-plans]"),
		Flags: []cli.Flag{
			flag_helpers.NewStringFlag("i", T("Maximum amount of memory an application instance can have(e.g. 1024M, 1G, 10G). -1 represents an unlimited amount.")),
			flag_helpers.NewStringFlag("m", T("Total amount of memory a space can have(e.g. 1024M, 1G, 10G)")),
			flag_helpers.NewIntFlag("r", T("Total number of routes")),
			flag_helpers.NewIntFlag("s", T("Total number of service instances")),
			cli.BoolFlag{Name: "allow-paid-service-plans", Usage: T("Can provision instances of paid service plans")},
		},
	}
}

func (cmd CreateSpaceQuota) GetRequirements(requirementsFactory requirements.Factory, context *cli.Context) ([]requirements.Requirement, error) {
	if len(context.Args()) != 1 {
		cmd.ui.FailWithUsage(context)
	}

	return []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedOrgRequirement(),
	}, nil
}

func (cmd CreateSpaceQuota) Run(context *cli.Context) {
	name := context.Args()[0]
	org := cmd.config.OrganizationFields()

	cmd.ui.Say(T("Creating space quota {{.QuotaName}} for org {{.OrgName}} as {{.Username}}...", map[string]interface{}{
		"QuotaName": terminal.EntityNameColor(name),
		"OrgName":   terminal.EntityNameColor(org.Name),
		"Username":  terminal.EntityNameColor(cmd.config.Username()),
	}))

	quota := models.SpaceQuota{
		Name:    name,
		OrgGuid: org.Guid,
	}

	memoryLimit := context.String("m")
	if memoryLimit != "" {
		parsedMemory, errr := formatters.ToMegabytes(memoryLimit)
		if errr != nil {
			cmd.ui.Failed(T("Invalid memory limit: {{.MemoryLimit}}\n{{.Err}}", map[string]interface{}{"MemoryLimit": memoryLimit, "Err": errr}))
		}

		quota.MemoryLimit = parsedMemory
	}

	instanceMemoryLimit := context.String("i")
	if instanceMemoryLimit != "" {
		var parsedMemory int64
		var err error

		if instanceMemoryLimit == "-1" {
			parsedMemory = -1
		} else {
			parsedMemory, err = formatters.ToMegabytes(instanceMemoryLimit)
			if err != nil {
				cmd.ui.Failed(T("Invalid instance memory limit: {{.MemoryLimit}}\n{{.Err}}", map[string]interface{}{"MemoryLimit": instanceMemoryLimit, "Err": err}))
			}
		}

		quota.InstanceMemoryLimit = parsedMemory
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
		cmd.ui.Warn(T("Space Quota Definition {{.QuotaName}} already exists", map[string]interface{}{"QuotaName": quota.Name}))
		return
	}

	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	cmd.ui.Ok()
}
