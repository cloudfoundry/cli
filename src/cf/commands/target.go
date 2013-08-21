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
var authorizer api.Authorizer

func Target(c *cli.Context, ui term.UI, a api.Authorizer) {
	argsCount := len(c.Args())
	org := c.String("o")
	termUI = ui
	authorizer = a

	if argsCount == 0 && org == "" {
		showCurrentTarget()
		return
	}

	if argsCount > 0 {
		setNewTarget(c.Args()[0])
		return
	}

	if org != "" {
		setOrganization(org)
		return
	}

	return
}

func showCurrentTarget() {
	config, err := configuration.Load()

	if err != nil {
		termUI.Failed("Error parsing configuration", err)
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
		return
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
	termUI.Say("CF Target Info (where apps will be pushed)")
	termUI.Say("  CF API endpoint: %s (API version: %s)",
		term.Yellow(config.Target),
		term.Yellow(config.ApiVersion))

	email := config.UserEmail()

	if email != "" {
		termUI.Say("  user:            %s", term.Yellow(email))

		if config.Organization != "" {
			termUI.Say("  org:             %s", term.Yellow(config.Organization))
		} else {
			termUI.Say("  No org targeted. Use 'cf target -o' to target an org.")
		}
	} else {
		termUI.Say("  Logged out. Use '%s' to login.", term.Yellow("cf login USERNAME"))
	}
}

func saveTarget(target string, info *InfoResponse) (config configuration.Configuration, err error) {
	config.Target = target
	config.ApiVersion = info.ApiVersion
	config.AuthorizationEndpoint = info.AuthorizationEndpoint
	err = config.Save()
	return
}

func setOrganization(org string) {
	if !authorizer.CanAccessOrg("", org) {
		termUI.Failed("You do not have access to that org.", nil)
		return
	}

	config, err := configuration.Load()

	if err != nil {
		termUI.Failed("Error loading configuration", err)
		return
	}

	if !config.IsLoggedIn() {
		termUI.Failed("You must be logged in to set an organization.", nil)
		return
	}

	config.Organization = org

	err = config.Save()
	if err != nil {
		termUI.Failed("Error saving configuration", err)
		return
	}
	showConfiguration(config)
}
