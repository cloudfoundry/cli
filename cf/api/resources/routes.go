package resources

import "github.com/cloudfoundry/cli/cf/models"

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

func (resource RouteResource) ToFields() (fields models.Route) {
	fields.Guid = resource.Metadata.Guid
	fields.Host = resource.Entity.Host
	return
}
func (resource RouteResource) ToModel() (route models.Route) {
	route.Host = resource.Entity.Host
	route.Guid = resource.Metadata.Guid
	route.Domain = resource.Entity.Domain.ToFields()
	route.Space = resource.Entity.Space.ToFields()
	for _, appResource := range resource.Entity.Apps {
		route.Apps = append(route.Apps, appResource.ToFields())
	}
	return
}
