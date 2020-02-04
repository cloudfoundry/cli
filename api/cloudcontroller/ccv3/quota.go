package ccv3

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/types"
)

type Quota struct {
	// GUID is the unique ID of the organization quota.
	GUID string `json:"guid,omitempty"`
	// Name is the name of the organization quota
	Name string `json:"name"`
	// Apps contain the various limits that are associated with applications
	Apps AppLimit `json:"apps"`
	// Services contain the various limits that are associated with services
	Services ServiceLimit `json:"services"`
	// Routes contain the various limits that are associated with routes
	Routes RouteLimit `json:"routes"`
}

type AppLimit struct {
	TotalMemory       *types.NullInt `json:"total_memory_in_mb,omitempty"`
	InstanceMemory    *types.NullInt `json:"per_process_memory_in_mb,omitempty"`
	TotalAppInstances *types.NullInt `json:"total_instances,omitempty"`
}

func (al *AppLimit) UnmarshalJSON(rawJSON []byte) error {
	type Alias AppLimit

	var aux Alias
	err := json.Unmarshal(rawJSON, &aux)
	if err != nil {
		return err
	}

	*al = AppLimit(aux)

	if al.TotalMemory == nil {
		al.TotalMemory = &types.NullInt{
			IsSet: false,
			Value: 0,
		}
	}

	if al.InstanceMemory == nil {
		al.InstanceMemory = &types.NullInt{
			IsSet: false,
			Value: 0,
		}
	}

	if al.TotalAppInstances == nil {
		al.TotalAppInstances = &types.NullInt{
			IsSet: false,
			Value: 0,
		}
	}

	return nil
}

type ServiceLimit struct {
	TotalServiceInstances *types.NullInt `json:"total_service_instances,omitempty"`
	PaidServicePlans      *bool          `json:"paid_services_allowed,omitempty"`
}

func (sl *ServiceLimit) UnmarshalJSON(rawJSON []byte) error {
	type Alias ServiceLimit

	var aux Alias
	err := json.Unmarshal(rawJSON, &aux)
	if err != nil {
		return err
	}

	*sl = ServiceLimit(aux)

	if sl.TotalServiceInstances == nil {
		sl.TotalServiceInstances = &types.NullInt{
			IsSet: false,
			Value: 0,
		}
	}

	return nil
}

type RouteLimit struct {
	TotalRoutes        *types.NullInt `json:"total_routes,omitempty"`
	TotalReservedPorts *types.NullInt `json:"total_reserved_ports,omitempty"`
}

func (sl *RouteLimit) UnmarshalJSON(rawJSON []byte) error {
	type Alias RouteLimit

	var aux Alias
	err := json.Unmarshal(rawJSON, &aux)
	if err != nil {
		return err
	}

	*sl = RouteLimit(aux)

	if sl.TotalRoutes == nil {
		sl.TotalRoutes = &types.NullInt{
			IsSet: false,
			Value: 0,
		}
	}

	if sl.TotalReservedPorts == nil {
		sl.TotalReservedPorts = &types.NullInt{
			IsSet: false,
			Value: 0,
		}
	}

	return nil
}
