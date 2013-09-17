package commands

import (
	"cf"
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	term "cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
	"strings"
)

type DeleteOrg struct {
	ui         term.UI
	orgRepo    api.OrganizationRepository
	orgReq     requirements.OrganizationRequirement
	configRepo configuration.ConfigurationRepository
}

func NewDeleteOrg(ui term.UI, sR api.OrganizationRepository, cR configuration.ConfigurationRepository) (cmd *DeleteOrg) {
	cmd = new(DeleteOrg)
	cmd.ui = ui
	cmd.orgRepo = sR
	cmd.configRepo = cR
	return
}

func (cmd *DeleteOrg) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	var orgName string

	if len(c.Args()) == 1 {
		orgName = c.Args()[0]
	}

	if orgName == "" {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "delete-org")
		return
	}

	cmd.orgReq = reqFactory.NewOrganizationRequirement(orgName)

	reqs = []requirements.Requirement{cmd.orgReq}
	return
}

func (cmd *DeleteOrg) Run(c *cli.Context) {
	org := cmd.orgReq.GetOrganization()
	force := c.Bool("f")

	if !force {
		response := strings.ToLower(cmd.ui.Ask(
			"Really delete org %s and everything associated with it?%s",
			term.EntityNameColor(org.Name),
			term.PromptColor(">"),
		))
		if response != "y" && response != "yes" {
			return
		}
	}

	cmd.ui.Say("Deleting org %s...", term.EntityNameColor(org.Name))
	err := cmd.orgRepo.Delete(org)
	if err != nil {
		cmd.ui.Failed(err.Error())
		return
	}
	config, apiErr := cmd.configRepo.Get()
	if apiErr != nil {
		cmd.ui.Failed("Couldn't reset your target. You should logout and log in again.")
		return
	}

	if org.Guid == config.Organization.Guid {
		config.Organization = cf.Organization{}
		config.Space = cf.Space{}
		cmd.configRepo.Save()
	}

	cmd.ui.Ok()
	return
}
