package commands

import (
	"cf/api"
	"cf/configuration"
	term "cf/terminal"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/codegangsta/cli"
	"io/ioutil"
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

	client := api.NewClient()

	req, err := http.NewRequest("POST", endpoint+"/oauth/token", strings.NewReader(data.Encode()))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte("cf:")))

	rawResponse, err := client.Do(req)

	if err != nil {
		return
	}

	if rawResponse.StatusCode > 299 {
		err = errors.New("Login error")
	}

	jsonBytes, err := ioutil.ReadAll(rawResponse.Body)
	rawResponse.Body.Close()
	if err != nil {
		return
	}

	err = json.Unmarshal(jsonBytes, &response)

	if err != nil {
		return
	}

	return
}
