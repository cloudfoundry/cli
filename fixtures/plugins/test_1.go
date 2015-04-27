/**
	* 1. Setup the server so cf can call it under main.
				e.g. `cf my-plugin` creates the callable server. now we can call the Run command
	* 2. Implement Run that is the actual code of the plugin!
	* 3. Return an error
**/

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/cli/plugin"
)

type Test1 struct {
}

func (c *Test1) Run(cliConnection plugin.CliConnection, args []string) {
	if args[0] == "new-api" {
		token, _ := cliConnection.AccessToken()
		fmt.Println("Access Token:", token)
		fmt.Println("")

		hasOrg, _ := cliConnection.HasOrganization()
		fmt.Println("Has Organization Targeted:", hasOrg)
		org, _ := cliConnection.GetCurrentOrg()
		fmt.Println("Current Org:", org)
		hasSpace, _ := cliConnection.HasSpace()
		fmt.Println("Has Space Targeted:", hasSpace)
		space, _ := cliConnection.GetCurrentSpace()
		fmt.Println("Current space:", space)

		loggregator, _ := cliConnection.LoggregatorEndpoint()
		fmt.Println("Loggregator Endpoint:", loggregator)
		dopplerEndpoint, _ := cliConnection.DopplerEndpoint()
		fmt.Println("Doppler Endpoint:", dopplerEndpoint)

		user, _ := cliConnection.Username()
		fmt.Println("Current user:", user)
		userGuid, _ := cliConnection.UserGuid()
		fmt.Println("Current user guid:", userGuid)
		email, _ := cliConnection.UserEmail()
		fmt.Println("Current user email:", email)

		hasAPI, _ := cliConnection.HasAPIEndpoint()
		fmt.Println("Has API Endpoint:", hasAPI)
		api, _ := cliConnection.ApiEndpoint()
		fmt.Println("Current api:", api)
		version, _ := cliConnection.ApiVersion()
		fmt.Println("Current api version:", version)

		loggedIn, _ := cliConnection.IsLoggedIn()
		fmt.Println("Is Logged In:", loggedIn)
		isSSLDisabled, _ := cliConnection.IsSSLDisabled()
		fmt.Println("Is SSL Disabled:", isSSLDisabled)
	} else if args[0] == "test_1_cmd1" {
		theFirstCmd()
	} else if args[0] == "test_1_cmd2" {
		theSecondCmd()
	} else if args[0] == "CLI-MESSAGE-UNINSTALL" {
		uninstalling()
	}
}

func (c *Test1) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "Test1",
		Version: plugin.VersionType{
			Major: 1,
			Minor: 2,
			Build: 4,
		},
		Commands: []plugin.Command{
			{
				Name:     "test_1_cmd1",
				Alias:    "test_1_cmd1_alias",
				HelpText: "help text for test_1_cmd1",
				UsageDetails: plugin.Usage{
					Usage: "Test plugin command\n   cf test_1_cmd1 [-a] [-b] [--no-ouput]",
					Options: map[string]string{
						"a":         "flag to do nothing",
						"b":         "another flag to do nothing",
						"no-output": "example option with no use",
					},
				},
			},
			{
				Name:     "test_1_cmd2",
				HelpText: "help text for test_1_cmd2",
			},
			{
				Name:     "new-api",
				HelpText: "test new api for plugins",
			},
		},
	}
}

func theFirstCmd() {
	fmt.Println("You called cmd1 in test_1")
}

func theSecondCmd() {
	fmt.Println("You called cmd2 in test_1")
}

func uninstalling() {
	os.Remove(filepath.Join(os.TempDir(), "uninstall-test-file-for-test_1.exe"))
}

func main() {
	plugin.Start(new(Test1))
}
