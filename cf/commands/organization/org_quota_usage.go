package organization

import (
	"fmt"
	"strconv"

	"github.com/cloudfoundry/cli/cf/api/organizations"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/formatters"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type GetQuotaUsage struct {
	ui      terminal.UI
	config  core_config.Reader
	orgRepo organizations.OrganizationRepository
}

func NewQuotaUsage(ui terminal.UI, config core_config.Reader, orgRepo organizations.OrganizationRepository) (cmd *GetQuotaUsage) {
	cmd = new(GetQuotaUsage)
	cmd.ui = ui
	cmd.config = config
	cmd.orgRepo = orgRepo
	return
}

func (cmd *GetQuotaUsage) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "org-quota-usage",
		Description: T("Get quota usage for org"),
		Usage:       T("CF_NAME org-quota-usage ORG_NAME \n\n") + T("TIP:\n") + T("View all organizations with 'CF_NAME orgs'"),
	}
}

func (cmd *GetQuotaUsage) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		cmd.ui.FailWithUsage(c)
	}

	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}
	return
}

func (cmd *GetQuotaUsage) Run(c *cli.Context) {
	orgName := c.Args()[0]

	cmd.ui.Say(T("Getting quota usage info for org {{.OrgName}} as {{.Username}}...",
		map[string]interface{}{
			"OrgName":  terminal.EntityNameColor(orgName),
			"Username": terminal.EntityNameColor(cmd.config.Username())}))

	org, err := cmd.orgRepo.FindByName(orgName)

	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	quotaUsage, apiErr := cmd.orgRepo.GetOrganizationQuotaUsage(org.Guid)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
	}

	servicesLimit := strconv.Itoa(quotaUsage.ServicesLimit)
	if servicesLimit == "-1" {
		servicesLimit = T("unlimited")
	}

	table := terminal.NewTable(cmd.ui, []string{"", ""})
	table.Add(T("Routes"), fmt.Sprintf("%d/%d", quotaUsage.OrgUsage.Routes, quotaUsage.RoutesLimit))
	table.Add(T("Services"), fmt.Sprintf("%s/%s", strconv.Itoa(quotaUsage.OrgUsage.Services), servicesLimit))
	table.Add(T("Memory"), fmt.Sprintf("%s/%s", formatters.ByteSize(quotaUsage.OrgUsage.Memory*formatters.MEGABYTE), formatters.ByteSize(quotaUsage.MemoryLimit*formatters.MEGABYTE)))
	table.Print()

	cmd.ui.Ok()
}
