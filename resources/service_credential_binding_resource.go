package resources

import (
	"code.cloudfoundry.org/cli/v8/types"
	"code.cloudfoundry.org/jsonry"
)

type ServiceCredentialBindingType string

const (
	AppBinding ServiceCredentialBindingType = "app"
	KeyBinding ServiceCredentialBindingType = "key"
)

type BindingStrategyType string

const (
	SingleBindingStrategy   BindingStrategyType = "single"
	MultipleBindingStrategy BindingStrategyType = "multiple"
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
	// AppSpaceGUID is the space guid of the app. It is not part of the API response, and is here as pragmatic convenience.
	AppSpaceGUID string `jsonry:"-"`
	// LastOperation is the last operation on the service credential binding
	LastOperation LastOperation `jsonry:"last_operation"`
	// Parameters can be specified when creating a binding
	Parameters types.OptionalObject `jsonry:"parameters"`
	// Strategy can be "single" (default) or "multiple"
	Strategy BindingStrategyType `jsonry:"strategy,omitempty"`
}

func (s ServiceCredentialBinding) MarshalJSON() ([]byte, error) {
	return jsonry.Marshal(s)
}

func (s *ServiceCredentialBinding) UnmarshalJSON(data []byte) error {
	return jsonry.Unmarshal(data, s)
}
