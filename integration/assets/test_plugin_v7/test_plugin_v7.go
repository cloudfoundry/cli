package main

import (
	"fmt"

	plugin "code.cloudfoundry.org/cli/plugin/v7"
)

type Test1 struct {
}

func (c *Test1) Run(cliConnection plugin.CliConnection, args []string) {
	switch args[0] {
	case "GetApp":
		result, _ := cliConnection.GetApp(args[1])
		fmt.Println("Done GetApp:", result)
	case "TestPluginCommandWithAliasV7", "FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFV7":
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
			{Name: "GetApp"},
			{
				Name:     "TestPluginCommandWithAliasV7",
				Alias:    "FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFV7",
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
