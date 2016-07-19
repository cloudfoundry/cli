package quota

import (
	"errors"
	"fmt"

	"encoding/json"

	"code.cloudfoundry.org/cli/cf"
	"code.cloudfoundry.org/cli/cf/api/quotas"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/flags"
	"code.cloudfoundry.org/cli/cf/formatters"
	. "code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
)

type UpdateQuota struct {
	ui        terminal.UI
	config    coreconfig.Reader
	quotaRepo quotas.QuotaRepository
}

func init() {
	commandregistry.Register(&UpdateQuota{})
}

func (cmd *UpdateQuota) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["allow-paid-service-plans"] = &flags.BoolFlag{Name: "allow-paid-service-plans", Usage: T("Can provision instances of paid service plans")}
	fs["disallow-paid-service-plans"] = &flags.BoolFlag{Name: "disallow-paid-service-plans", Usage: T("Cannot provision instances of paid service plans")}
	fs["i"] = &flags.StringFlag{ShortName: "i", Usage: T("Maximum amount of memory an application instance can have (e.g. 1024M, 1G, 10G)")}
	fs["m"] = &flags.StringFlag{ShortName: "m", Usage: T("Total amount of memory (e.g. 1024M, 1G, 10G)")}
	fs["n"] = &flags.StringFlag{ShortName: "n", Usage: T("New name")}
	fs["r"] = &flags.IntFlag{ShortName: "r", Usage: T("Total number of routes")}
	fs["s"] = &flags.IntFlag{ShortName: "s", Usage: T("Total number of service instances")}
	fs["a"] = &flags.IntFlag{ShortName: "a", Usage: T("Total number of application instances. -1 represents an unlimited amount.")}
	fs["reserved-route-ports"] = &flags.StringFlag{Name: "reserved-route-ports", Usage: T("Maximum number of routes that may be created with reserved ports")}

	return commandregistry.CommandMetadata{
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
			"[--allow-paid-service-plans | --disallow-paid-service-plans] ",
			fmt.Sprintf("[--reserved-route-ports %s] ", T("RESERVED_ROUTE_PORTS")),
		},
		Flags: fs,
	}
}

func (cmd *UpdateQuota) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + commandregistry.Commands.CommandUsage("update-quota"))
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

func (cmd *UpdateQuota) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.quotaRepo = deps.RepoLocator.GetQuotaRepository()
	return cmd
}

func (cmd *UpdateQuota) Execute(c flags.FlagContext) error {
	oldQuotaName := c.Args()[0]
	quota, err := cmd.quotaRepo.FindByName(oldQuotaName)

	if err != nil {
		return err
	}

	allowPaidServices := c.Bool("allow-paid-service-plans")
	disallowPaidServices := c.Bool("disallow-paid-service-plans")
	if allowPaidServices && disallowPaidServices {
		return errors.New(T("Please choose either allow or disallow. Both flags are not permitted to be passed in the same command."))
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
				return errors.New(T("Incorrect Usage") + "\n\n" + commandregistry.Commands.CommandUsage("update-quota"))
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
			return errors.New(T("Incorrect Usage") + "\n\n" + commandregistry.Commands.CommandUsage("update-quota"))
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

	if c.IsSet("reserved-route-ports") {
		quota.ReservedRoutePorts = json.Number(c.String("reserved-route-ports"))
	}

	cmd.ui.Say(T("Updating quota {{.QuotaName}} as {{.Username}}...", map[string]interface{}{
		"QuotaName": terminal.EntityNameColor(oldQuotaName),
		"Username":  terminal.EntityNameColor(cmd.config.Username()),
	}))

	err = cmd.quotaRepo.Update(quota)
	if err != nil {
		return err
	}
	cmd.ui.Ok()
	return nil
}
