package quota

import (
	"github.com/cloudfoundry/cli/cf/api/quotas"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/formatters"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
	"github.com/cloudfoundry/cli/flags/flag"
)

type updateQuota struct {
	ui        terminal.UI
	config    core_config.Reader
	quotaRepo quotas.QuotaRepository
}

func init() {
	command_registry.Register(&updateQuota{})
}

func (cmd *updateQuota) MetaData() command_registry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["allow-paid-service-plans"] = &cliFlags.BoolFlag{Name: "allow-paid-service-plans", Usage: T("Can provision instances of paid service plans")}
	fs["disallow-paid-service-plans"] = &cliFlags.BoolFlag{Name: "disallow-paid-service-plans", Usage: T("Cannot provision instances of paid service plans")}
	fs["i"] = &cliFlags.StringFlag{ShortName: "i", Usage: T("Maximum amount of memory an application instance can have (e.g. 1024M, 1G, 10G)")}
	fs["m"] = &cliFlags.StringFlag{ShortName: "m", Usage: T("Total amount of memory (e.g. 1024M, 1G, 10G)")}
	fs["n"] = &cliFlags.StringFlag{ShortName: "n", Usage: T("New name")}
	fs["r"] = &cliFlags.IntFlag{ShortName: "r", Usage: T("Total number of routes")}
	fs["s"] = &cliFlags.IntFlag{ShortName: "s", Usage: T("Total number of service instances")}

	return command_registry.CommandMetadata{
		Name:        "update-quota",
		Description: T("Update an existing resource quota"),
		Usage:       T("CF_NAME update-quota QUOTA [-m TOTAL_MEMORY] [-i INSTANCE_MEMORY][-n NEW_NAME] [-r ROUTES] [-s SERVICE_INSTANCES] [--allow-paid-service-plans | --disallow-paid-service-plans]"),
		Flags:       fs,
	}
}

func (cmd *updateQuota) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + command_registry.Commands.CommandUsage("update-quota"))
	}

	return []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}, nil
}

func (cmd *updateQuota) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.quotaRepo = deps.RepoLocator.GetQuotaRepository()
	return cmd
}

func (cmd *updateQuota) Execute(c flags.FlagContext) {
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

	if c.String("i") != "" {
		var memory int64

		if c.String("i") == "-1" {
			memory = -1
		} else {
			var formatError error

			memory, formatError = formatters.ToMegabytes(c.String("i"))

			if formatError != nil {
				cmd.ui.Failed(T("Incorrect Usage.\n\n") + command_registry.Commands.CommandUsage("update-quota"))
			}
		}

		quota.InstanceMemoryLimit = memory
	}

	if c.String("m") != "" {
		memory, formatError := formatters.ToMegabytes(c.String("m"))

		if formatError != nil {
			cmd.ui.Failed(T("Incorrect Usage.\n\n") + command_registry.Commands.CommandUsage("update-quota"))
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
