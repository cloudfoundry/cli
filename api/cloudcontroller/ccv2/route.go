package ccv2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
	"code.cloudfoundry.org/cli/types"
)

// Route represents a Cloud Controller Route.
type Route struct {

	// GUID is the unique Route identifier.
	GUID string `json:"-"`

	// Host is the hostname of the route.
	Host string `json:"host,omitempty"`

	// Path is the path of the route.
	Path string `json:"path,omitempty"`

	// Port is the port number of the route.
	Port types.NullInt `json:"port,omitempty"`

	// DomainGUID is the unique Domain identifier.
	DomainGUID string `json:"domain_guid"`

	// SpaceGUID is the unique Space identifier.
	SpaceGUID string `json:"space_guid"`
}

// UnmarshalJSON helps unmarshal a Cloud Controller Route response.
func (route *Route) UnmarshalJSON(data []byte) error {
	var ccRoute struct {
		Metadata internal.Metadata `json:"metadata"`
		Entity   struct {
			Host       string        `json:"host"`
			Path       string        `json:"path"`
			Port       types.NullInt `json:"port"`
			DomainGUID string        `json:"domain_guid"`
			SpaceGUID  string        `json:"space_guid"`
		} `json:"entity"`
	}
	err := cloudcontroller.DecodeJSON(data, &ccRoute)
	if err != nil {
		return err
	}

	route.GUID = ccRoute.Metadata.GUID
	route.Host = ccRoute.Entity.Host
	route.Path = ccRoute.Entity.Path
	route.Port = ccRoute.Entity.Port
	route.DomainGUID = ccRoute.Entity.DomainGUID
	route.SpaceGUID = ccRoute.Entity.SpaceGUID
	return nil
}

// CheckRoute returns true if the route exists in the CF instance.
func (client *Client) CheckRoute(route Route) (bool, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetRouteReservedRequest,
		URIParams:   map[string]string{"domain_guid": route.DomainGUID},
	})
	if err != nil {
		return false, nil, err
	}

	queryParams := url.Values{}
	if route.Host != "" {
		queryParams.Add("host", route.Host)
	}
	if route.Path != "" {
		queryParams.Add("path", route.Path)
	}
	if route.Port.IsSet {
		queryParams.Add("port", fmt.Sprint(route.Port.Value))
	}
	request.URL.RawQuery = queryParams.Encode()

	var response cloudcontroller.Response
	err = client.connection.Make(request, &response)
	if _, ok := err.(ccerror.ResourceNotFoundError); ok {
		return false, response.Warnings, nil
	}

	return response.HTTPResponse.StatusCode == http.StatusNoContent, response.Warnings, err
}

// CreateRoute creates the route with the given properties; SpaceGUID and
// DomainGUID are required Route properties. Additional configuration rules:
// - generatePort = true to generate a random port on the cloud controller.
// - generatePort takes precedence over the provided port. Setting the port and
// generatePort only works with CC API 2.53.0 or higher and when TCP router
// groups are enabled.
func (client *Client) CreateRoute(route Route, generatePort bool) (Route, Warnings, error) {
	body, err := json.Marshal(route)
	if err != nil {
		return Route{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PostRouteRequest,
		Body:        bytes.NewReader(body),
	})
	if err != nil {
		return Route{}, nil, err
	}

	if generatePort {
		query := url.Values{}
		query.Add("generate_port", "true")
		request.URL.RawQuery = query.Encode()
	}

	var updatedRoute Route
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &updatedRoute,
	}

	err = client.connection.Make(request, &response)
	return updatedRoute, response.Warnings, err
}

// DeleteRoute deletes the Route associated with the provided Route GUID.
func (client *Client) DeleteRoute(routeGUID string) (Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.DeleteRouteRequest,
		URIParams:   map[string]string{"route_guid": routeGUID},
	})
	if err != nil {
		return nil, err
	}

	var response cloudcontroller.Response
	err = client.connection.Make(request, &response)
	return response.Warnings, err
}

// DeleteRouteApplication removes the link between the route and application.
func (client *Client) DeleteRouteApplication(routeGUID string, appGUID string) (Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.DeleteRouteAppRequest,
		URIParams: map[string]string{
			"app_guid":   appGUID,
			"route_guid": routeGUID,
		},
	})
	if err != nil {
		return nil, err
	}

	var response cloudcontroller.Response
	err = client.connection.Make(request, &response)
	return response.Warnings, err
}

func (client *Client) DeleteUnmappedRoutes(spaceGUID string) (Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.DeleteUnmappedRoutesRequest,
		URIParams:   map[string]string{"space_guid": spaceGUID},
	})
	if err != nil {
		return nil, err
	}

	var response cloudcontroller.Response
	err = client.connection.Make(request, &response)
	return response.Warnings, err
}

// GetApplicationRoutes returns a list of Routes associated with the provided
// Application GUID, and filtered by the provided filters.
func (client *Client) GetApplicationRoutes(appGUID string, filters ...Filter) ([]Route, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetAppRoutesRequest,
		URIParams:   map[string]string{"app_guid": appGUID},
		Query:       ConvertFilterParameters(filters),
	})
	if err != nil {
		return nil, nil, err
	}

	var fullRoutesList []Route
	warnings, err := client.paginate(request, Route{}, func(item interface{}) error {
		if route, ok := item.(Route); ok {
			fullRoutesList = append(fullRoutesList, route)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   Route{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullRoutesList, warnings, err
}

// GetRoute returns a route with the provided guid.
func (client *Client) GetRoute(guid string) (Route, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetRouteRequest,
		URIParams:   Params{"route_guid": guid},
	})
	if err != nil {
		return Route{}, nil, err
	}

	var route Route
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &route,
	}

	err = client.connection.Make(request, &response)
	return route, response.Warnings, err
}

// GetRoutes returns a list of Routes based off of the provided filters.
func (client *Client) GetRoutes(filters ...Filter) ([]Route, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetRoutesRequest,
		Query:       ConvertFilterParameters(filters),
	})
	if err != nil {
		return nil, nil, err
	}

	var fullRoutesList []Route
	warnings, err := client.paginate(request, Route{}, func(item interface{}) error {
		if route, ok := item.(Route); ok {
			fullRoutesList = append(fullRoutesList, route)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   Route{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullRoutesList, warnings, err
}

// GetSpaceRoutes returns a list of Routes associated with the provided Space
// GUID, and filtered by the provided filters.
func (client *Client) GetSpaceRoutes(spaceGUID string, filters ...Filter) ([]Route, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetSpaceRoutesRequest,
		URIParams:   map[string]string{"space_guid": spaceGUID},
		Query:       ConvertFilterParameters(filters),
	})
	if err != nil {
		return nil, nil, err
	}

	var fullRoutesList []Route
	warnings, err := client.paginate(request, Route{}, func(item interface{}) error {
		if route, ok := item.(Route); ok {
			fullRoutesList = append(fullRoutesList, route)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   Route{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullRoutesList, warnings, err
}

// UpdateRouteApplication creates a link between the route and application.
func (client *Client) UpdateRouteApplication(routeGUID string, appGUID string) (Route, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PutRouteAppRequest,
		URIParams: map[string]string{
			"app_guid":   appGUID,
			"route_guid": routeGUID,
		},
	})
	if err != nil {
		return Route{}, nil, err
	}

	var route Route
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &route,
	}
	err = client.connection.Make(request, &response)

	return route, response.Warnings, err
}

func (client *Client) checkRouteDeprecated(domainGUID string, host string, path string) (bool, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetRouteReservedDeprecatedRequest,
		URIParams:   map[string]string{"domain_guid": domainGUID, "host": host},
	})
	if err != nil {
		return false, nil, err
	}

	queryParams := url.Values{}
	if path != "" {
		queryParams.Add("path", path)
	}
	request.URL.RawQuery = queryParams.Encode()

	var response cloudcontroller.Response
	err = client.connection.Make(request, &response)
	if _, ok := err.(ccerror.ResourceNotFoundError); ok {
		return false, response.Warnings, nil
	}

	return response.HTTPResponse.StatusCode == http.StatusNoContent, response.Warnings, err
}
