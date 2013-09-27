package commands

import (
	"cf"
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	"cf/terminal"
	"fmt"
	"github.com/codegangsta/cli"
)

type InfoResponse struct {
	ApiVersion            string `json:"api_version"`
	AuthorizationEndpoint string `json:"authorization_endpoint"`
}

type Target struct {
	ui         terminal.UI
	config     *configuration.Configuration
	configRepo configuration.ConfigurationRepository
	orgRepo    api.OrganizationRepository
	spaceRepo  api.SpaceRepository
}

func NewTarget(ui terminal.UI, configRepo configuration.ConfigurationRepository, orgRepo api.OrganizationRepository, spaceRepo api.SpaceRepository) (cmd Target) {
	cmd.ui = ui
	cmd.configRepo = configRepo
	cmd.config, _ = configRepo.Get()
	cmd.orgRepo = orgRepo
	cmd.spaceRepo = spaceRepo

	return
}

func (cmd Target) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
	}
	return
}

func (cmd Target) Run(c *cli.Context) {
	argsCount := len(c.Args())
	orgName := c.String("o")
	spaceName := c.String("s")

	if argsCount == 0 && orgName == "" && spaceName == "" {
		cmd.ui.ShowConfiguration(cmd.config)

		if !cmd.config.HasOrganization() {
			cmd.ui.Say("No org targeted. Use '%s target -o' to target an org.", cf.Name)
		}
		if !cmd.config.HasSpace() {
			cmd.ui.Say("No space targeted. Use '%s target -s' to target a space.", cf.Name)
		}
		return
	}

	if orgName != "" {
		cmd.setOrganization(orgName)
		if cmd.config.IsLoggedIn() {
			cmd.ui.Say("No space targeted. Use '%s target -s' to target a space.", cf.Name)
		}
		return
	}

	if spaceName != "" {
		cmd.setSpace(spaceName)
		return
	}

	return
}

func (cmd Target) setOrganization(orgName string) {
	if !cmd.config.IsLoggedIn() {
		cmd.ui.Failed("You must be logged in to set an organization. Use '%s login'.", cf.Name)
		return
	}

	org, found, err := cmd.orgRepo.FindByName(orgName)
	if err != nil {
		cmd.ui.Failed("Could not set organization.")
		return
	}

	if !found {
		cmd.ui.Failed(fmt.Sprintf("Organization %s not found.", orgName))
		return
	}

	cmd.config.Organization = org
	cmd.config.Space = cf.Space{}
	cmd.saveAndShowConfig()
}

func (cmd Target) setSpace(spaceName string) {
	if !cmd.config.IsLoggedIn() {
		cmd.ui.Failed("You must be logged in to set a space. Use '%s login'.", cf.Name)
		return
	}

	if !cmd.config.HasOrganization() {
		cmd.ui.Failed("Organization must be set before targeting space.")
		return
	}

	space, found, err := cmd.spaceRepo.FindByName(spaceName)
	if err != nil {
		cmd.ui.Failed("You do not have access to that space.")
		return
	}

	if !found {
		cmd.ui.Failed(fmt.Sprintf("Space %s not found.", spaceName))
		return
	}

	cmd.config.Space = space
	cmd.saveAndShowConfig()
}

func (cmd Target) saveAndShowConfig() {
	err := cmd.configRepo.Save()
	if err != nil {
		cmd.ui.Failed(err.Error())
		return
	}
	cmd.ui.ShowConfiguration(cmd.config)
}
