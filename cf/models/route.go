package models

import (
	"fmt"
	"net/url"
	"strings"
)

type Route struct {
	GUID    string
	Host    string
	Domain  DomainFields
	Path    string
	Port    int
	Options map[string]string
	//Options         RouteOptions
	Space           SpaceFields
	Apps            []ApplicationFields
	ServiceInstance ServiceInstanceFields
}

/*type RouteOptions struct {
	LoadBalancing string `json:"load_balancing"`
	SessionCookie string `json:"session_cookie"`
}*/

func (r Route) URL() string {
	return (&RoutePresenter{
		Host:    r.Host,
		Domain:  r.Domain.Name,
		Path:    r.Path,
		Port:    r.Port,
		Options: r.Options,
	}).URL()
}

type RoutePresenter struct {
	Host    string
	Domain  string
	Path    string
	Port    int
	Options map[string]string
}

func (r *RoutePresenter) URL() string {
	var host string
	if r.Host != "" {
		host = r.Host + "." + r.Domain
	} else {
		host = r.Domain
	}

	if r.Port != 0 {
		host = fmt.Sprintf("%s:%d", host, r.Port)
	}

	u := url.URL{
		Host: host,
		Path: r.Path,
	}

	return strings.TrimPrefix(u.String(), "//") // remove the empty scheme
}

type ManifestRoute struct {
	Route string
}
