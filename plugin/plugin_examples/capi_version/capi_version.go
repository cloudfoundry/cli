package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"code.cloudfoundry.org/cli/plugin"
)

type CapiVersionPlugin struct {
	Stdout io.Writer
	Stderr io.Writer
}

type VersionInfo struct {
	Name           string `json:"name"`
	Build          string `json:"build"`
	SupportAddress string `json:"support"`
	Description    string `json:"description"`
}

func NewCapiVersionPlugin(stdout io.Writer, stderr io.Writer) *CapiVersionPlugin {
	c := new(CapiVersionPlugin)
	c.Stdout = stdout
	c.Stderr = stderr
	return c
}

func (p *CapiVersionPlugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "capi-version",
		Version: plugin.VersionType{Major: 1, Minor: 0, Build: 0},
		Commands: []plugin.Command{
			{Name: "capi-version"},
		},
	}
}

func (p *CapiVersionPlugin) Run(conn plugin.CliConnection, args []string) {
	output, err := conn.CliCommandWithoutTerminalOutput("curl", "/v2/info")
	if err != nil {
		fmt.Fprintln(p.Stderr, "Problem retrieving version information")
		return
	}
	b := []byte(strings.Join(output, "\n"))

	var v VersionInfo
	err = json.Unmarshal(b, &v)
	if err != nil {
		fmt.Fprintln(p.Stderr, "Unexpected response from server")
		return
	}
	fmt.Fprintf(p.Stdout, "name: %s\n", v.Name)
	fmt.Fprintf(p.Stdout, "build: %s\n", v.Build)
	fmt.Fprintf(p.Stdout, "support address: %s\n", v.SupportAddress)
	fmt.Fprintf(p.Stdout, "description: %s\n", v.Description)
}

func main() {
	plugin.Start(NewCapiVersionPlugin(os.Stdout, os.Stderr))
}
