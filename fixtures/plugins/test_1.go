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

	"code.cloudfoundry.org/cli/plugin"
)

type Test1 struct {
}

func (c *Test1) Run(cliConnection plugin.CliConnection, args []string) {
	if args[0] == "new-api" {
		token, _ := cliConnection.AccessToken()
		fmt.Println("Access Token:", token)
		fmt.Println("")

		app, err := cliConnection.GetApp("test_app")
		fmt.Println("err for test_app", err)
		fmt.Println("test_app is: ", app)

		hasOrg, _ := cliConnection.HasOrganization()
		fmt.Println("Has Organization Targeted:", hasOrg)
		currentOrg, _ := cliConnection.GetCurrentOrg()
		fmt.Println("Current Org:", currentOrg)
		org, _ := cliConnection.GetOrg(currentOrg.Name)
		fmt.Println(currentOrg.Name, " Org:", org)
		orgs, _ := cliConnection.GetOrgs()
		fmt.Println("Orgs:", orgs)
		hasSpace, _ := cliConnection.HasSpace()
		fmt.Println("Has Space Targeted:", hasSpace)
		currentSpace, _ := cliConnection.GetCurrentSpace()
		fmt.Println("Current space:", currentSpace)
		space, _ := cliConnection.GetSpace(currentSpace.Name)
		fmt.Println("Space:", space)
		spaces, _ := cliConnection.GetSpaces()
		fmt.Println("Spaces:", spaces)
		loggregator, _ := cliConnection.LoggregatorEndpoint()
		fmt.Println("Loggregator Endpoint:", loggregator)
		dopplerEndpoint, _ := cliConnection.DopplerEndpoint()
		fmt.Println("Doppler Endpoint:", dopplerEndpoint)

		user, _ := cliConnection.Username()
		fmt.Println("Current user:", user)
		userGUID, _ := cliConnection.UserGuid()
		fmt.Println("Current user guid:", userGUID)
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
		MinCliVersion: plugin.VersionType{
			Major: 5,
			Minor: 0,
			Build: 0,
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
