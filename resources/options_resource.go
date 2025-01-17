package resources

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"encoding/json"
	"strings"
)

type RouteOption struct {
	LoadBalancing string `json:"loadbalancing,omitempty"`
}

type RouteOptions struct {
	RouteOption *RouteOption `json:"options,omitempty"`
}

func (opts RouteOptions) MarshalJSON() ([]byte, error) {
	type ccRouteOption struct {
		LoadBalancing string `json:"loadbalancing,omitempty"`
	}
	var ccRouteOptions struct {
		RouteOptions *ccRouteOption `json:"options,omitempty"`
	}

	opts.RouteOption.LoadBalancing = ccRouteOptions.RouteOptions.LoadBalancing
	return json.Marshal(opts)
}

// UnmarshalJSON helps unmarshal a Cloud Controller Package response.
func (opts *RouteOptions) UnmarshalJSON(data []byte) error {
	var ccRouteOptions struct {
		RouteOptions struct {
			Loadbalancing string `json:"loadbalancing"`
		} `json:"options"`
	}

	err := cloudcontroller.DecodeJSON(data, &ccRouteOptions)
	if err != nil {
		return err
	}

	opts.RouteOption.LoadBalancing = ccRouteOptions.RouteOptions.Loadbalancing
	return nil
}

func NewRouteOptions(options []string) *RouteOption {
	routeOption := RouteOption{}
	for _, option := range options {
		key, value, found := strings.Cut(option, "=")
		if found && key == "loadbalancing" {
			routeOption.LoadBalancing = value
		}
	}
	return &routeOption
}
