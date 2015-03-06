package models

import (
	"fmt"
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
		case "authors":
			if v == nil {
				plugin.Authors = []Author{}
			} else {
				for _, author := range v.([]interface{}) {
					plugin.Authors = append(plugin.Authors, p.extractAuthors(author))
				}
			}
		case "homepage":
			plugin.Homepage = optionalStringField(v)
		case "company":
			plugin.Company = optionalStringField(v)
		case "created":
			plugin.Created = v.(time.Time)
		case "updated":
			plugin.Updated = v.(time.Time)
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

func (p *Plugins) extractAuthors(input interface{}) Author {
	author := Author{}
	for k, v := range input.(map[interface{}]interface{}) {
		switch k.(string) {
		case "name":
			author.Name = v.(string)
		case "homepage":
			author.Homepage = optionalStringField(v)
		case "contact":
			author.Contact = optionalStringField(v)
		default:
			p.logger.Write([]byte("unexpected field in Authors: %s" + k.(string) + "\n"))
		}
	}
	return author
}

func optionalStringField(v interface{}) string {
	if v != nil {
		switch v := v.(type) {
		default:
			return fmt.Sprintf("%v", v)
		case float64:
			return fmt.Sprintf("%.1f", v)
		case int64:
			return fmt.Sprintf("%d", v)
		case bool:
			return fmt.Sprintf("%t", v)
		}
	}
	return ""
}
