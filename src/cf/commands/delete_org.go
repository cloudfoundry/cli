package commands

import (
	"cf"
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
	"strings"
)

type DeleteOrg struct {
	ui         terminal.UI
	orgRepo    api.OrganizationRepository
	orgReq     requirements.OrganizationRequirement
	configRepo configuration.ConfigurationRepository
}

func NewDeleteOrg(ui terminal.UI, sR api.OrganizationRepository, cR configuration.ConfigurationRepository) (cmd *DeleteOrg) {
	cmd = new(DeleteOrg)
	cmd.ui = ui
	cmd.orgRepo = sR
	cmd.configRepo = cR
	return
}

func (cmd *DeleteOrg) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "delete-org")
		return

	}
	return
}

func (cmd *DeleteOrg) Run(c *cli.Context) {
	orgName := c.Args()[0]
	cmd.ui.Say("Deleting org %s...", terminal.EntityNameColor(orgName))

	org, found, apiErr := cmd.orgRepo.FindByName(orgName)

	if !found {
		cmd.ui.Ok()
		cmd.ui.Say("Orgaization %s was already deleted.", orgName)
		return
	}

	force := c.Bool("f")

	if !force {
		response := strings.ToLower(cmd.ui.Ask(
			"Really delete org %s and everything associated with it?%s",
			terminal.EntityNameColor(org.Name),
			terminal.PromptColor(">"),
		))
		if response != "y" && response != "yes" {
			return
		}
	}

	apiErr = cmd.orgRepo.Delete(org)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}
	config, err := cmd.configRepo.Get()
	if err != nil {
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
