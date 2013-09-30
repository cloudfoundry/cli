package commands

import (
	"cf"
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	"cf/terminal"
	"errors"
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
	orgName := c.String("o")
	spaceName := c.String("s")

	if orgName == "" && spaceName == "" {
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
		err := cmd.setOrganization(orgName)

		if spaceName == "" && cmd.config.IsLoggedIn() {
			cmd.showConfig()
			cmd.ui.Say("No space targeted. Use '%s target -s' to target a space.", cf.Name)
			return
		}

		if err != nil {
			return
		}
	}

	if spaceName != "" {
		err := cmd.setSpace(spaceName)

		if err != nil {
			return
		}
	}
	cmd.showConfig()
	return
}

func (cmd Target) setOrganization(orgName string) (err error) {
	if !cmd.config.IsLoggedIn() {
		cmd.ui.Failed("You must be logged in to set an organization. Use '%s login'.", cf.Name)
		return
	}

	org, found, apiErr := cmd.orgRepo.FindByName(orgName)
	if apiErr != nil {
		err = apiErr
		cmd.ui.Failed("Could not set organization.")
		return
	}

	if !found {
		cmd.ui.Failed(fmt.Sprintf("Organization %s not found.", orgName))
		return errors.New("Org not found")
	}

	cmd.config.Organization = org
	cmd.config.Space = cf.Space{}
	cmd.saveConfig()
	return
}

func (cmd Target) setSpace(spaceName string) (err error) {
	if !cmd.config.IsLoggedIn() {
		cmd.ui.Failed("You must be logged in to set a space. Use '%s login'.", cf.Name)
		return
	}

	if !cmd.config.HasOrganization() {
		cmd.ui.Failed("Organization must be set before targeting space.")
		return
	}

	space, found, apiErr := cmd.spaceRepo.FindByName(spaceName)
	if apiErr != nil {
		err = apiErr
		cmd.ui.Failed("You do not have access to that space.")
		return
	}

	if !found {
		cmd.ui.Failed(fmt.Sprintf("Space %s not found.", spaceName))
		return errors.New("Space not found")
	}

	cmd.config.Space = space
	cmd.saveConfig()
	return
}

func (cmd Target) saveConfig() {
	err := cmd.configRepo.Save()
	if err != nil {
		cmd.ui.Failed(err.Error())
		return
	}
}

func (cmd Target) showConfig() {
	cmd.ui.ShowConfiguration(cmd.config)
}
