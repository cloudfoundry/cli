package web

import "time"

var ValidPlatforms = []string{"osx", "linux32", "linux64", "win32", "win64"}

type Plugin struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Version     string    `json:"version"`
	Created     time.Time `json:"created"`
	Updated     time.Time `json:"updated"`
	Company     string    `json:"company"`
	Authors     []Author  `json:"authors"`
	Homepage    string    `json:"homepage"`
	Binaries    []Binary  `json:"binaries"`
}

type Binary struct {
	Platform string `json:"platform"`
	Url      string `json:"url"`
	Checksum string `json:"checksum"`
}

type PluginsJson struct {
	Plugins []Plugin `json:"plugins"`
}

type Author struct {
	Name     string `json:"name"`
	Homepage string `json:"homepage"`
	Contact  string `json:"contact"`
}

func (p PluginsJson) Len() int {
	return len(p.Plugins)
}

func (p PluginsJson) Less(i, j int) bool {
	return p.Plugins[i].Updated.After(p.Plugins[j].Updated)
}

func (p PluginsJson) Swap(i, j int) {
	p.Plugins[i], p.Plugins[j] = p.Plugins[j], p.Plugins[i]
}
