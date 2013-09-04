package commands

import (
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	term "cf/terminal"
	"github.com/codegangsta/cli"
	"strings"
)

type Services struct {
	ui        term.UI
	config    *configuration.Configuration
	spaceRepo api.SpaceRepository
}

func NewServices(ui term.UI, config *configuration.Configuration, spaceRepo api.SpaceRepository) (cmd Services) {
	cmd.ui = ui
	cmd.config = config
	cmd.spaceRepo = spaceRepo
	return
}

func (cmd Services) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []Requirement, err error) {
	return
}

func (cmd Services) Run(c *cli.Context) {
	cmd.ui.Say("Getting services in %s", cmd.config.Space.Name)

	space, err := cmd.spaceRepo.GetSummary(cmd.config)

	if err != nil {
		cmd.ui.Failed("Error loading applications", err)
		return
	}

	cmd.ui.Ok()

	cmd.ui.Say("name \t service \t provider \t version \t plan \t bound apps")

	for _, instance := range space.ServiceInstances {
		cmd.ui.Say(
			"%s \t %s \t %s \t %s \t %s \t %s",
			term.Cyan(instance.Name),
			instance.ServicePlan.ServiceOffering.Label,
			instance.ServicePlan.ServiceOffering.Provider,
			instance.ServicePlan.ServiceOffering.Version,
			instance.ServicePlan.Name,
			strings.Join(instance.ApplicationNames, ", "),
		)
	}
}
