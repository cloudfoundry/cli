package resources

import "cf/models"

type RouteResource struct {
	Resource
	Entity RouteEntity
}

type RouteEntity struct {
	Host   string
	Domain DomainResource
	Space  SpaceResource
	Apps   []ApplicationResource
}

func (resource RouteResource) ToFields() (fields models.RouteFields) {
	fields.Guid = resource.Metadata.Guid
	fields.Host = resource.Entity.Host
	return
}
func (resource RouteResource) ToModel() (route models.Route) {
	route.RouteFields = resource.ToFields()
	route.Domain = resource.Entity.Domain.ToFields()
	route.Space = resource.Entity.Space.ToFields()
	for _, appResource := range resource.Entity.Apps {
		route.Apps = append(route.Apps, appResource.ToFields())
	}
	return
}
