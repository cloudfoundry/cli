package ccv2

import "github.com/tedsuo/rata"

const (
	InfoRequest                 = "Info"
	AppsRequest                 = "Apps"
	ServiceInstancesRequest     = "ServiceInstances"
	ServiceBindingsRequest      = "ServiceBindings"
	DeleteServiceBindingRequest = "DeleteServiceBinding"
)

var routes = rata.Routes{
	{Path: "/v2/info", Method: "GET", Name: InfoRequest},
	{Path: "/v2/apps", Method: "GET", Name: AppsRequest},
	{Path: "/v2/service_instances", Method: "GET", Name: ServiceInstancesRequest},
	{Path: "/v2/service_bindings", Method: "GET", Name: ServiceBindingsRequest},
	{Path: "/v2/service_bindings/:service_binding_guid", Method: "DELETE", Name: DeleteServiceBindingRequest},
}
