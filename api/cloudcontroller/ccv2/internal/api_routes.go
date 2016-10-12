package internal

import "github.com/tedsuo/rata"

const (
	AppsFromRouteRequest          = "AppsFromRoute"
	AppsRequest                   = "Apps"
	DeleteRouteRequest            = "DeleteRoute"
	DeleteServiceBindingRequest   = "DeleteServiceBinding"
	InfoRequest                   = "Info"
	PrivateDomainRequest          = "PrivateDomain"
	RouteMappingsFromRouteRequest = "RouteMappingsFromRoute"
	RoutesFromSpaceRequest        = "RoutesFromSpace"
	ServiceBindingsRequest        = "ServiceBindings"
	ServiceInstancesRequest       = "ServiceInstances"
	SharedDomainRequest           = "SharedDomain"
	SpaceServiceInstancesRequest  = "SpaceServiceInstances"
)

var APIRoutes = rata.Routes{
	{Path: "/v2/apps", Method: "GET", Name: AppsRequest},
	{Path: "/v2/info", Method: "GET", Name: InfoRequest},
	{Path: "/v2/private_domains/:private_domain_guid", Method: "GET", Name: PrivateDomainRequest},
	{Path: "/v2/routes/:route_guid", Method: "DELETE", Name: DeleteRouteRequest},
	{Path: "/v2/routes/:route_guid/apps", Method: "GET", Name: AppsFromRouteRequest},
	{Path: "/v2/routes/:route_guid/route_mappings", Method: "GET", Name: RouteMappingsFromRouteRequest},
	{Path: "/v2/service_bindings", Method: "GET", Name: ServiceBindingsRequest},
	{Path: "/v2/service_bindings/:service_binding_guid", Method: "DELETE", Name: DeleteServiceBindingRequest},
	{Path: "/v2/service_instances", Method: "GET", Name: ServiceInstancesRequest},
	{Path: "/v2/shared_domains/:shared_domain_guid", Method: "GET", Name: SharedDomainRequest},
	{Path: "/v2/spaces/:guid/service_instances", Method: "GET", Name: SpaceServiceInstancesRequest},
	{Path: "/v2/spaces/:space_guid/routes", Method: "GET", Name: RoutesFromSpaceRequest},
}
