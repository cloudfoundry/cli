package v2actions

import "fmt"

type Route struct {
	GUID     string
	Hostname string
	Domain   string
	Path     string
	Port     int
}

func (actor Actor) GetOrphanedRoutesBySpace(spaceGUID string) ([]Route, Warnings, error) {
	return nil, nil, nil
}

func (actor Actor) DeleteRouteByGUID(routeGUID string) (Warnings, error) {
	return nil, nil
}

func (r Route) String() string {
	var routeString string

	if r.Hostname != "" {
		routeString = fmt.Sprintf("%s.%s", r.Hostname, r.Domain)
	} else {
		routeString = r.Domain
	}

	if r.Port != 0 {
		routeString = fmt.Sprintf("%s:%d", routeString, r.Port)
		return routeString
	}

	if r.Path != "" {
		routeString = fmt.Sprintf("%s/%s", routeString, r.Path)
	}

	return routeString
}
