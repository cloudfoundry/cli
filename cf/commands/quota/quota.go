package quota

import (
	"fmt"
	"strconv"

	"code.cloudfoundry.org/cli/cf/flags"
	. "code.cloudfoundry.org/cli/cf/i18n"

	"code.cloudfoundry.org/cli/cf/api/quotas"
	"code.cloudfoundry.org/cli/cf/api/resources"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/formatters"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
)

type showQuota struct {
	ui        terminal.UI
	config    coreconfig.Reader
	quotaRepo quotas.QuotaRepository
}

func init() {
	commandregistry.Register(&showQuota{})
}

func (cmd *showQuota) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name: "quota",
		Usage: []string{
			T("CF_NAME quota QUOTA"),
		},
		Description: T("Show quota info"),
	}
}

func (cmd *showQuota) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + commandregistry.Commands.CommandUsage("quota"))
		return nil, fmt.Errorf("Incorrect usage: %d arguments of %d required", len(fc.Args()), 1)
	}

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}

	return reqs, nil
}

func (cmd *showQuota) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.quotaRepo = deps.RepoLocator.GetQuotaRepository()
	return cmd
}

func (cmd *showQuota) Execute(c flags.FlagContext) error {
	quotaName := c.Args()[0]
	cmd.ui.Say(T("Getting quota {{.QuotaName}} info as {{.Username}}...", map[string]interface{}{"QuotaName": quotaName, "Username": cmd.config.Username()}))

	quota, err := cmd.quotaRepo.FindByName(quotaName)
	if err != nil {
		return err
	}

	cmd.ui.Ok()

	var megabytes string
	if quota.InstanceMemoryLimit == -1 {
		megabytes = T("unlimited")
	} else {
		megabytes = formatters.ByteSize(quota.InstanceMemoryLimit * formatters.MEGABYTE)
	}

	servicesLimit := strconv.Itoa(quota.ServicesLimit)
	if servicesLimit == "-1" {
		servicesLimit = T("unlimited")
	}

	appInstanceLimit := strconv.Itoa(quota.AppInstanceLimit)
	if quota.AppInstanceLimit == resources.UnlimitedAppInstances {
		appInstanceLimit = T("unlimited")
	}

	reservedRoutePorts := string(quota.ReservedRoutePorts)
	if reservedRoutePorts == resources.UnlimitedReservedRoutePorts {
		reservedRoutePorts = T("unlimited")
	}

	table := cmd.ui.Table([]string{"", ""})
	table.Add(T("Total Memory"), formatters.ByteSize(quota.MemoryLimit*formatters.MEGABYTE))
	table.Add(T("Instance Memory"), megabytes)
	table.Add(T("Routes"), fmt.Sprint(quota.RoutesLimit))
	table.Add(T("Services"), servicesLimit)
	table.Add(T("Paid service plans"), formatters.Allowed(quota.NonBasicServicesAllowed))
	table.Add(T("App instance limit"), appInstanceLimit)
	if reservedRoutePorts != "" {
		table.Add(T("Reserved Route Ports"), reservedRoutePorts)
	}
	err = table.Print()
	if err != nil {
		return err
	}
	return nil
}
