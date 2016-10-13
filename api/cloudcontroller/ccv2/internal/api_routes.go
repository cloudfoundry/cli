package internal

import "github.com/tedsuo/rata"

const (
	AppsRequest                  = "Apps"
	DeleteServiceBindingRequest  = "DeleteServiceBinding"
	InfoRequest                  = "Info"
	ServiceBindingsRequest       = "ServiceBindings"
	ServiceInstancesRequest      = "ServiceInstances"
	SpaceServiceInstancesRequest = "SpaceServiceInstances"
)

var APIRoutes = rata.Routes{
	{Path: "/v2/apps", Method: "GET", Name: AppsRequest},
	{Path: "/v2/info", Method: "GET", Name: InfoRequest},
	{Path: "/v2/service_bindings", Method: "GET", Name: ServiceBindingsRequest},
	{Path: "/v2/service_bindings/:service_binding_guid", Method: "DELETE", Name: DeleteServiceBindingRequest},
	{Path: "/v2/service_instances", Method: "GET", Name: ServiceInstancesRequest},
	{Path: "/v2/spaces/:guid/service_instances", Method: "GET", Name: SpaceServiceInstancesRequest},
}
