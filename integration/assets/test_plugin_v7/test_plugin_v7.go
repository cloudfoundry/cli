// +build V7

package main

import (
	"fmt"

	plugin "code.cloudfoundry.org/cli/plugin/v7"
)

type Test1 struct {
}

func (c *Test1) Run(cliConnection plugin.CliConnection, args []string) {
	commandName := args[0]

	switch commandName {
	case "AccessToken":
		result, err := cliConnection.AccessToken()
		if err != nil {
			fmt.Printf("Error %s: %s\n", commandName, err)
		} else {
			fmt.Println("Done AccessToken:", result)
		}
	case "ApiEndpoint":
		result, err := cliConnection.ApiEndpoint()
		if err != nil {
			fmt.Printf("Error %s: %s\n", commandName, err)
		} else {
			fmt.Println("Done ApiEndpoint:", result)
		}
	case "GetApp":
		result, err := cliConnection.GetApp(args[1])
		if err != nil {
			fmt.Printf("Error %s: %s\n", commandName, err)
		} else {
			fmt.Println("Done GetApp:", result)
		}
	case "GetApps":
		apps, err := cliConnection.GetApps()
		if err != nil {
			fmt.Printf("Error %s: %s\n", commandName, err)
		} else if len(apps) == 0 {
			fmt.Println("No apps in the current space")
		} else {
			fmt.Println("Current Apps:")
			for _, app := range apps {
				fmt.Printf("result: %+v, name: %s, guid: %s, ", app, app.Name, app.GUID)
				fmt.Printf("metadata: %+v\n", *app.Metadata)
			}
		}
	case "GetCurrentOrg":
		result, err := cliConnection.GetCurrentOrg()
		if err != nil {
			fmt.Printf("Error %s: %s\n", commandName, err)
		} else {
			fmt.Printf("Done GetCurrentOrg:, result:%+v, name: %s, guid: %s\n", result, result.Name, result.GUID)
		}
	case "GetCurrentSpace":
		result, err := cliConnection.GetCurrentSpace()
		if err != nil {
			fmt.Printf("Error %s: %s\n", commandName, err)
		} else {
			fmt.Printf("Done GetCurrentSpace:, result:%+v, name: %s, guid: %s\n", result, result.Name, result.GUID)
		}
	case "GetOrg":
		result, err := cliConnection.GetOrg(args[1])
		if err != nil {
			fmt.Printf("Error %s: %s\n", commandName, err)
		} else {
			fmt.Printf("Done GetOrg: name: %s, guid: %s\n", result.Name, result.GUID)
		}
	case "IsLoggedIn":
		result, err := cliConnection.IsLoggedIn()
		if err != nil {
			fmt.Printf("Error %s: %s\n", commandName, err)
		} else {
			fmt.Printf("Done IsLoggedIn: %v", result)
		}
	case "IsSkipSSLValidation":
		result, err := cliConnection.IsSkipSSLValidation()
		if err != nil {
			fmt.Printf("Error %s: %s\n", commandName, err)
		} else {
			fmt.Println("Done IsSkipSSLValidation:", result)
		}
	case "Username":
		result, err := cliConnection.Username()
		if err != nil {
			fmt.Printf("Error %s: %s\n", commandName, err)
		} else {
			fmt.Println("Done Username:", result)
		}
	case "TestPluginCommandWithAliasV7", "Cool-V7":
		fmt.Println("You called Test Plugin Command V7 With Alias!")
	case "CoolTest":
		fmt.Println("I am a test plugin")
	}
}
func (c *Test1) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "CF-CLI-Integration-Test-Plugin",
		Version: plugin.VersionType{
			Major: 6,
			Minor: 0,
			Build: 0,
		},
		MinCliVersion: plugin.VersionType{
			Major: 6,
			Minor: 0,
			Build: 0,
		},
		Commands: []plugin.Command{
			{Name: "AccessToken"},
			{Name: "ApiEndpoint"},
			{Name: "GetApp"},
			{Name: "GetApps"},
			{Name: "GetCurrentOrg"},
			{Name: "GetCurrentSpace"},
			{Name: "GetOrg"},
			{Name: "IsLoggedIn"},
			{Name: "IsSkipSSLValidation"},
			{Name: "Username"},
			{
				Name:     "TestPluginCommandWithAliasV7",
				Alias:    "Cool-V7",
				HelpText: "This is my plugin help test. Banana.",
				UsageDetails: plugin.Usage{
					Usage: "I R Usage",
					Options: map[string]string{
						"--dis-flag": "is a flag",
					},
				},
			},
			{Name: "CoolTest"},
		},
	}
}

// func uninstalling() {
//  os.Remove(filepath.Join(os.TempDir(), "uninstall-test-file-for-test_1.exe"))
// }

func main() {
	plugin.Start(new(Test1))
}
