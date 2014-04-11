package models

import "fmt"

type Route struct {
	Guid   string
	Host   string
	Domain DomainFields

	Space SpaceFields
	Apps  []ApplicationFields
}

func (route Route) URL() string {
	if route.Host == "" {
		return route.Domain.Name
	}
	return fmt.Sprintf("%s.%s", route.Host, route.Domain.Name)
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
