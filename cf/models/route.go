package models

import (
	"fmt"
	"net/url"
	"strings"
)

type Route struct {
	Guid   string
	Host   string
	Port   int
	Domain DomainFields
	Path   string

	Space           SpaceFields
	Apps            []ApplicationFields
	ServiceInstance ServiceInstanceFields
}

func (r Route) URL() string {
	return urlStringFromParts(r.Host, r.Port, r.Domain.Name, r.Path)
}

func urlStringFromParts(hostName string, port int, domainName string, path string) string {
	var host string
	if hostName != "" {
		host = fmt.Sprintf("%s.%s", hostName, domainName)
	} else {
		host = domainName
	}

	if port > 0 {
		host = fmt.Sprintf("%s:%d", host, port)
	}

	u := url.URL{
		Host: host,
		Path: path,
	}

	return strings.TrimPrefix(u.String(), "//") // remove the empty scheme
}
