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

type SpaceQuota struct {
	ui             terminal.UI
	config         configuration.Reader
	spaceQuotaRepo space_quotas.SpaceQuotaRepository
}

func NewSpaceQuota(ui terminal.UI, config configuration.Reader, spaceQuotaRepo space_quotas.SpaceQuotaRepository) (cmd *SpaceQuota) {
	return &SpaceQuota{
		ui:             ui,
		config:         config,
		spaceQuotaRepo: spaceQuotaRepo,
	}
}

func (cmd *SpaceQuota) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "space-quota",
		Description: T("Show space quota info"),
		Usage:       T("CF_NAME space-quota SPACE_QUOTA_NAME"),
	}
}

func (cmd *SpaceQuota) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		cmd.ui.FailWithUsage(c)
	}

	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedOrgRequirement(),
	}
	return
}

func (cmd *SpaceQuota) Run(c *cli.Context) {
	name := c.Args()[0]

	cmd.ui.Say(T("Getting space quota {{.Quota}} info as {{.Username}}...",
		map[string]interface{}{
			"Quota":    terminal.EntityNameColor(name),
			"Username": terminal.EntityNameColor(cmd.config.Username()),
		}))

	spaceQuota, apiErr := cmd.spaceQuotaRepo.FindByName(name)

	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say("")
	var megabytes string

	table := terminal.NewTable(cmd.ui, []string{"", ""})
	table.Add(T("total memory limit"), formatters.ByteSize(spaceQuota.MemoryLimit*formatters.MEGABYTE))
	if spaceQuota.InstanceMemoryLimit == -1 {
		megabytes = "-1"
	} else {
		megabytes = formatters.ByteSize(spaceQuota.InstanceMemoryLimit * formatters.MEGABYTE)
	}
	table.Add(T("instance memory limit"), megabytes)
	table.Add(T("routes"), fmt.Sprintf("%d", spaceQuota.RoutesLimit))
	table.Add(T("services"), fmt.Sprintf("%d", spaceQuota.ServicesLimit))
	table.Add(T("non basic services"), formatters.Allowed(spaceQuota.NonBasicServicesAllowed))

	table.Print()

}
