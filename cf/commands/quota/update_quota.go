package quota

import (
	"fmt"

	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/api/quotas"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/formatters"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
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
	fs["allow-paid-service-plans"] = &flags.BoolFlag{Name: "allow-paid-service-plans", Usage: T("Can provision instances of paid service plans")}
	fs["disallow-paid-service-plans"] = &flags.BoolFlag{Name: "disallow-paid-service-plans", Usage: T("Cannot provision instances of paid service plans")}
	fs["i"] = &flags.StringFlag{ShortName: "i", Usage: T("Maximum amount of memory an application instance can have (e.g. 1024M, 1G, 10G)")}
	fs["m"] = &flags.StringFlag{ShortName: "m", Usage: T("Total amount of memory (e.g. 1024M, 1G, 10G)")}
	fs["n"] = &flags.StringFlag{ShortName: "n", Usage: T("New name")}
	fs["r"] = &flags.IntFlag{ShortName: "r", Usage: T("Total number of routes")}
	fs["s"] = &flags.IntFlag{ShortName: "s", Usage: T("Total number of service instances")}
	fs["a"] = &flags.IntFlag{ShortName: "a", Usage: T("Total number of application instances. -1 represents an unlimited amount. (Default: unlimited)")}

	return command_registry.CommandMetadata{
		Name:        "update-quota",
		Description: T("Update an existing resource quota"),
		Usage: []string{
			"CF_NAME update-quota ",
			T("QUOTA"),
			" ",
			fmt.Sprintf("[-m %s] ", T("TOTAL_MEMORY")),
			fmt.Sprintf("[-i %s] ", T("INSTANCE_MEMORY")),
			fmt.Sprintf("[-n %s] ", T("NEW_NAME")),
			fmt.Sprintf("[-r %s] ", T("ROUTES")),
			fmt.Sprintf("[-s %s] ", T("SERVICE_INSTANCES")),
			fmt.Sprintf("[-a %s] ", T("APP_INSTANCES")),
			"[--allow-paid-service-plans | --disallow-paid-service-plans]",
		},
		Flags: fs,
	}
}

func (cmd *updateQuota) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + command_registry.Commands.CommandUsage("update-quota"))
	}

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}

	if fc.IsSet("a") {
		reqs = append(reqs, requirementsFactory.NewMinAPIVersionRequirement("Option '-a'", cf.OrgAppInstanceLimitMinimumApiVersion))
	}

	return reqs
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

	if c.IsSet("i") {
		var memory int64

		if c.String("i") == "-1" {
			memory = -1
		} else {
			var formatError error

			memory, formatError = formatters.ToMegabytes(c.String("i"))

			if formatError != nil {
				cmd.ui.Failed(T("Incorrect Usage") + "\n\n" + command_registry.Commands.CommandUsage("update-quota"))
			}
		}

		quota.InstanceMemoryLimit = memory
	}

	if c.IsSet("a") {
		quota.AppInstanceLimit = c.Int("a")
	}

	if c.IsSet("m") {
		memory, formatError := formatters.ToMegabytes(c.String("m"))

		if formatError != nil {
			cmd.ui.Failed(T("Incorrect Usage") + "\n\n" + command_registry.Commands.CommandUsage("update-quota"))
		}

		quota.MemoryLimit = memory
	}

	if c.IsSet("n") {
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
