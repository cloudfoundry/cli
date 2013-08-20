package commands

import (
	"cf/configuration"
	term "cf/terminal"
	"crypto/tls"
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
		ui.Say("FAILED")
		return
	}

	ui.Say("target: %s", config.Target)
	email := ui.Ask("Email>")
	password := ui.Ask("Password>")
	ui.Say("Authenticating...")

	_, err = authenticate(config.AuthorizationEndpoint, email, password)

	if err != nil {
		ui.Say("FAILED")
		return
	}

	ui.Say("OK")
}

func authenticate(endpoint string, email string, password string) (response AuthenticationResponse, err error) {
	data := url.Values{"username": {email}, "password": {password}}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	req, err := http.NewRequest("POST", endpoint+"/oauth/token", strings.NewReader(data.Encode()))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)

	if err != nil {
		return
	}

	if resp.StatusCode > 299 {
		err = errors.New("Login error")
	}

	return
}
