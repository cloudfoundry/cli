package commands

import (
	"cf/configuration"
	term "cf/terminal"
	"crypto/tls"
	"encoding/json"
	"github.com/codegangsta/cli"
	"io/ioutil"
	"net/http"
)

type InfoResponse struct {
	ApiVersion string `json:"api_version"`
}

func Target(c *cli.Context) {
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
	term.Say("Setting target to %s...", term.Yellow(url))

	req, err := http.NewRequest("GET", url+"/v2/info", nil)

	if err != nil {
		failedSettingTarget("URL invalid.", err)
		return
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	response, err := client.Do(req)

	if err != nil || response.StatusCode > 299 {
		failedSettingTarget("Target refused connection.", err)
		return
	}

	jsonBytes, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		failedSettingTarget("Could not read response body.", err)
		return
	}

	serverResponse := new(InfoResponse)
	err = json.Unmarshal(jsonBytes, &serverResponse)

	if err != nil {
		failedSettingTarget("Invalid JSON response from server.", err)
		return
	}

	newConfiguration, err := saveTarget(url, serverResponse.ApiVersion)

	if err != nil {
		failedSettingTarget("Error saving configuration", err)
		return
	}

	term.Say(term.Green("OK"))
	showConfiguration(newConfiguration)
}

func showConfiguration(config configuration.Configuration) {
	term.Say("CF instance: %s (API version: %s)",
		term.Yellow(config.Target),
		term.Yellow(config.ApiVersion))

	term.Say("Logged out. Use '%s' to login.",
		term.Yellow("cf login USERNAME"))
}

func failedSettingTarget(message string, err error) {
	term.Say(term.Red("FAILED"))
	term.Say(message)

	if err != nil {
		term.Say(err.Error())
	}
}

func saveTarget(target string, apiVersion string) (config configuration.Configuration, err error) {
	config.Target = target
	config.ApiVersion = apiVersion
	err = config.Save()
	return
}
