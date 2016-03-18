package quota

import (
	"fmt"

	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/api/quotas"
	"github.com/cloudfoundry/cli/cf/api/resources"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/formatters"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
)

type CreateQuota struct {
	ui        terminal.UI
	config    core_config.Reader
	quotaRepo quotas.QuotaRepository
}

func init() {
	command_registry.Register(&CreateQuota{})
}

func (cmd *CreateQuota) MetaData() command_registry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["allow-paid-service-plans"] = &flags.BoolFlag{Name: "allow-paid-service-plans", Usage: T("Can provision instances of paid service plans")}
	fs["i"] = &flags.StringFlag{ShortName: "i", Usage: T("Maximum amount of memory an application instance can have (e.g. 1024M, 1G, 10G). -1 represents an unlimited amount.")}
	fs["m"] = &flags.StringFlag{ShortName: "m", Usage: T("Total amount of memory (e.g. 1024M, 1G, 10G)")}
	fs["r"] = &flags.IntFlag{ShortName: "r", Usage: T("Total number of routes")}
	fs["s"] = &flags.IntFlag{ShortName: "s", Usage: T("Total number of service instances")}
	fs["a"] = &flags.IntFlag{ShortName: "a", Usage: T("Total number of application instances. -1 represents an unlimited amount. (Default: unlimited)")}

	return command_registry.CommandMetadata{
		Name:        "create-quota",
		Description: T("Define a new resource quota"),
		Usage: []string{
			"CF_NAME create-quota ",
			T("QUOTA"),
			" ",
			fmt.Sprintf("[-m %s] ", T("TOTAL_MEMORY")),
			fmt.Sprintf("[-i %s] ", T("INSTANCE_MEMORY")),
			fmt.Sprintf("[-r %s] ", T("ROUTES")),
			fmt.Sprintf("[-s %s] ", T("SERVICE_INSTANCES")),
			fmt.Sprintf("[-a %s] ", T("APP_INSTANCES")),
			"[--allow-paid-service-plans]",
		},
		Flags: fs,
	}
}

func (cmd *CreateQuota) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + command_registry.Commands.CommandUsage("create-quota"))
	}

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}

	if fc.IsSet("a") {
		reqs = append(reqs, requirementsFactory.NewMinAPIVersionRequirement("Option '-a'", cf.OrgAppInstanceLimitMinimumApiVersion))
	}

	return reqs
}

func (cmd *CreateQuota) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.quotaRepo = deps.RepoLocator.GetQuotaRepository()
	return cmd
}

func (cmd *CreateQuota) Execute(context flags.FlagContext) {
	name := context.Args()[0]

	cmd.ui.Say(T("Creating quota {{.QuotaName}} as {{.Username}}...", map[string]interface{}{
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
			cmd.ui.Failed(T("Invalid memory limit: {{.MemoryLimit}}\n{{.Err}}", map[string]interface{}{"MemoryLimit": memoryLimit, "Err": err}))
		}

		quota.MemoryLimit = parsedMemory
	}

	instanceMemoryLimit := context.String("i")
	if instanceMemoryLimit == "-1" || instanceMemoryLimit == "" {
		quota.InstanceMemoryLimit = -1
	} else {
		parsedMemory, errr := formatters.ToMegabytes(instanceMemoryLimit)
		if errr != nil {
			cmd.ui.Failed(T("Invalid instance memory limit: {{.MemoryLimit}}\n{{.Err}}", map[string]interface{}{"MemoryLimit": instanceMemoryLimit, "Err": errr}))
		}
		quota.InstanceMemoryLimit = parsedMemory
	}

	if context.IsSet("r") {
		quota.RoutesLimit = context.Int("r")
	}

	if context.IsSet("s") {
		quota.ServicesLimit = context.Int("s")
	}

	if context.IsSet("a") {
		quota.AppInstanceLimit = context.Int("a")
	} else {
		quota.AppInstanceLimit = resources.UnlimitedAppInstances
	}

	if context.IsSet("allow-paid-service-plans") {
		quota.NonBasicServicesAllowed = true
	}

	err := cmd.quotaRepo.Create(quota)

	httpErr, ok := err.(errors.HttpError)
	if ok && httpErr.ErrorCode() == errors.QuotaDefinitionNameTaken {
		cmd.ui.Ok()
		cmd.ui.Warn(T("Quota Definition {{.QuotaName}} already exists", map[string]interface{}{"QuotaName": quota.Name}))
		return
	}

	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	cmd.ui.Ok()
}
