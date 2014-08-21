package spacequota

import (
	"fmt"

	"github.com/cloudfoundry/cli/cf/api/space_quotas"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/formatters"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type ListSpaceQuotas struct {
	ui             terminal.UI
	config         configuration.Reader
	spaceQuotaRepo space_quotas.SpaceQuotaRepository
}

func NewListSpaceQuotas(ui terminal.UI, config configuration.Reader, spaceQuotaRepo space_quotas.SpaceQuotaRepository) (cmd *ListSpaceQuotas) {
	return &ListSpaceQuotas{
		ui:             ui,
		config:         config,
		spaceQuotaRepo: spaceQuotaRepo,
	}
}

func (cmd *ListSpaceQuotas) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "space-quotas",
		Description: T("List available space resource quotas"),
		Usage:       T("CF_NAME space-quotas"),
	}
}

func (cmd *ListSpaceQuotas) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedOrgRequirement(),
	}
	return
}

func (cmd *ListSpaceQuotas) Run(c *cli.Context) {
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
			megabytes = "-1"
		} else {
			megabytes = formatters.ByteSize(quota.InstanceMemoryLimit * formatters.MEGABYTE)
		}

		table.Add(
			quota.Name,
			formatters.ByteSize(quota.MemoryLimit*formatters.MEGABYTE),
			megabytes,
			fmt.Sprintf("%d", quota.RoutesLimit),
			fmt.Sprintf("%d", quota.ServicesLimit),
			formatters.Allowed(quota.NonBasicServicesAllowed),
		)
	}

	table.Print()

}
