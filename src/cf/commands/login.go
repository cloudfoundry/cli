package commands

import (
	"cf/api"
	"cf/configuration"
	term "cf/terminal"
	"encoding/base64"
	"fmt"
	"github.com/codegangsta/cli"
	"net/http"
	"net/url"
	"strings"
)

type AuthenticationResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

const maxLoginTries = 3

func Login(c *cli.Context, ui term.UI) {
	config, err := configuration.Load()
	if err != nil {
		ui.Failed("Error loading configuration", err)
		return
	}

	ui.Say("target: %s", term.Cyan(config.Target))
	email := ui.Ask("Email%s", term.Cyan(">"))

	for i := 0; i < maxLoginTries; i++ {
		password := ui.Ask("Password%s", term.Cyan(">"))
		ui.Say("Authenticating...")

		response, err := authenticate(config.AuthorizationEndpoint, email, password)

		if err != nil {
			ui.Failed("Error Authenticating", err)
			continue
		}

		config.AccessToken = fmt.Sprintf("%s %s", response.TokenType, response.AccessToken)
		err = config.Save()

		if err != nil {
			ui.Failed("Error Persisting Session", err)
			return
		}

		ui.Ok()
		return
	}
}

func authenticate(endpoint string, email string, password string) (response AuthenticationResponse, err error) {
	data := url.Values{
		"username":   {email},
		"password":   {password},
		"grant_type": {"password"},
		"scope":      {""},
	}

	request, err := http.NewRequest("POST", endpoint+"/oauth/token", strings.NewReader(data.Encode()))
	if err != nil {
		return
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte("cf:")))

	err = api.PerformRequest(request, &response)

	return
}
