package plugin

import (
	"net/url"
	"path"
	"strings"
)

// PluginRepository represents a plugin repository
type PluginRepository struct {
	Plugins []Plugin `json:"plugins"`
}

type PluginBinary struct {
	Platform string `json:"platform"`
	URL      string `json:"url"`
	Checksum string `json:"checksum"`
}

type Plugin struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Version     string         `json:"version"`
	Binaries    []PluginBinary `json:"binaries"`
}

func (client *Client) GetPluginRepository(repositoryURL string) (PluginRepository, error) {
	parsedURL, err := url.Parse(repositoryURL)
	if err != nil {
		return PluginRepository{}, err
	}

	parsedURL.Path = strings.TrimSuffix(parsedURL.Path, "/")
	if !strings.HasSuffix(parsedURL.Path, "/list") {
		parsedURL.Path = path.Join(parsedURL.Path, "list")
	}

	request, err := client.newGETRequest(parsedURL.String())
	if err != nil {
		return PluginRepository{}, err
	}

	var pluginRepository PluginRepository
	response := Response{
		Result: &pluginRepository,
	}
	err = client.connection.Make(request, &response, nil)
	if err != nil {
		return PluginRepository{}, err
	}

	return pluginRepository, nil
}
