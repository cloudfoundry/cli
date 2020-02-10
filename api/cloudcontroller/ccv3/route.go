package ccv3

import (
	"bytes"
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/resources"
)

func (client Client) CreateRoute(route resources.Route) (resources.Route, Warnings, error) {
	bodyBytes, err := json.Marshal(route)
	if err != nil {
		return resources.Route{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PostRouteRequest,
		Body:        bytes.NewReader(bodyBytes),
	})

	if err != nil {
		return resources.Route{}, nil, err
	}

	var ccRoute resources.Route
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &ccRoute,
	}

	err = client.connection.Make(request, &response)

	return ccRoute, response.Warnings, err
}

func (client Client) DeleteOrphanedRoutes(spaceGUID string) (JobURL, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.DeleteOrphanedRoutesRequest,
		URIParams: map[string]string{
			"space_guid": spaceGUID,
		},
		Query: []Query{{Key: UnmappedFilter, Values: []string{"true"}}},
	})
	if err != nil {
		return "", nil, err
	}

	response := cloudcontroller.Response{}
	err = client.connection.Make(request, &response)

	return JobURL(response.ResourceLocationURL), response.Warnings, err
}

func (client Client) DeleteRoute(routeGUID string) (JobURL, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		URIParams: map[string]string{
			"route_guid": routeGUID,
		},
		RequestName: internal.DeleteRouteRequest,
	})
	if err != nil {
		return "", nil, err
	}

	response := cloudcontroller.Response{}
	err = client.connection.Make(request, &response)

	return JobURL(response.ResourceLocationURL), response.Warnings, err
}

func (client Client) GetApplicationRoutes(appGUID string) ([]resources.Route, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetApplicationRoutesRequest,
		URIParams:   internal.Params{"app_guid": appGUID},
	})
	if err != nil {
		return nil, nil, err
	}

	var fullRoutesList []resources.Route
	warnings, err := client.paginate(request, resources.Route{}, func(item interface{}) error {
		if route, ok := item.(resources.Route); ok {
			fullRoutesList = append(fullRoutesList, route)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   resources.Route{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullRoutesList, warnings, err
}

func (client Client) GetRouteDestinations(routeGUID string) ([]resources.RouteDestination, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetRouteDestinationsRequest,
		URIParams:   internal.Params{"route_guid": routeGUID},
	})

	if err != nil {
		return nil, nil, err
	}

	var destinationResponse struct {
		Destinations []resources.RouteDestination `json:"destinations"`
	}

	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &destinationResponse,
	}

	err = client.connection.Make(request, &response)
	return destinationResponse.Destinations, response.Warnings, err
}

func (client Client) GetRoutes(query ...Query) ([]resources.Route, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetRoutesRequest,
		Query:       query,
	})
	if err != nil {
		return nil, nil, err
	}

	var fullRoutesList []resources.Route
	warnings, err := client.paginate(request, resources.Route{}, func(item interface{}) error {
		if route, ok := item.(resources.Route); ok {
			fullRoutesList = append(fullRoutesList, route)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   resources.Route{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullRoutesList, warnings, err
}

func (client Client) MapRoute(routeGUID string, appGUID string) (Warnings, error) {
	type destinationProcess struct {
		ProcessType string `json:"process_type"`
	}

	type destinationApp struct {
		GUID    string              `json:"guid"`
		Process *destinationProcess `json:"process,omitempty"`
	}
	type destination struct {
		App destinationApp `json:"app"`
	}

	type body struct {
		Destinations []destination `json:"destinations"`
	}

	requestBody := body{
		Destinations: []destination{
			{App: destinationApp{GUID: appGUID}},
		},
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.MapRouteRequest,
		URIParams: map[string]string{
			"route_guid": routeGUID,
		},
		Body: bytes.NewReader(bodyBytes),
	})
	if err != nil {
		return nil, err
	}

	response := cloudcontroller.Response{}
	err = client.connection.Make(request, &response)

	return response.Warnings, err
}

func (client Client) UnmapRoute(routeGUID string, destinationGUID string) (Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.UnmapRouteRequest,
		URIParams: map[string]string{
			"route_guid":       routeGUID,
			"destination_guid": destinationGUID,
		},
	})
	if err != nil {
		return nil, err
	}

	response := cloudcontroller.Response{}
	err = client.connection.Make(request, &response)

	return response.Warnings, err
}
