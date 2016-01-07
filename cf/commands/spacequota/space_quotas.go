package spacequota

import (
	"fmt"
	"strconv"

	"github.com/cloudfoundry/cli/cf/api/space_quotas"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/formatters"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
)

type ListSpaceQuotas struct {
	ui             terminal.UI
	config         core_config.Reader
	spaceQuotaRepo space_quotas.SpaceQuotaRepository
}

func init() {
	command_registry.Register(&ListSpaceQuotas{})
}

func (cmd *ListSpaceQuotas) MetaData() command_registry.CommandMetadata {
	return command_registry.CommandMetadata{
		Name:        "space-quotas",
		Description: T("List available space resource quotas"),
		Usage:       T("CF_NAME space-quotas"),
	}
}

func (cmd *ListSpaceQuotas) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(fc.Args()) != 0 {
		cmd.ui.Failed(T("Incorrect Usage. No argument required\n\n") + command_registry.Commands.CommandUsage("space-quotas"))
	}

	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedOrgRequirement(),
	}
	return
}

func (cmd *ListSpaceQuotas) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.spaceQuotaRepo = deps.RepoLocator.GetSpaceQuotaRepository()
	return cmd
}

func (cmd *ListSpaceQuotas) Execute(c flags.FlagContext) {
	cmd.ui.Say(T("Getting space quotas as {{.Username}}...", map[string]interface{}{"Username": terminal.EntityNameColor(cmd.config.Username())}))

	quotas, apiErr := cmd.spaceQuotaRepo.FindByOrg(cmd.config.OrganizationFields().Guid)

	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say("")

	table := terminal.NewTable(cmd.ui, []string{T("name"), T("total memory limit"), T("instance memory limit"), T("routes"), T("service instances"), T("paid service plans")})
	var megabytes string

	for _, quota := range quotas {
		if quota.InstanceMemoryLimit == -1 {
			megabytes = T("unlimited")
		} else {
			megabytes = formatters.ByteSize(quota.InstanceMemoryLimit * formatters.MEGABYTE)
		}

		servicesLimit := strconv.Itoa(quota.ServicesLimit)
		if servicesLimit == "-1" {
			servicesLimit = T("unlimited")
		}

		table.Add(
			quota.Name,
			formatters.ByteSize(quota.MemoryLimit*formatters.MEGABYTE),
			megabytes,
			fmt.Sprintf("%d", quota.RoutesLimit),
			fmt.Sprintf(servicesLimit),
			formatters.Allowed(quota.NonBasicServicesAllowed),
		)
	}

	table.Print()

}
