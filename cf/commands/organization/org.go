package organization

import (
	"fmt"
	"strings"

	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/formatters"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type ShowOrg struct {
	ui     terminal.UI
	config core_config.Reader
	orgReq requirements.OrganizationRequirement
}

func NewShowOrg(ui terminal.UI, config core_config.Reader) (cmd *ShowOrg) {
	cmd = new(ShowOrg)
	cmd.ui = ui
	cmd.config = config
	return
}

func (cmd *ShowOrg) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "org",
		Description: T("Show org info"),
		Usage:       T("CF_NAME org ORG"),
		Flags: []cli.Flag{
			cli.BoolFlag{Name: "guid", Usage: T("Retrieve and display the given org's guid.  All other output for the org is suppressed.")},
		},
	}
}

func (cmd *ShowOrg) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		cmd.ui.FailWithUsage(c)
	}

	if cmd.orgReq == nil {
		cmd.orgReq = requirementsFactory.NewOrganizationRequirement(c.Args()[0])
	} else {
		cmd.orgReq.SetOrganizationName(c.Args()[0])
	}
	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		cmd.orgReq,
	}

	return
}

func (cmd *ShowOrg) Run(c *cli.Context) {
	org := cmd.orgReq.GetOrganization()

	if c.Bool("guid") {
		cmd.ui.Say(org.Guid)
	} else {
		cmd.ui.Say(T("Getting info for org {{.OrgName}} as {{.Username}}...",
			map[string]interface{}{
				"OrgName":  terminal.EntityNameColor(org.Name),
				"Username": terminal.EntityNameColor(cmd.config.Username())}))
		cmd.ui.Ok()
		cmd.ui.Say("")

		table := terminal.NewTable(cmd.ui, []string{terminal.EntityNameColor(org.Name) + ":", "", ""})

		domains := []string{}
		for _, domain := range org.Domains {
			domains = append(domains, domain.Name)
		}

		spaces := []string{}
		for _, space := range org.Spaces {
			spaces = append(spaces, space.Name)
		}

		spaceQuotas := []string{}
		for _, spaceQuota := range org.SpaceQuotas {
			spaceQuotas = append(spaceQuotas, spaceQuota.Name)
		}

		quota := org.QuotaDefinition
		orgQuota := fmt.Sprintf(T("{{.QuotaName}} ({{.MemoryLimit}}M memory limit, {{.InstanceMemoryLimit}} instance memory limit, {{.RoutesLimit}} routes, {{.ServicesLimit}} services, paid services {{.NonBasicServicesAllowed}})",
			map[string]interface{}{
				"QuotaName":               quota.Name,
				"MemoryLimit":             quota.MemoryLimit,
				"InstanceMemoryLimit":     formatters.InstanceMemoryLimit(quota.InstanceMemoryLimit),
				"RoutesLimit":             quota.RoutesLimit,
				"ServicesLimit":           quota.ServicesLimit,
				"NonBasicServicesAllowed": formatters.Allowed(quota.NonBasicServicesAllowed)}))

		table.Add("", T("domains:"), terminal.EntityNameColor(strings.Join(domains, ", ")))
		table.Add("", T("quota:"), terminal.EntityNameColor(orgQuota))
		table.Add("", T("spaces:"), terminal.EntityNameColor(strings.Join(spaces, ", ")))
		table.Add("", T("space quotas:"), terminal.EntityNameColor(strings.Join(spaceQuotas, ", ")))

		table.Print()
	}
}
