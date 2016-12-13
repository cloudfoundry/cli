package main

import "code.cloudfoundry.org/cli/plugin"

type TestPluginWithPanic struct {
}

func (c *TestPluginWithPanic) Run(cliConnection plugin.CliConnection, args []string) {
	panic("oh muuuuuuuuuuuuuuuuuuuuuy!")
}

func (c *TestPluginWithPanic) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "CF-CLI-Panic-Integration-Test-Plugin",
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
			{Name: "freak-out"},
		},
	}
}

func uninstalling() {
}

func main() {
	plugin.Start(new(TestPluginWithPanic))
}
