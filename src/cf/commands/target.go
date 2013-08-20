package commands

import (
	"cf/api"
	"cf/configuration"
	term "cf/terminal"
	"encoding/json"
	"github.com/codegangsta/cli"
	"io/ioutil"
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

	req, err := http.NewRequest("GET", url+"/v2/info", nil)

	if err != nil {
		termUI.Failed("URL invalid.", err)
		return
	}

	client := api.NewClient()
	response, err := client.Do(req)

	if err != nil || response.StatusCode > 299 {
		termUI.Failed("Target refused connection.", err)
		return
	}

	jsonBytes, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		termUI.Failed("Could not read response body.", err)
		return
	}

	serverResponse := new(InfoResponse)
	err = json.Unmarshal(jsonBytes, &serverResponse)

	if err != nil {
		termUI.Failed("Invalid JSON response from server.", err)
		return
	}

	newConfiguration, err := saveTarget(url, serverResponse)

	if err != nil {
		termUI.Failed("Error saving configuration", err)
		return
	}

	termUI.Say(term.Green("OK"))
	showConfiguration(newConfiguration)
}

func showConfiguration(config configuration.Configuration) {
	termUI.Say("CF instance: %s (API version: %s)",
		term.Yellow(config.Target),
		term.Yellow(config.ApiVersion))

	termUI.Say("Logged out. Use '%s' to login.",
		term.Yellow("cf login USERNAME"))
}

func saveTarget(target string, info *InfoResponse) (config configuration.Configuration, err error) {
	config.Target = target
	config.ApiVersion = info.ApiVersion
	config.AuthorizationEndpoint = info.AuthorizationEndpoint
	err = config.Save()
	return
}
