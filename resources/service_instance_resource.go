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
	Type ServiceInstanceType `jsonry:"type"`
	// GUID is a unique service instance identifier.
	GUID string `jsonry:"guid,omitempty"`
	// Name is the name of the service instance.
	Name string `jsonry:"name"`
	// SpaceGUID is the space that this service instance relates to
	SpaceGUID string `jsonry:"relationships.space.data.guid"`
}

func (s ServiceInstance) MarshalJSON() ([]byte, error) {
	return jsonry.Marshal(s)
}
