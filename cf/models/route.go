package models

import (
	"fmt"
	"net/url"
	"strings"
)

type Route struct {
	Guid   string
	Host   string
	Domain DomainFields
	Path   string

	Space SpaceFields
	Apps  []ApplicationFields
}

func (r Route) URL() string {
	return urlStringFromParts(r.Host, r.Domain.Name, r.Path)
}

func urlStringFromParts(hostName, domainName, path string) string {
	var host string
	if hostName != "" {
		host = fmt.Sprintf("%s.%s", hostName, domainName)
	} else {
		host = domainName
	}

	u := url.URL{
		Host: host,
		Path: path,
	}

	return strings.TrimPrefix(u.String(), "//") // remove the empty scheme
}
