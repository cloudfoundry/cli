package space

import (
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
	cmd.ui.Say("\n%s:", terminal.EntityNameColor(space.Name))
	cmd.ui.Say(T("  Org: {{.OrgName}}", map[string]interface{}{"OrgName": terminal.EntityNameColor(space.Organization.Name)}))

	apps := []string{}
	for _, app := range space.Applications {
		apps = append(apps, app.Name)
	}
	cmd.ui.Say(T("  Apps: {{.ApplicationNames}}", map[string]interface{}{"ApplicationNames": terminal.EntityNameColor(strings.Join(apps, ", "))}))

	domains := []string{}
	for _, domain := range space.Domains {
		domains = append(domains, domain.Name)
	}
	cmd.ui.Say(T("  Domains: {{.DomainNames}}", map[string]interface{}{"DomainNames": terminal.EntityNameColor(strings.Join(domains, ", "))}))

	services := []string{}
	for _, service := range space.ServiceInstances {
		services = append(services, service.Name)
	}
	cmd.ui.Say(T("  Services: {{.ServiceNames}}", map[string]interface{}{"ServiceNames": terminal.EntityNameColor(strings.Join(services, ", "))}))
}
