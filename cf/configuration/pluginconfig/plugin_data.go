package pluginconfig

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/plugin"
)

type PluginData struct {
	Plugins map[string]PluginMetadata
}

type PluginMetadata struct {
	Location string
	Version  plugin.VersionType
	Commands []plugin.Command
}

func NewData() *PluginData {
	return &PluginData{
		Plugins: make(map[string]PluginMetadata),
	}
}

func (pd *PluginData) JSONMarshalV3() (output []byte, err error) {
	return json.MarshalIndent(pd, "", "  ")
}

func (pd *PluginData) JSONUnmarshalV3(input []byte) (err error) {
	return json.Unmarshal(input, pd)
}
