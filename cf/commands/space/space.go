package space

import (
	. "github.com/cloudfoundry/cli/cf/i18n"
	"strings"

	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type ShowSpace struct {
	ui       terminal.UI
	config   configuration.Reader
	spaceReq requirements.SpaceRequirement
}

func NewShowSpace(ui terminal.UI, config configuration.Reader) (cmd *ShowSpace) {
	cmd = new(ShowSpace)
	cmd.ui = ui
	cmd.config = config
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

	table.Print()
}
