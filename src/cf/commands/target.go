package commands

import (
	"github.com/codegangsta/cli"
	"cf/configuration"
	term "cf/terminal"
	"net/http"
	"crypto/tls"
)

func Target(c *cli.Context) {
	if len(c.Args()) == 0 {
		config := configuration.Default()

		term.Say("CF instance: %s (API version: %s)",
			term.Yellow(config.Target),
			term.Yellow(config.ApiVersion))

		term.Say("Logged out. Use '%s' to login.",
			term.Yellow("cf login USERNAME"))

	} else {
		target := c.Args()[0]
		term.Say("Setting target to %s...", term.Yellow(target))

		req, err := http.NewRequest("GET", target + "/v2/info", nil)

		if err != nil {
			term.Say(term.Red("FAILED"))
			term.Say("URL invalid.")
			return
		}

		tr := &http.Transport{
			TLSClientConfig:    &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{Transport: tr}
		response, err := client.Do(req)

		if err != nil || response.StatusCode > 299 {
			term.Say(term.Red("FAILED"))
			term.Say("Target refused connection.")
		} else {
			term.Say(term.Green("OK"))

			term.Say("API Endpoint: %s (API version: %s)",
				term.Yellow(target),
				"2")

			term.Say("Logged out. Use '%s' to login.",
				term.Green("cf login USERNAME"))
		}
	}

	return
}

