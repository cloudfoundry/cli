package plugin_config

import (
	"encoding/json"

	"github.com/cloudfoundry/cli/plugin"
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

func (pd *PluginData) JsonMarshalV3() (output []byte, err error) {
	return json.MarshalIndent(pd, "", "  ")
}

func (pd *PluginData) JsonUnmarshalV3(input []byte) (err error) {
	return json.Unmarshal(input, pd)
}
