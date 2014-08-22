package space

import (
	"fmt"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"strings"

	"github.com/cloudfoundry/cli/cf/api/space_quotas"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/formatters"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type ShowSpace struct {
	ui        terminal.UI
	config    configuration.Reader
	spaceReq  requirements.SpaceRequirement
	quotaRepo space_quotas.SpaceQuotaRepository
}

func NewShowSpace(ui terminal.UI, config configuration.Reader, quotaRepo space_quotas.SpaceQuotaRepository) (cmd *ShowSpace) {
	cmd = new(ShowSpace)
	cmd.ui = ui
	cmd.config = config
	cmd.quotaRepo = quotaRepo
	return
}

func (cmd *ShowSpace) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "space",
		Description: T("Show space info"),
		Usage:       T("CF_NAME space SPACE"),
	}
}

func (cmd *ShowSpace) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		cmd.ui.FailWithUsage(c)
	}

	cmd.spaceReq = requirementsFactory.NewSpaceRequirement(c.Args()[0])
	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedOrgRequirement(),
		cmd.spaceReq,
	}
	return
}

func (cmd *ShowSpace) Run(c *cli.Context) {
	space := cmd.spaceReq.GetSpace()
	cmd.ui.Say(T("Getting info for space {{.TargetSpace}} in org {{.OrgName}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"TargetSpace": terminal.EntityNameColor(space.Name),
			"OrgName":     terminal.EntityNameColor(space.Organization.Name),
			"CurrentUser": terminal.EntityNameColor(cmd.config.Username()),
		}))

	quotaString := cmd.quotaString(space)

	cmd.ui.Ok()
	cmd.ui.Say("")
	cmd.ui.Say(terminal.EntityNameColor(space.Name) + ":")

	table := terminal.NewTable(cmd.ui, []string{"", "", ""})
	table.Add("", T("Org:"), terminal.EntityNameColor(space.Organization.Name))

	apps := []string{}
	for _, app := range space.Applications {
		apps = append(apps, terminal.EntityNameColor(app.Name))
	}
	table.Add("", T("Apps:"), strings.Join(apps, ", "))

	domains := []string{}
	for _, domain := range space.Domains {
		domains = append(domains, terminal.EntityNameColor(domain.Name))
	}
	table.Add("", T("Domains:"), strings.Join(domains, ", "))

	services := []string{}
	for _, service := range space.ServiceInstances {
		services = append(services, terminal.EntityNameColor(service.Name))
	}
	table.Add("", T("Services:"), strings.Join(services, ", "))

	securityGroups := []string{}
	for _, group := range space.SecurityGroups {
		securityGroups = append(securityGroups, terminal.EntityNameColor(group.Name))
	}
	table.Add("", T("Security Groups:"), strings.Join(securityGroups, ", "))

	table.Add("", T("Space Quota:"))

	table.Print()
}

func (cmd *ShowSpace) quotaString(space models.Space) string {
	var instance_memory string

	if space.SpaceQuotaGuid == "" {
		return ""
	}

	quota, err := cmd.quotaRepo.FindByGuid(space.SpaceQuotaGuid)
	if err != nil {
		cmd.ui.Failed(err.Error())
		return ""
	}

	if quota.InstanceMemoryLimit == -1 {
		instance_memory = "-1"
	} else {
		instance_memory = formatters.ByteSize(quota.InstanceMemoryLimit * formatters.MEGABYTE)
	}
	memory := formatters.ByteSize(quota.MemoryLimit * formatters.MEGABYTE)

	spaceQuota := fmt.Sprintf("%s (%s memory limit, %s instance memory limit, %d routes, %d services, paid services %s)", quota.Name, memory, instance_memory, quota.RoutesLimit, quota.ServicesLimit, formatters.Allowed(quota.NonBasicServicesAllowed))
	//	spaceQuota := fmt.Sprintf(T("{{.QuotaName}} ({{.MemoryLimit}} memory limit, {{.InstanceMemoryLimit}} instance memory limit, {{.RoutesLimit}} routes, {{.ServicesLimit}} services, paid services {{.NonBasicServicesAllowed}})",
	//		map[string]interface{}{
	//			"QuotaName":               quota.Name,
	//			"MemoryLimit":             memory,
	//			"InstanceMemoryLimit":     instance_memory,
	//			"RoutesLimit":             quota.RoutesLimit,
	//			"ServicesLimit":           quota.ServicesLimit,
	//			"NonBasicServicesAllowed": formatters.Allowed(quota.NonBasicServicesAllowed)}))

	return spaceQuota
}
