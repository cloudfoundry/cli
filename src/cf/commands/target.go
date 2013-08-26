package commands

import (
	"cf/api"
	"cf/configuration"
	term "cf/terminal"
	"github.com/codegangsta/cli"
)

type InfoResponse struct {
	ApiVersion            string `json:"api_version"`
	AuthorizationEndpoint string `json:"authorization_endpoint"`
}

type Target struct {
	ui        term.UI
	orgRepo   api.OrganizationRepository
	spaceRepo api.SpaceRepository
}

func NewTarget(ui term.UI, orgRepo api.OrganizationRepository, spaceRepo api.SpaceRepository) (t Target) {
	t.ui = ui
	t.orgRepo = orgRepo
	t.spaceRepo = spaceRepo

	return
}

func (t Target) Run(c *cli.Context) {
	argsCount := len(c.Args())
	orgName := c.String("o")
	spaceName := c.String("s")
	config, err := configuration.Load()

	if err != nil {
		t.ui.Failed("Error loading configuration", err)
		return
	}

	if argsCount == 0 && orgName == "" && spaceName == "" {
		t.ui.ShowConfiguration(config)
		return
	}

	if argsCount > 0 {
		t.setNewTarget(c.Args()[0])
		return
	}

	if orgName != "" {
		t.setOrganization(config, orgName)
		return
	}

	if spaceName != "" {
		t.setSpace(config, spaceName)
		return
	}

	return
}

func (t Target) setNewTarget(target string) {
	url := "https://" + target
	t.ui.Say("Setting target to %s...", term.Yellow(url))

	request, err := api.NewAuthorizedRequest("GET", url+"/v2/info", "", nil)

	if err != nil {
		t.ui.Failed("URL invalid.", err)
		return
	}

	serverResponse := new(InfoResponse)
	err = api.PerformRequestAndParseResponse(request, &serverResponse)

	if err != nil {
		t.ui.Failed("", err)
		return
	}

	newConfiguration, err := t.saveTarget(url, serverResponse)

	if err != nil {
		t.ui.Failed("Error saving configuration", err)
		return
	}

	t.ui.Ok()
	t.ui.ShowConfiguration(newConfiguration)
}

func (t Target) saveTarget(target string, info *InfoResponse) (config *configuration.Configuration, err error) {
	config = new(configuration.Configuration)
	config.Target = target
	config.ApiVersion = info.ApiVersion
	config.AuthorizationEndpoint = info.AuthorizationEndpoint
	err = config.Save()
	return
}

func (t Target) setOrganization(config *configuration.Configuration, orgName string) {
	if !config.IsLoggedIn() {
		t.ui.Failed("You must be logged in to set an organization.", nil)
		return
	}

	org, err := t.orgRepo.FindByName(config, orgName)
	if err != nil {
		t.ui.Failed("Could not set organization.", nil)
		return
	}

	config.Organization = org
	t.saveAndShowConfig(config)
}

func (t Target) setSpace(config *configuration.Configuration, spaceName string) {
	if !config.IsLoggedIn() {
		t.ui.Failed("You must be logged in to set a space.", nil)
		return
	}

	if !config.HasOrganization() {
		t.ui.Failed("Organization must be set before targeting space.", nil)
		return
	}

	space, err := t.spaceRepo.FindByName(config, spaceName)
	if err != nil {
		t.ui.Failed("You do not have access to that space.", nil)
		return
	}

	config.Space = space
	t.saveAndShowConfig(config)
}

func (t Target) saveAndShowConfig(config *configuration.Configuration) {
	err := config.Save()
	if err != nil {
		t.ui.Failed("Error saving configuration", err)
		return
	}
	t.ui.ShowConfiguration(config)
}
