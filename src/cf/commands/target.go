package commands

import (
	"cf/api"
	"cf/configuration"
	term "cf/terminal"
	"github.com/codegangsta/cli"
	"net/http"
)

type InfoResponse struct {
	ApiVersion            string `json:"api_version"`
	AuthorizationEndpoint string `json:"authorization_endpoint"`
}

func Target(c *cli.Context, termUI term.UI, or api.OrganizationRepository, sr api.SpaceRepository) {
	ui = termUI
	orgRepo = or
	spaceRepo = sr

	argsCount := len(c.Args())
	orgName := c.String("o")
	spaceName := c.String("s")
	config, err := configuration.Load()

	if err != nil {
		ui.Failed("Error loading configuration", err)
		return
	}

	if argsCount == 0 && orgName == "" && spaceName == "" {
		showConfiguration(config)
		return
	}

	if argsCount > 0 {
		setNewTarget(c.Args()[0])
		return
	}

	if orgName != "" {
		setOrganization(config, orgName)
		return
	}

	if spaceName != "" {
		setSpace(config, spaceName)
		return
	}

	return
}

func showConfiguration(config *configuration.Configuration) {
	ui.Say("CF Target Info (where apps will be pushed)")
	ui.Say("  CF API endpoint: %s (API version: %s)",
		term.Yellow(config.Target),
		term.Yellow(config.ApiVersion))

	if !config.IsLoggedIn() {
		ui.Say("  Logged out. Use '%s' to login.", term.Yellow("cf login USERNAME"))
		return
	}

	ui.Say("  user:            %s", term.Yellow(config.UserEmail()))

	if config.HasOrganization() {
		ui.Say("  org:             %s", term.Yellow(config.Organization.Name))
	} else {
		ui.Say("  No org targeted. Use 'cf target -o' to target an org.")
	}

	if config.HasSpace() {
		ui.Say("  space:           %s", term.Yellow(config.Space.Name))
	} else {
		ui.Say("  No space targeted. Use 'cf target -s' to target a space.")
	}
}

func setNewTarget(target string) {
	url := "https://" + target
	ui.Say("Setting target to %s...", term.Yellow(url))

	request, err := http.NewRequest("GET", url+"/v2/info", nil)

	if err != nil {
		ui.Failed("URL invalid.", err)
		return
	}

	serverResponse := new(InfoResponse)
	err = api.PerformRequest(request, &serverResponse)

	if err != nil {
		ui.Failed("", err)
		return
	}

	newConfiguration, err := saveTarget(url, serverResponse)

	if err != nil {
		ui.Failed("Error saving configuration", err)
		return
	}

	ui.Ok()
	showConfiguration(newConfiguration)
}

func saveTarget(target string, info *InfoResponse) (config *configuration.Configuration, err error) {
	config = new(configuration.Configuration)
	config.Target = target
	config.ApiVersion = info.ApiVersion
	config.AuthorizationEndpoint = info.AuthorizationEndpoint
	err = config.Save()
	return
}

func setOrganization(config *configuration.Configuration, orgName string) {
	if !config.IsLoggedIn() {
		ui.Failed("You must be logged in to set an organization.", nil)
		return
	}

	org, err := orgRepo.FindOrganizationByName(config, orgName)
	if err != nil {
		ui.Failed("Could not set organization.", nil)
		return
	}

	config.Organization = org
	saveAndShowConfig(config)
}

func setSpace(config *configuration.Configuration, spaceName string) {
	if !config.IsLoggedIn() {
		ui.Failed("You must be logged in to set a space.", nil)
		return
	}

	if !config.HasOrganization() {
		ui.Failed("Organization must be set before targeting space.", nil)
		return
	}

	space, err := spaceRepo.FindSpaceByName(config, spaceName)
	if err != nil {
		ui.Failed("You do not have access to that space.", nil)
		return
	}

	config.Space = space
	saveAndShowConfig(config)
}

func saveAndShowConfig(config *configuration.Configuration) {
	err := config.Save()
	if err != nil {
		ui.Failed("Error saving configuration", err)
		return
	}
	showConfiguration(config)
}
