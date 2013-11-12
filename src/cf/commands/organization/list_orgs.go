package organization

import (
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	"cf/terminal"
	"github.com/codegangsta/cli"
)

type ListOrgs struct {
	ui      terminal.UI
	config  *configuration.Configuration
	orgRepo api.OrganizationRepository
}

func NewListOrgs(ui terminal.UI, config *configuration.Configuration, orgRepo api.OrganizationRepository) (cmd ListOrgs) {
	cmd.ui = ui
	cmd.config = config
	cmd.orgRepo = orgRepo
	return
}

func (cmd ListOrgs) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
	}
	return
}

func (cmd ListOrgs) Run(c *cli.Context) {
	cmd.ui.Say("Getting orgs as %s...\n", terminal.EntityNameColor(cmd.config.Username()))

	stopChan := make(chan bool)
	defer close(stopChan)

	orgsChan, statusChan := cmd.orgRepo.ListOrgs(stopChan)

	table := cmd.ui.Table([]string{"name"})
	noOrgs := true

	for orgs := range orgsChan {
		rows := [][]string{}
		for i := len(orgs) - 1; i >= 0; i-- {
			org := orgs[i]
			rows = append(rows, []string{org.Name})
		}
		table.Print(rows)
		noOrgs = false
	}

	apiStatus := <-statusChan
	if apiStatus.IsNotSuccessful() {
		cmd.ui.Failed("Failed fetching orgs.\n%s", apiStatus.Message)
		return
	}

	if noOrgs {
		cmd.ui.Say("No orgs found")
	}
}
