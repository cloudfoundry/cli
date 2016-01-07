package quota

import (
	"fmt"
	"strconv"

	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/flags"

	"github.com/cloudfoundry/cli/cf/api/quotas"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/formatters"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type showQuota struct {
	ui        terminal.UI
	config    core_config.Reader
	quotaRepo quotas.QuotaRepository
}

func init() {
	command_registry.Register(&showQuota{})
}

func (cmd *showQuota) MetaData() command_registry.CommandMetadata {
	return command_registry.CommandMetadata{
		Name:        "quota",
		Usage:       T("CF_NAME quota QUOTA"),
		Description: T("Show quota info"),
	}
}

func (cmd *showQuota) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + command_registry.Commands.CommandUsage("quota"))
	}

	return []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}, nil
}

func (cmd *showQuota) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.quotaRepo = deps.RepoLocator.GetQuotaRepository()
	return cmd
}

func (cmd *showQuota) Execute(c flags.FlagContext) {
	quotaName := c.Args()[0]
	cmd.ui.Say(T("Getting quota {{.QuotaName}} info as {{.Username}}...", map[string]interface{}{"QuotaName": quotaName, "Username": cmd.config.Username()}))

	quota, err := cmd.quotaRepo.FindByName(quotaName)
	if err != nil {
		cmd.ui.Failed(err.Error())
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
	table := terminal.NewTable(cmd.ui, []string{"", ""})
	table.Add(T("Total Memory"), formatters.ByteSize(quota.MemoryLimit*formatters.MEGABYTE))
	table.Add(T("Instance Memory"), megabytes)
	table.Add(T("Routes"), fmt.Sprintf("%d", quota.RoutesLimit))
	table.Add(T("Services"), servicesLimit)
	table.Add(T("Paid service plans"), formatters.Allowed(quota.NonBasicServicesAllowed))
	table.Print()
}
