package ccv2

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
)

// RouteMapping represents a Cloud Controller map between an application and route.
type RouteMapping struct {
	// GUID is the unique route mapping identifier.
	GUID string

	// AppGUID is the unique application identifier.
	AppGUID string

	// RouteGUID is the unique route identifier.
	RouteGUID string
}

// UnmarshalJSON helps unmarshal a Cloud Controller Route Mapping
func (routeMapping *RouteMapping) UnmarshalJSON(data []byte) error {
	var ccRouteMapping struct {
		Metadata internal.Metadata `json:"metadata"`
		Entity   struct {
			AppGUID   string `json:"app_guid"`
			RouteGUID string `json:"route_guid"`
		} `json:"entity"`
	}

	err := cloudcontroller.DecodeJSON(data, &ccRouteMapping)
	if err != nil {
		return err
	}

	routeMapping.GUID = ccRouteMapping.Metadata.GUID
	routeMapping.AppGUID = ccRouteMapping.Entity.AppGUID
	routeMapping.RouteGUID = ccRouteMapping.Entity.RouteGUID
	return nil
}

// GetRouteMapping returns a route mapping with the provided guid.
func (client *Client) GetRouteMapping(guid string) (RouteMapping, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetRouteMappingRequest,
		URIParams:   Params{"route_mapping_guid": guid},
	})
	if err != nil {
		return RouteMapping{}, nil, err
	}

	var routeMapping RouteMapping
	response := cloudcontroller.Response{
		Result: &routeMapping,
	}

	err = client.connection.Make(request, &response)
	return routeMapping, response.Warnings, err
}

// GetRouteMappings returns a list of RouteMappings based off of the provided queries.
func (client *Client) GetRouteMappings(filters ...Filter) ([]RouteMapping, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetRouteMappingsRequest,
		Query:       ConvertFilterParameters(filters),
	})
	if err != nil {
		return nil, nil, err
	}

	var fullRouteMappingsList []RouteMapping
	warnings, err := client.paginate(request, RouteMapping{}, func(item interface{}) error {
		if routeMapping, ok := item.(RouteMapping); ok {
			fullRouteMappingsList = append(fullRouteMappingsList, routeMapping)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   RouteMapping{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullRouteMappingsList, warnings, err
}
