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
	configRepo configuration.ConfigurationRepository
	orgRepo    api.OrganizationRepository
	spaceRepo  api.SpaceRepository
}

func NewTarget(ui terminal.UI, configRepo configuration.ConfigurationRepository, orgRepo api.OrganizationRepository, spaceRepo api.SpaceRepository) (cmd Target) {
	cmd.ui = ui
	cmd.configRepo = configRepo
	cmd.orgRepo = orgRepo
	cmd.spaceRepo = spaceRepo

	return
}

func (cmd Target) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) > 0 {
		err = errors.New("incorrect usage")
		cmd.ui.FailWithUsage(c, "target")
		return
	}

	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
	}
	return
}

func (cmd Target) Run(c *cli.Context) {
	orgName := c.String("o")
	spaceName := c.String("s")

	config, err := cmd.configRepo.Get()
	if err != nil {
		cmd.ui.ConfigFailure(err)
		return
	}

	if orgName == "" && spaceName == "" {
		cmd.ui.ShowConfiguration(config)

		if !config.HasOrganization() {
			cmd.ui.Say("No org targeted. Use '%s target -o' to target an org.", cf.Name)
		}
		if !config.HasSpace() {
			cmd.ui.Say("No space targeted. Use '%s target -s' to target a space.", cf.Name)
		}
		return
	}

	if orgName != "" {
		config = cmd.setOrganization(orgName)

		if spaceName == "" && config.IsLoggedIn() {
			cmd.ui.ShowConfiguration(config)
			cmd.ui.Say("No space targeted. Use '%s target -s' to target a space.", cf.Name)
			return
		}
	}

	if spaceName != "" {
		config = cmd.setSpace(spaceName)
	}

	cmd.ui.ShowConfiguration(config)
	return
}

func (cmd Target) setOrganization(orgName string) (config configuration.Configuration) {
	config, err := cmd.configRepo.Get()
	if err != nil {
		cmd.ui.ConfigFailure(err)
		return
	}

	if !config.IsLoggedIn() {
		cmd.ui.Failed("You must be logged in to set an organization. Use '%s login'.", cf.Name)
		return
	}

	org, apiStatus := cmd.orgRepo.FindByName(orgName)
	if apiStatus.IsError() {
		cmd.ui.Failed("Could not set organization.")
		return
	}

	if apiStatus.IsNotFound() {
		cmd.ui.Failed(fmt.Sprintf("Organization %s not found.", orgName))
		return
	}

	config.Organization = org
	config.Space = cf.Space{}
	cmd.saveConfig(config)
	return
}

func (cmd Target) setSpace(spaceName string) (config configuration.Configuration) {
	config, err := cmd.configRepo.Get()
	if err != nil {
		cmd.ui.ConfigFailure(err)
		return
	}

	if !config.IsLoggedIn() {
		cmd.ui.Failed("You must be logged in to set a space. Use '%s login'.", cf.Name)
		return
	}

	if !config.HasOrganization() {
		cmd.ui.Failed("Organization must be set before targeting space.")
		return
	}

	space, apiStatus := cmd.spaceRepo.FindByName(spaceName)

	if apiStatus.IsError() {
		cmd.ui.Failed("You do not have access to that space.")
		return
	}

	if apiStatus.IsNotFound() {
		cmd.ui.Failed(fmt.Sprintf("Space %s not found.", spaceName))
		return
	}

	config.Space = space
	cmd.saveConfig(config)
	return
}

func (cmd Target) saveConfig(config configuration.Configuration) {
	err := cmd.configRepo.Save(config)
	if err != nil {
		cmd.ui.Failed(err.Error())
		return
	}
}
