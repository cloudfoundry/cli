package spacequota

import (
	"fmt"

	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/api/space_quotas"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/formatters"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
)

type UpdateSpaceQuota struct {
	ui             terminal.UI
	config         core_config.Reader
	spaceQuotaRepo space_quotas.SpaceQuotaRepository
}

func init() {
	command_registry.Register(&UpdateSpaceQuota{})
}

func (cmd *UpdateSpaceQuota) MetaData() command_registry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["i"] = &flags.StringFlag{ShortName: "i", Usage: T("Maximum amount of memory an application instance can have (e.g. 1024M, 1G, 10G). -1 represents an unlimited amount.")}
	fs["m"] = &flags.StringFlag{ShortName: "m", Usage: T("Total amount of memory a space can have (e.g. 1024M, 1G, 10G)")}
	fs["n"] = &flags.StringFlag{ShortName: "n", Usage: T("New name")}
	fs["r"] = &flags.IntFlag{ShortName: "r", Usage: T("Total number of routes")}
	fs["s"] = &flags.IntFlag{ShortName: "s", Usage: T("Total number of service instances")}
	fs["allow-paid-service-plans"] = &flags.BoolFlag{Name: "allow-paid-service-plans", Usage: T("Can provision instances of paid service plans")}
	fs["disallow-paid-service-plans"] = &flags.BoolFlag{Name: "disallow-paid-service-plans", Usage: T("Can not provision instances of paid service plans")}
	fs["a"] = &flags.IntFlag{ShortName: "a", Usage: T("Total number of application instances. -1 represents an unlimited amount. (Default: unlimited)")}

	return command_registry.CommandMetadata{
		Name:        "update-space-quota",
		Description: T("Update an existing space quota"),
		Usage: []string{
			"CF_NAME update-space-quota ",
			T("QUOTA"),
			" ",
			fmt.Sprintf("[-i %s] ", T("INSTANCE_MEMORY")),
			fmt.Sprintf("[-m %s] ", T("MEMORY")),
			fmt.Sprintf("[-n %s] ", T("NAME")),
			fmt.Sprintf("[-r %s] ", T("ROUTES")),
			fmt.Sprintf("[-s %s] ", T("SERVICE_INSTANCES")),
			fmt.Sprintf("[-a %s] ", T("APP_INSTANCES")),
			"[--allow-paid-service-plans | --disallow-paid-service-plans]",
		},
		Flags: fs,
	}
}

func (cmd *UpdateSpaceQuota) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + command_registry.Commands.CommandUsage("update-space-quota"))
	}

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedOrgRequirement(),
	}

	if fc.IsSet("a") {
		reqs = append(reqs, requirementsFactory.NewMinAPIVersionRequirement("Option '-a'", cf.SpaceAppInstanceLimitMinimumApiVersion))
	}

	return reqs
}

func (cmd *UpdateSpaceQuota) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.spaceQuotaRepo = deps.RepoLocator.GetSpaceQuotaRepository()
	return cmd
}

func (cmd *UpdateSpaceQuota) Execute(c flags.FlagContext) {
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
				cmd.ui.Failed(T("Incorrect Usage") + "\n\n" + command_registry.Commands.CommandUsage("update-space-quota"))
			}
		}

		spaceQuota.InstanceMemoryLimit = memory
	}

	if c.String("m") != "" {
		memory, formatError := formatters.ToMegabytes(c.String("m"))

		if formatError != nil {
			cmd.ui.Failed(T("Incorrect Usage") + "\n\n" + command_registry.Commands.CommandUsage("update-space-quota"))
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

	if c.IsSet("a") {
		spaceQuota.AppInstanceLimit = c.Int("a")
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
