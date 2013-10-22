package space

import (
	"cf/configuration"
	"cf/requirements"
	"cf/terminal"
	"github.com/codegangsta/cli"
	"strings"
)

type ShowSpace struct {
	ui     terminal.UI
	config *configuration.Configuration
}

func NewShowSpace(ui terminal.UI, config *configuration.Configuration) (cmd *ShowSpace) {
	cmd = new(ShowSpace)
	cmd.ui = ui
	cmd.config = config
	return
}

func (cmd *ShowSpace) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {

	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
		reqFactory.NewTargetedSpaceRequirement(),
	}
	return
}

func (cmd *ShowSpace) Run(c *cli.Context) {
	space := cmd.config.Space
	cmd.ui.Say("Getting info for space %s in org %s as %s...",
		terminal.EntityNameColor(space.Name),
		terminal.EntityNameColor(cmd.config.Organization.Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)
	cmd.ui.Ok()
	cmd.ui.Say("\n%s:", terminal.EntityNameColor(space.Name))
	cmd.ui.Say("  Org: %s", terminal.EntityNameColor(space.Organization.Name))

	apps := []string{}
	for _, app := range space.Applications {
		apps = append(apps, app.Name)
	}
	cmd.ui.Say("  Apps: %s", terminal.EntityNameColor(strings.Join(apps, ", ")))

	domains := []string{}
	for _, domain := range space.Domains {
		domains = append(domains, domain.Name)
	}
	cmd.ui.Say("  Domains: %s", terminal.EntityNameColor(strings.Join(domains, ", ")))

	services := []string{}
	for _, service := range space.ServiceInstances {
		services = append(services, service.Name)
	}
	cmd.ui.Say("  Services: %s", terminal.EntityNameColor(strings.Join(services, ", ")))
}
