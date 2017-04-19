/**
* This plugin demonstrate the use of Test driven development using the test rpc server
* This allows the plugin to be tested independently without relying on CF CLI
 */
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"code.cloudfoundry.org/cli/cf/flags"
	"code.cloudfoundry.org/cli/plugin"
)

type DemoCmd struct{}

type AppsModel struct {
	NextURL   string     `json:"next_url,omitempty"`
	Resources []AppModel `json:"resources"`
}

type EntityModel struct {
	Name  string `json:"name"`
	State string `json:"state"`
}

type AppModel struct {
	Entity EntityModel `json:"entity"`
}

func (c *DemoCmd) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "App-Lister",
		Commands: []plugin.Command{
			{
				Name:     "list-apps",
				HelpText: "curl /v2/apps to get a list of apps",
				UsageDetails: plugin.Usage{
					Usage: "cf list-apps [--started | --stopped]",
					Options: map[string]string{
						"--started": "Shows only apps that are started",
						"--stopped": "Shows only apps that are stopped",
					},
				},
			},
		},
	}
}

func main() {
	plugin.Start(new(DemoCmd))
}

func (c *DemoCmd) Run(cliConnection plugin.CliConnection, args []string) {
	switch args[0] {
	case "list-apps":
		fc, err := parseArguments(args)
		if err != nil {
			exit1(err.Error())
		}

		endpoint, err := cliConnection.ApiEndpoint()
		if err != nil {
			exit1("Error getting targeted endpoint: " + err.Error())
		}
		fmt.Printf("Listing apps @ endpoint %s/v2/apps...\n\n", endpoint)

		allApps, err := getAllApps(cliConnection)
		if err != nil {
			exit1("Error curling v2/apps: " + err.Error())
		}

		for _, app := range allApps.Resources {
			if (fc.IsSet("started") && app.Entity.State == "STARTED") ||
				(fc.IsSet("stopped") && app.Entity.State == "STOPPED") ||
				(!fc.IsSet("stopped") && !fc.IsSet("started")) {
				fmt.Println(app.Entity.Name)
			}
		}

	case "CLI-MESSAGE-UNINSTALL":
		fmt.Println("Thanks for using this demo")
	}
}

func getAllApps(cliConnection plugin.CliConnection) (AppsModel, error) {
	nextURL := "v2/apps"
	allApps := AppsModel{}

	for nextURL != "" {
		output, err := cliConnection.CliCommandWithoutTerminalOutput("curl", nextURL)
		if err != nil {
			return AppsModel{}, err
		}

		apps := AppsModel{}
		err = json.Unmarshal([]byte(output[0]), &apps)
		if err != nil {
			return AppsModel{}, err
		}

		allApps.Resources = append(allApps.Resources, apps.Resources...)

		nextURL = apps.NextURL
	}

	return allApps, nil
}

func parseArguments(args []string) (flags.FlagContext, error) {
	fc := flags.New()
	fc.NewBoolFlag("started", "s", "Shows only apps that are started")
	fc.NewBoolFlag("stopped", "o", "Shows only apps that are stopped")
	err := fc.Parse(args...)

	return fc, err
}

func exit1(err string) {
	fmt.Println("FAILED\n" + err)
	os.Exit(1)
}
