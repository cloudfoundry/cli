package ccv3

import "code.cloudfoundry.org/cli/types"

type AppLimit struct {
	TotalMemory       types.NullInt `json:"total_memory_in_mb"`
	InstanceMemory    types.NullInt `json:"per_process_memory_in_mb"`
	TotalAppInstances types.NullInt `json:"total_instances"`
}

type ServiceLimit struct {
	TotalServiceInstances types.NullInt `json:"total_service_instances"`
	PaidServicePlans      bool          `json:"paid_services_allowed"`
}

type RouteLimit struct {
	TotalRoutes        types.NullInt `json:"total_routes"`
	TotalReservedPorts types.NullInt `json:"total_reserved_ports"`
}
