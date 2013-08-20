package commands

import (
	"cf/api"
	"cf/configuration"
	term "cf/terminal"
	"encoding/base64"
	"errors"
	"github.com/codegangsta/cli"
	"net/http"
	"net/url"
	"strings"
)

type AuthenticationResponse struct {
	AccessToken string `json:"access_token"`
}

func Login(c *cli.Context, ui term.UI) {
	config, err := configuration.Load()
	if err != nil {
		ui.Failed("Error loading configuration", err)
		return
	}

	ui.Say("target: %s", config.Target)
	email := ui.Ask("Email>")
	password := ui.Ask("Password>")
	ui.Say("Authenticating...")

	_, err = authenticate(config.AuthorizationEndpoint, email, password)

	if err != nil {
		ui.Failed("Error Authenticating", err)
		return
	}

	ui.Say("OK")
}

func authenticate(endpoint string, email string, password string) (response AuthenticationResponse, err error) {
	data := url.Values{
		"username":   {email},
		"password":   {password},
		"grant_type": {"password"},
		"scope":      {""},
	}

	client := api.NewClient()

	req, err := http.NewRequest("POST", endpoint+"/oauth/token", strings.NewReader(data.Encode()))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte("cf:")))

	resp, err := client.Do(req)

	if err != nil {
		return
	}

	if resp.StatusCode > 299 {
		err = errors.New("Login error")
	}

	return
}
