package space

import (
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	"cf/terminal"
	"github.com/codegangsta/cli"
)

type ListSpaces struct {
	ui        terminal.UI
	config    *configuration.Configuration
	spaceRepo api.SpaceRepository
}

func NewListSpaces(ui terminal.UI, config *configuration.Configuration, spaceRepo api.SpaceRepository) (cmd ListSpaces) {
	cmd.ui = ui
	cmd.config = config
	cmd.spaceRepo = spaceRepo
	return
}

func (cmd ListSpaces) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
		reqFactory.NewTargetedOrgRequirement(),
	}
	return
}

func (cmd ListSpaces) Run(c *cli.Context) {
	cmd.ui.Say("Getting spaces in org %s as %s...\n",
		terminal.EntityNameColor(cmd.config.Organization.Name),
		terminal.EntityNameColor(cmd.config.Username()))

	stopChan := make(chan bool)
	defer close(stopChan)

	spacesChan, statusChan := cmd.spaceRepo.ListSpaces(stopChan)

	table := cmd.ui.Table([]string{"name"})
	noSpaces := true

	for spaces := range spacesChan {
		rows := [][]string{}
		for _, space := range spaces {
			rows = append(rows, []string{space.Name})
		}
		table.Print(rows)
		noSpaces = false
	}

	apiStatus := <-statusChan
	if apiStatus.IsNotSuccessful() {
		cmd.ui.Failed("Failed fetching spaces.\n%s", apiStatus.Message)
		return
	}

	if noSpaces {
		cmd.ui.Say("No spaces found")
	}
}
