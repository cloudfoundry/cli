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
	switch args[0] {
	case "CliCommandWithoutTerminalOutput":
		result, _ := cliConnection.CliCommandWithoutTerminalOutput("target")
		fmt.Println("Done CliCommandWithoutTerminalOutput:", result)
	case "CliCommand":
		result, _ := cliConnection.CliCommand("target")
		fmt.Println("Done CliCommand:", result)
	case "GetCurrentOrg":
		result, _ := cliConnection.GetCurrentOrg()
		fmt.Println("Done GetCurrentOrg:", result)
	case "GetCurrentSpace":
		result, _ := cliConnection.GetCurrentSpace()
		fmt.Println("Done GetCurrentSpace:", result)
	case "Username":
		result, _ := cliConnection.Username()
		fmt.Println("Done Username:", result)
	case "UserGuid":
		result, _ := cliConnection.UserGuid()
		fmt.Println("Done UserGuid:", result)
	case "UserEmail":
		result, _ := cliConnection.UserEmail()
		fmt.Println("Done UserEmail:", result)
	case "IsLoggedIn":
		result, _ := cliConnection.IsLoggedIn()
		fmt.Println("Done IsLoggedIn:", result)
	case "IsSSLDisabled":
		result, err := cliConnection.IsSSLDisabled()
		if err != nil {
			fmt.Println("Error in IsSSLDisabled()", err)
		}
		fmt.Println("Done IsSSLDisabled:", result)
	case "ApiEndpoint":
		result, _ := cliConnection.ApiEndpoint()
		fmt.Println("Done ApiEndpoint:", result)
	case "ApiVersion":
		result, _ := cliConnection.ApiVersion()
		fmt.Println("Done ApiVersion:", result)
	case "HasAPIEndpoint":
		result, err := cliConnection.HasAPIEndpoint()
		if err != nil {
			fmt.Println("Error in HasAPIEndpoint()", err)
		}
		fmt.Println("Done HasAPIEndpoint:", result)
	case "HasOrganization":
		result, _ := cliConnection.HasOrganization()
		fmt.Println("Done HasOrganization:", result)
	case "HasSpace":
		result, _ := cliConnection.HasSpace()
		fmt.Println("Done HasSpace:", result)
	case "LoggregatorEndpoint":
		result, _ := cliConnection.LoggregatorEndpoint()
		fmt.Println("Done LoggregatorEndpoint:", result)
	case "DopplerEndpoint":
		result, _ := cliConnection.DopplerEndpoint()
		fmt.Println("Done DopplerEndpoint:", result)
	case "AccessToken":
		result, _ := cliConnection.AccessToken()
		fmt.Println("Done AccessToken:", result)
	case "GetApp":
		result, _ := cliConnection.GetApp(args[1])
		fmt.Println("Done GetApp:", result)
	case "GetApps":
		result, _ := cliConnection.GetApps()
		fmt.Println("Done GetApps:", result)
	case "GetOrg":
		result, _ := cliConnection.GetOrg(args[1])
		fmt.Println("Done GetOrg:", result)
	case "GetOrgs":
		result, _ := cliConnection.GetOrgs()
		fmt.Println("Done GetOrgs:", result)
	case "GetSpace":
		result, _ := cliConnection.GetSpace(args[1])
		fmt.Println("Done GetSpace:", result)
	case "GetSpaces":
		result, _ := cliConnection.GetSpaces()
		fmt.Println("Done GetSpaces:", result)
	case "GetOrgUsers":
		result, _ := cliConnection.GetOrgUsers(args[1], args[2:]...)
		fmt.Println("Done GetOrgUsers:", result)
	case "GetSpaceUsers":
		result, _ := cliConnection.GetSpaceUsers(args[1], args[2])
		fmt.Println("Done GetSpaceUsers:", result)
	case "GetServices":
		result, _ := cliConnection.GetServices()
		fmt.Println("Done GetServices:", result)
	case "GetService":
		result, _ := cliConnection.GetService(args[1])
		fmt.Println("Done GetService:", result)
	case "TestPluginCommandWithAlias", "FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF":
		fmt.Println("You called Test Plugin Command With Alias!")
	}
}

func (c *Test1) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "CF-CLI-Integration-Test-Plugin",
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
			{Name: "CliCommandWithoutTerminalOutput"},
			{Name: "CliCommand"},
			{Name: "GetCurrentSpace"},
			{Name: "GetCurrentOrg"},
			{Name: "Username"},
			{Name: "UserGuid"},
			{Name: "UserEmail"},
			{Name: "IsLoggedIn"},
			{Name: "IsSSLDisabled"},
			{Name: "ApiEndpoint"},
			{Name: "ApiVersion"},
			{Name: "HasAPIEndpoint"},
			{Name: "HasOrganization"},
			{Name: "HasSpace"},
			{Name: "LoggregatorEndpoint"},
			{Name: "DopplerEndpoint"},
			{Name: "AccessToken"},
			{Name: "GetApp"},
			{Name: "GetApps"},
			{Name: "GetOrg"},
			{Name: "GetOrgs"},
			{Name: "GetSpace"},
			{Name: "GetSpaces"},
			{Name: "GetOrgUsers"},
			{Name: "GetSpaceUsers"},
			{Name: "GetServices"},
			{Name: "GetService"},
			{
				Name:     "TestPluginCommandWithAlias",
				Alias:    "FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF",
				HelpText: "This is my plugin help test. Banana.",
				UsageDetails: plugin.Usage{
					Usage: "I R Usage",
					Options: map[string]string{
						"--dis-flag": "is a flag",
					},
				},
			},
		},
	}
}

func uninstalling() {
	os.Remove(filepath.Join(os.TempDir(), "uninstall-test-file-for-test_1.exe"))
}

func main() {
	plugin.Start(new(Test1))
}
