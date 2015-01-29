package models

import (
	"io"
	"time"
)

type PluginModel interface {
	PopulateModel(interface{}) []Plugin
}

type Plugins struct {
	logger io.Writer
}

type Plugin struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Version     string    `json:"version"`
	Date        time.Time `json:"date"`
	Company     string    `json:"company"`
	Author      string    `json:"author"`
	Contact     string    `json:"contact"`
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

func NewPlugins(logger io.Writer) PluginModel {
	return &Plugins{
		logger: logger,
	}
}

func (p *Plugins) PopulateModel(input interface{}) []Plugin {
	plugins := []Plugin{}
	if contents, ok := input.(map[interface{}]interface{})["plugins"].([]interface{}); ok {
		for _, plugin := range contents {
			plugins = append(plugins, p.extractPlugin(plugin))
		}
	} else {
		p.logger.Write([]byte("unexpected yaml structure, 'plugins' field not found.\n"))
	}
	return plugins
}

func (p *Plugins) extractPlugin(rawData interface{}) Plugin {
	plugin := Plugin{}
	for k, v := range rawData.(map[interface{}]interface{}) {
		switch k.(string) {
		case "name":
			plugin.Name = v.(string)
		case "description":
			plugin.Description = v.(string)
		case "binaries":
			for _, binary := range v.([]interface{}) {
				plugin.Binaries = append(plugin.Binaries, p.extractBinaries(binary))
			}
		case "version":
			plugin.Version = optionalStringField(v)
		case "author":
			plugin.Author = optionalStringField(v)
		case "contact":
			plugin.Contact = optionalStringField(v)
		case "homepage":
			plugin.Homepage = optionalStringField(v)
		case "company":
			plugin.Company = optionalStringField(v)
		case "date":
			plugin.Date = v.(time.Time)
		default:
			p.logger.Write([]byte("unexpected field in plugins: " + k.(string) + "\n"))
		}
	}
	return plugin
}

func (p *Plugins) extractBinaries(input interface{}) Binary {
	binary := Binary{}
	for k, v := range input.(map[interface{}]interface{}) {
		switch k.(string) {
		case "platform":
			binary.Platform = v.(string)
		case "url":
			binary.Url = v.(string)
		case "checksum":
			binary.Checksum = v.(string)
		default:
			p.logger.Write([]byte("unexpected field in binaries: %s" + k.(string) + "\n"))
		}
	}
	return binary
}

func optionalStringField(v interface{}) string {
	if v != nil {
		return v.(string)
	}
	return ""
}
