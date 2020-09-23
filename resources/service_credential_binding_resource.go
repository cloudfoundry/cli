package resources

import (
	"code.cloudfoundry.org/jsonry"
)

type ServiceCredentialBindingType string

const (
	AppBinding ServiceCredentialBindingType = "app"
	KeyBinding ServiceCredentialBindingType = "key"
)

type ServiceCredentialBinding struct {
	// Type is either "app" or "key"
	Type ServiceCredentialBindingType `jsonry:"type,omitempty"`
	// GUID is a unique service credential binding identifier.
	GUID string `jsonry:"guid,omitempty"`
	// Name is the name of the service credential binding.
	Name string `jsonry:"name,omitempty"`
	// ServiceInstanceGUID is the service instance that this binding originates from
	ServiceInstanceGUID string `jsonry:"relationships.service_instance.data.guid,omitempty"`
	// AppGUID is the application that this binding is attached to
	AppGUID string `jsonry:"relationships.app.data.guid,omitempty"`
	// AppName is the application name. It is not part of the API response, and is here as pragmatic convenience.
	AppName string `jsonry:"-"`
	// LastOperation is the last operation on the service credential binding
	LastOperation LastOperation `jsonry:"last_operation"`
}

func (s ServiceCredentialBinding) MarshalJSON() ([]byte, error) {
	return jsonry.Marshal(s)
}

func (s *ServiceCredentialBinding) UnmarshalJSON(data []byte) error {
	return jsonry.Unmarshal(data, s)
}
