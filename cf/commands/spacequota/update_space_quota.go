package spacequota

import (
	"github.com/cloudfoundry/cli/cf/api/space_quotas"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/flag_helpers"
	"github.com/cloudfoundry/cli/cf/formatters"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type UpdateSpaceQuota struct {
	ui             terminal.UI
	config         core_config.Reader
	spaceQuotaRepo space_quotas.SpaceQuotaRepository
}

func NewUpdateSpaceQuota(ui terminal.UI, config core_config.Reader, spaceQuotaRepo space_quotas.SpaceQuotaRepository) (cmd *UpdateSpaceQuota) {
	return &UpdateSpaceQuota{
		ui:             ui,
		config:         config,
		spaceQuotaRepo: spaceQuotaRepo,
	}
}

func (cmd *UpdateSpaceQuota) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "update-space-quota",
		Description: T("update an existing space quota"),
		Usage:       T("CF_NAME update-space-quota SPACE-QUOTA-NAME [-i MAX-INSTANCE-MEMORY] [-m MEMORY] [-n NEW_NAME] [-r ROUTES] [-s SERVICES] [--allow-paid-service-plans | --disallow-paid-service-plans]"),
		Flags: []cli.Flag{
			flag_helpers.NewStringFlag("i", T("Maximum amount of memory an application instance can have (e.g. 1024M, 1G, 10G). -1 represents an unlimited amount.")),
			flag_helpers.NewStringFlag("m", T("Total amount of memory a space can have (e.g. 1024M, 1G, 10G)")),
			flag_helpers.NewStringFlag("n", T("New name")),
			flag_helpers.NewIntFlag("r", T("Total number of routes")),
			flag_helpers.NewIntFlag("s", T("Total number of service instances")),
			cli.BoolFlag{Name: "allow-paid-service-plans", Usage: T("Can provision instances of paid service plans")},
			cli.BoolFlag{Name: "disallow-paid-service-plans", Usage: T("Can not provision instances of paid service plans")},
		},
	}
}

func (cmd *UpdateSpaceQuota) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		cmd.ui.FailWithUsage(c)
	}

	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedOrgRequirement(),
	}
	return
}

func (cmd *UpdateSpaceQuota) Run(c *cli.Context) {
	name := c.Args()[0]

	spaceQuota, apiErr := cmd.spaceQuotaRepo.FindByName(name)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	allowPaidServices := c.Bool("allow-paid-service-plans")
	disallowPaidServices := c.Bool("disallow-paid-service-plans")
	if allowPaidServices && disallowPaidServices {
		cmd.ui.Failed(T("Please choose either allow or disallow. Both flags are not permitted to be passed in the same command."))
	}

	if allowPaidServices {
		spaceQuota.NonBasicServicesAllowed = true
	}

	if disallowPaidServices {
		spaceQuota.NonBasicServicesAllowed = false
	}

	if c.String("i") != "" {
		var memory int64
		var formatError error

		memFlag := c.String("i")

		if memFlag == "-1" {
			memory = -1
		} else {
			memory, formatError = formatters.ToMegabytes(memFlag)
			if formatError != nil {
				cmd.ui.FailWithUsage(c)
			}
		}

		spaceQuota.InstanceMemoryLimit = memory
	}

	if c.String("m") != "" {
		memory, formatError := formatters.ToMegabytes(c.String("m"))

		if formatError != nil {
			cmd.ui.FailWithUsage(c)
		}

		spaceQuota.MemoryLimit = memory
	}

	if c.String("n") != "" {
		spaceQuota.Name = c.String("n")
	}

	if c.IsSet("s") {
		spaceQuota.ServicesLimit = c.Int("s")
	}

	if c.IsSet("r") {
		spaceQuota.RoutesLimit = c.Int("r")
	}

	cmd.ui.Say(T("Updating space quota {{.Quota}} as {{.Username}}...",
		map[string]interface{}{
			"Quota":    terminal.EntityNameColor(name),
			"Username": terminal.EntityNameColor(cmd.config.Username()),
		}))

	apiErr = cmd.spaceQuotaRepo.Update(spaceQuota)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Ok()
}
