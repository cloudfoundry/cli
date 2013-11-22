package space

import (
	"cf/configuration"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
	"strings"
)

type ShowSpace struct {
	ui       terminal.UI
	config   *configuration.Configuration
	spaceReq requirements.SpaceRequirement
}

func NewShowSpace(ui terminal.UI, config *configuration.Configuration) (cmd *ShowSpace) {
	cmd = new(ShowSpace)
	cmd.ui = ui
	cmd.config = config
	return
}

func (cmd *ShowSpace) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "space")
		return
	}

	cmd.spaceReq = reqFactory.NewSpaceRequirement(c.Args()[0])
	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
		reqFactory.NewTargetedOrgRequirement(),
		cmd.spaceReq,
	}
	return
}

func (cmd *ShowSpace) Run(c *cli.Context) {
	space := cmd.spaceReq.GetSpace()
	cmd.ui.Say("Getting info for space %s in org %s as %s...",
		terminal.EntityNameColor(space.Name),
		terminal.EntityNameColor(space.Organization.Name),
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
