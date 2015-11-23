package models

import "fmt"

type Route struct {
	Guid   string
	Host   string
	Port   int
	Domain DomainFields

	Space SpaceFields
	Apps  []ApplicationFields
}

func (route Route) URL() string {
	if route.Host == "" && route.Port == 0 {
		return route.Domain.Name
	}
	if route.Port == 0 {
		return fmt.Sprintf("%s.%s", route.Host, route.Domain.Name)
	}
	if route.Host == "" {
		return fmt.Sprintf("%s:%d", route.Domain.Name, route.Port)
	}
	return fmt.Sprintf("%s.%s:%d", route.Host, route.Domain.Name, route.Port)
}

type RouteSummary struct {
	Guid   string
	Host   string
	Domain DomainFields
}

func (model RouteSummary) URL() string {
	if model.Host == "" {
		return model.Domain.Name
	}
	return fmt.Sprintf("%s.%s", model.Host, model.Domain.Name)
}
