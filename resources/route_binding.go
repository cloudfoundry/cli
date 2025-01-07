package resources

import (
	"code.cloudfoundry.org/cli/v8/types"
	"code.cloudfoundry.org/jsonry"
)

type RouteBinding struct {
	GUID                string               `jsonry:"guid,omitempty"`
	RouteServiceURL     string               `jsonry:"route_service_url,omitempty"`
	ServiceInstanceGUID string               `jsonry:"relationships.service_instance.data.guid,omitempty"`
	RouteGUID           string               `jsonry:"relationships.route.data.guid,omitempty"`
	LastOperation       LastOperation        `jsonry:"last_operation"`
	Parameters          types.OptionalObject `jsonry:"parameters"`
}

func (s RouteBinding) MarshalJSON() ([]byte, error) {
	return jsonry.Marshal(s)
}

func (s *RouteBinding) UnmarshalJSON(data []byte) error {
	return jsonry.Unmarshal(data, s)
}
