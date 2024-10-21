package resources

import (
	"code.cloudfoundry.org/cli/v9/types"
	"code.cloudfoundry.org/jsonry"
)

type ServiceInstanceType string

const (
	UserProvidedServiceInstance ServiceInstanceType = "user-provided"
	ManagedServiceInstance      ServiceInstanceType = "managed"
)

type ServiceInstance struct {
	// Type is either "user-provided" or "managed"
	Type ServiceInstanceType `jsonry:"type,omitempty"`
	// GUID is a unique service instance identifier.
	GUID string `jsonry:"guid,omitempty"`
	// Name is the name of the service instance.
	Name string `jsonry:"name,omitempty"`
	// SpaceGUID is the space that this service instance relates to
	SpaceGUID string `jsonry:"relationships.space.data.guid,omitempty"`
	// ServicePlanGUID is the service plan that this service instance relates to
	ServicePlanGUID string `jsonry:"relationships.service_plan.data.guid,omitempty"`
	// Tags are used by apps to identify service instances.
	Tags types.OptionalStringSlice `jsonry:"tags"`
	// SyslogDrainURL is where logs are streamed
	SyslogDrainURL types.OptionalString `jsonry:"syslog_drain_url"`
	// RouteServiceURL is where requests for bound routes will be forwarded
	RouteServiceURL types.OptionalString `jsonry:"route_service_url"`
	// DashboardURL is where the service can be monitored
	DashboardURL types.OptionalString `jsonry:"dashboard_url"`
	// Credentials are passed to the app
	Credentials types.OptionalObject `jsonry:"credentials"`
	// UpgradeAvailable says whether the plan is at a higher version
	UpgradeAvailable types.OptionalBoolean `json:"upgrade_available"`
	// MaintenanceInfoVersion is the version this service is at
	MaintenanceInfoVersion string `jsonry:"maintenance_info.version,omitempty"`
	// Parameters are passed to the service broker
	Parameters types.OptionalObject `jsonry:"parameters"`
	// LastOperation is the last operation on the service instance
	LastOperation LastOperation `jsonry:"last_operation"`

	Metadata *Metadata `json:"metadata,omitempty"`
}

func (s ServiceInstance) MarshalJSON() ([]byte, error) {
	return jsonry.Marshal(s)
}

func (s *ServiceInstance) UnmarshalJSON(data []byte) error {
	return jsonry.Unmarshal(data, s)
}
