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

var termUI term.UI

func Target(c *cli.Context, ui term.UI) {
	termUI = ui

	if len(c.Args()) == 0 {
		showCurrentTarget()
	} else {
		setNewTarget(c.Args()[0])
	}

	return
}

func showCurrentTarget() {
	config, err := configuration.Load()

	if err != nil {
		config = configuration.Default()
	}

	showConfiguration(config)
}

func setNewTarget(target string) {
	url := "https://" + target
	termUI.Say("Setting target to %s...", term.Yellow(url))

	request, err := http.NewRequest("GET", url+"/v2/info", nil)

	if err != nil {
		termUI.Failed("URL invalid.", err)
		return
	}

	serverResponse := new(InfoResponse)
	err = api.PerformRequest(request, &serverResponse)

	if err != nil {
		termUI.Failed("", err)
	}

	newConfiguration, err := saveTarget(url, serverResponse)

	if err != nil {
		termUI.Failed("Error saving configuration", err)
		return
	}

	termUI.Ok()
	showConfiguration(newConfiguration)
}

func showConfiguration(config configuration.Configuration) {
	termUI.Say("CF instance: %s (API version: %s)",
		term.Yellow(config.Target),
		term.Yellow(config.ApiVersion))

	email := config.UserEmail()

	if email != "" {
		termUI.Say("user: %s", term.Yellow(email))
		termUI.Say("No org targeted. Use 'cf target -o' to target an org.")
	} else {
		termUI.Say("Logged out. Use '%s' to login.", term.Yellow("cf login USERNAME"))
	}
}

func saveTarget(target string, info *InfoResponse) (config configuration.Configuration, err error) {
	config.Target = target
	config.ApiVersion = info.ApiVersion
	config.AuthorizationEndpoint = info.AuthorizationEndpoint
	err = config.Save()
	return
}
