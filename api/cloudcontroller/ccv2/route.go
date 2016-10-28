package ccv2

import (
	"encoding/json"
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
)

// Route represents a Cloud Controller Route.
type Route struct {
	GUID       string
	Host       string
	Path       string
	Port       int
	DomainGUID string
}

// UnmarshalJSON helps unmarshal a Cloud Controller Route response.
func (route *Route) UnmarshalJSON(data []byte) error {
	var ccRoute struct {
		Metadata internal.Metadata `json:"metadata"`
		Entity   struct {
			Host       string `json:"host"`
			Path       string `json:"path"`
			Port       int    `json:"port"`
			DomainGUID string `json:"domain_guid"`
		} `json:"entity"`
	}
	if err := json.Unmarshal(data, &ccRoute); err != nil {
		return err
	}

	route.GUID = ccRoute.Metadata.GUID
	route.Host = ccRoute.Entity.Host
	route.Path = ccRoute.Entity.Path
	route.Port = ccRoute.Entity.Port
	route.DomainGUID = ccRoute.Entity.DomainGUID
	return nil
}

// GetSpaceRoutes returns a list of Routes associated with the provided Space
// GUID, and filtered by the provided queries.
func (client *Client) GetSpaceRoutes(spaceGUID string, queryParams []Query) ([]Route, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.RoutesFromSpaceRequest,
		URIParams:   map[string]string{"space_guid": spaceGUID},
		Query:       FormatQueryParameters(queryParams),
	})
	if err != nil {
		return nil, nil, err
	}

	fullRoutesList := []Route{}
	fullWarningsList := Warnings{}

	for {
		var routes []Route
		wrapper := PaginatedWrapper{
			Resources: &routes,
		}
		response := cloudcontroller.Response{
			Result: &wrapper,
		}

		err := client.connection.Make(request, &response)
		fullWarningsList = append(fullWarningsList, response.Warnings...)
		if err != nil {
			return nil, fullWarningsList, err
		}
		fullRoutesList = append(fullRoutesList, routes...)

		if wrapper.NextURL == "" {
			break
		}
		request, err = client.newHTTPRequest(requestOptions{
			URI:    wrapper.NextURL,
			Method: http.MethodGet,
		})
		if err != nil {
			return nil, fullWarningsList, err
		}
	}

	return fullRoutesList, fullWarningsList, nil
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
