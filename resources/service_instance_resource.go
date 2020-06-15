package resources

import (
	"code.cloudfoundry.org/jsonry"
)

type ServiceInstanceType string

const (
	UserProvidedServiceInstance ServiceInstanceType = "user-provided"
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
	// Tags are used by apps to identify service instances.
	Tags []string `jsonry:"tags,omitempty"`
	// SyslogDrainURL is where logs are streamed
	SyslogDrainURL string `jsonry:"syslog_drain_url,omitempty"`
	// RouteServiceURL is where requests for bound routes will be forwarded
	RouteServiceURL string `jsonry:"route_service_url,omitempty"`
	// Credentials are passed to the app
	Credentials map[string]interface{} `jsonry:"credentials,omitempty"`
}

func (s ServiceInstance) MarshalJSON() ([]byte, error) {
	return jsonry.Marshal(s)
}

func (s *ServiceInstance) UnmarshalJSON(data []byte) error {
	return jsonry.Unmarshal(data, s)
}
