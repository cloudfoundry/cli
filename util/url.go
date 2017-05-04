package util

import (
	"net/url"
	"strings"
)

const (
	PluginHTTPPath            = iota
	PluginFilePath            = iota
	PluginUnsupportedPathType = iota
)

func DeterminePathType(s string) int {
	_, err := url.ParseRequestURI(s)

	if err == nil && strings.Contains(s, "://") {
		if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
			return PluginHTTPPath
		} else {
			return PluginUnsupportedPathType
		}
	}

	return PluginFilePath
}
