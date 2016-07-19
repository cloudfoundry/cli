package spacequota

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"code.cloudfoundry.org/cli/cf"
	"code.cloudfoundry.org/cli/cf/api/spacequotas"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/flags"
	"code.cloudfoundry.org/cli/cf/formatters"
	. "code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
)

type UpdateSpaceQuota struct {
	ui             terminal.UI
	config         coreconfig.Reader
	spaceQuotaRepo spacequotas.SpaceQuotaRepository
}

func init() {
	commandregistry.Register(&UpdateSpaceQuota{})
}

func (cmd *UpdateSpaceQuota) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["i"] = &flags.StringFlag{ShortName: "i", Usage: T("Maximum amount of memory an application instance can have (e.g. 1024M, 1G, 10G). -1 represents an unlimited amount.")}
	fs["m"] = &flags.StringFlag{ShortName: "m", Usage: T("Total amount of memory a space can have (e.g. 1024M, 1G, 10G)")}
	fs["n"] = &flags.StringFlag{ShortName: "n", Usage: T("New name")}
	fs["r"] = &flags.IntFlag{ShortName: "r", Usage: T("Total number of routes")}
	fs["s"] = &flags.IntFlag{ShortName: "s", Usage: T("Total number of service instances")}
	fs["allow-paid-service-plans"] = &flags.BoolFlag{Name: "allow-paid-service-plans", Usage: T("Can provision instances of paid service plans")}
	fs["disallow-paid-service-plans"] = &flags.BoolFlag{Name: "disallow-paid-service-plans", Usage: T("Can not provision instances of paid service plans")}
	fs["a"] = &flags.IntFlag{ShortName: "a", Usage: T("Total number of application instances. -1 represents an unlimited amount.")}
	fs["reserved-route-ports"] = &flags.IntFlag{Name: "reserved-route-ports", Usage: T("Maximum number of routes that may be created with reserved ports")}

	return commandregistry.CommandMetadata{
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
			"[--allow-paid-service-plans | --disallow-paid-service-plans] ",
			fmt.Sprintf("[--reserved-route-ports %s] ", T("RESERVED_ROUTE_PORTS")),
		},
		Flags: fs,
	}
}

func (cmd *UpdateSpaceQuota) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + commandregistry.Commands.CommandUsage("update-space-quota"))
		return nil, fmt.Errorf("Incorrect usage: %d arguments of %d required", len(fc.Args()), 1)
	}

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedOrgRequirement(),
	}

	if fc.IsSet("a") {
		reqs = append(reqs, requirementsFactory.NewMinAPIVersionRequirement("Option '-a'", cf.SpaceAppInstanceLimitMinimumAPIVersion))
	}

	return reqs, nil
}

func (cmd *UpdateSpaceQuota) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.spaceQuotaRepo = deps.RepoLocator.GetSpaceQuotaRepository()
	return cmd
}

func (cmd *UpdateSpaceQuota) Execute(c flags.FlagContext) error {
	name := c.Args()[0]

	spaceQuota, err := cmd.spaceQuotaRepo.FindByName(name)
	if err != nil {
		return err
	}

	allowPaidServices := c.Bool("allow-paid-service-plans")
	disallowPaidServices := c.Bool("disallow-paid-service-plans")
	if allowPaidServices && disallowPaidServices {
		return errors.New(T("Please choose either allow or disallow. Both flags are not permitted to be passed in the same command."))
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
				return errors.New(T("Incorrect Usage") + "\n\n" + commandregistry.Commands.CommandUsage("update-space-quota"))
			}
		}

		spaceQuota.InstanceMemoryLimit = memory
	}

	if c.String("m") != "" {
		memory, formatError := formatters.ToMegabytes(c.String("m"))

		if formatError != nil {
			return errors.New(T("Incorrect Usage") + "\n\n" + commandregistry.Commands.CommandUsage("update-space-quota"))
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

	if c.IsSet("reserved-route-ports") {
		spaceQuota.ReservedRoutePortsLimit = json.Number(strconv.Itoa(c.Int("reserved-route-ports")))
	}

	cmd.ui.Say(T("Updating space quota {{.Quota}} as {{.Username}}...",
		map[string]interface{}{
			"Quota":    terminal.EntityNameColor(name),
			"Username": terminal.EntityNameColor(cmd.config.Username()),
		}))

	err = cmd.spaceQuotaRepo.Update(spaceQuota)
	if err != nil {
		return err
	}

	cmd.ui.Ok()
	return nil
}
