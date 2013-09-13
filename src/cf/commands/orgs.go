package commands

import (
	"cf/api"
	"cf/requirements"
	term "cf/terminal"
	"github.com/codegangsta/cli"
)

type ListOrganizations struct {
	ui      term.UI
	orgRepo api.OrganizationRepository
}

func NewListOrganizations(ui term.UI, orgRepo api.OrganizationRepository) (cmd ListOrganizations) {
	cmd.ui = ui
	cmd.orgRepo = orgRepo
	return
}

func (cmd ListOrganizations) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	return
}

func (cmd ListOrganizations) Run(c *cli.Context) {
	cmd.ui.Say("Getting organizations...")

	orgs, err := cmd.orgRepo.FindAll()
	if err != nil {
		cmd.ui.Failed("Error loading organizations", err)
		return
	}

	cmd.ui.Ok()

	table := [][]string{
		[]string{"name"},
	}

	for _, org := range orgs {
		table = append(table, []string{
			org.Name,
		})
	}

	cmd.ui.DisplayTable(table, nil)
}
