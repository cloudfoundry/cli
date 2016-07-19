package quota

import (
	"fmt"

	"encoding/json"

	"code.cloudfoundry.org/cli/cf"
	"code.cloudfoundry.org/cli/cf/api/quotas"
	"code.cloudfoundry.org/cli/cf/api/resources"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/flags"
	"code.cloudfoundry.org/cli/cf/formatters"
	. "code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
)

type CreateQuota struct {
	ui        terminal.UI
	config    coreconfig.Reader
	quotaRepo quotas.QuotaRepository
}

func init() {
	commandregistry.Register(&CreateQuota{})
}

func (cmd *CreateQuota) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["allow-paid-service-plans"] = &flags.BoolFlag{Name: "allow-paid-service-plans", Usage: T("Can provision instances of paid service plans")}
	fs["i"] = &flags.StringFlag{ShortName: "i", Usage: T("Maximum amount of memory an application instance can have (e.g. 1024M, 1G, 10G). -1 represents an unlimited amount.")}
	fs["m"] = &flags.StringFlag{ShortName: "m", Usage: T("Total amount of memory (e.g. 1024M, 1G, 10G)")}
	fs["r"] = &flags.IntFlag{ShortName: "r", Usage: T("Total number of routes")}
	fs["s"] = &flags.IntFlag{ShortName: "s", Usage: T("Total number of service instances")}
	fs["a"] = &flags.IntFlag{ShortName: "a", Usage: T("Total number of application instances. -1 represents an unlimited amount. (Default: unlimited)")}
	fs["reserved-route-ports"] = &flags.StringFlag{Name: "reserved-route-ports", Usage: T("Maximum number of routes that may be created with reserved ports (Default: 0)")}

	return commandregistry.CommandMetadata{
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
			"[--allow-paid-service-plans] ",
			fmt.Sprintf("[--reserved-route-ports %s] ", T("RESERVED_ROUTE_PORTS")),
		},
		Flags: fs,
	}
}

func (cmd *CreateQuota) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + commandregistry.Commands.CommandUsage("create-quota"))
		return nil, fmt.Errorf("Incorrect usage: %d arguments of %d required", len(fc.Args()), 1)
	}

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}

	if fc.IsSet("a") {
		reqs = append(reqs, requirementsFactory.NewMinAPIVersionRequirement("Option '-a'", cf.OrgAppInstanceLimitMinimumAPIVersion))
	}

	if fc.IsSet("reserved-route-ports") {
		reqs = append(reqs, requirementsFactory.NewMinAPIVersionRequirement("Option '--reserved-route-ports'", cf.ReservedRoutePortsMinimumAPIVersion))
	}

	return reqs, nil
}

func (cmd *CreateQuota) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.quotaRepo = deps.RepoLocator.GetQuotaRepository()
	return cmd
}

func (cmd *CreateQuota) Execute(context flags.FlagContext) error {
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
			return errors.New(T("Invalid memory limit: {{.MemoryLimit}}\n{{.Err}}", map[string]interface{}{"MemoryLimit": memoryLimit, "Err": err}))
		}

		quota.MemoryLimit = parsedMemory
	}

	instanceMemoryLimit := context.String("i")
	if instanceMemoryLimit == "-1" || instanceMemoryLimit == "" {
		quota.InstanceMemoryLimit = -1
	} else {
		parsedMemory, errr := formatters.ToMegabytes(instanceMemoryLimit)
		if errr != nil {
			return errors.New(T("Invalid instance memory limit: {{.MemoryLimit}}\n{{.Err}}", map[string]interface{}{"MemoryLimit": instanceMemoryLimit, "Err": errr}))
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

	if context.IsSet("reserved-route-ports") {
		quota.ReservedRoutePorts = json.Number(context.String("reserved-route-ports"))
	}

	err := cmd.quotaRepo.Create(quota)

	httpErr, ok := err.(errors.HTTPError)
	if ok && httpErr.ErrorCode() == errors.QuotaDefinitionNameTaken {
		cmd.ui.Ok()
		cmd.ui.Warn(T("Quota Definition {{.QuotaName}} already exists", map[string]interface{}{"QuotaName": quota.Name}))
		return nil
	}

	if err != nil {
		return err
	}

	cmd.ui.Ok()
	return nil
}
