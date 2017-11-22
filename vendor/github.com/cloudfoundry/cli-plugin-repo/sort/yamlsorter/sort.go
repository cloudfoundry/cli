package yamlsorter

import (
	"fmt"
	"os"
	"sort"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

type Plugin struct {
	Authors     []Author `yaml:"authors"`
	Binaries    []Binary `yaml:"binaries"`
	Company     *string  `yaml:"company"`
	Created     *string  `yaml:"created"`
	Description *string  `yaml:"description"`
	Homepage    *string  `yaml:"homepage"`
	Name        *string  `yaml:"name"`
	Updated     *string  `yaml:"updated"`
	Version     *string  `yaml:"version"`
}

type Binary struct {
	Checksum *string `yaml:"checksum"`
	Platform *string `yaml:"platform"`
	URL      *string `yaml:"url"`
}

type PluginsYAML struct {
	Plugins []Plugin `yaml:"plugins"`
}

type Author struct {
	Contact  *string `yaml:"contact,omitempty"`
	Homepage *string `yaml:"homepage,omitempty"`
	Name     *string `yaml:"name"`
}

func (p PluginsYAML) Len() int {
	return len(p.Plugins)
}

func (p PluginsYAML) Less(i, j int) bool {
	return strings.ToLower(*p.Plugins[i].Name) < strings.ToLower(*p.Plugins[j].Name)
}

func (p PluginsYAML) Swap(i, j int) {
	p.Plugins[i], p.Plugins[j] = p.Plugins[j], p.Plugins[i]
}

type YAMLSorter struct {
}

func (YAMLSorter) Sort(unsortedYAML []byte) ([]byte, error) {
	var plugins PluginsYAML

	err := yaml.Unmarshal(unsortedYAML, &plugins)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	sort.Sort(plugins)

	return yaml.Marshal(&plugins)
}
