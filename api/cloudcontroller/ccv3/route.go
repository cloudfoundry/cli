package ccv3

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/resources"
)

func (client Client) CreateRoute(route resources.Route) (resources.Route, Warnings, error) {
	var responseBody resources.Route

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.PostRouteRequest,
		RequestBody:  route,
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

func (client Client) DeleteOrphanedRoutes(spaceGUID string) (JobURL, Warnings, error) {
	jobURL, warnings, err := client.MakeRequest(RequestParams{
		RequestName: internal.DeleteOrphanedRoutesRequest,
		URIParams:   internal.Params{"space_guid": spaceGUID},
		Query:       []Query{{Key: UnmappedFilter, Values: []string{"true"}}},
	})

	return jobURL, warnings, err
}

func (client Client) DeleteRoute(routeGUID string) (JobURL, Warnings, error) {
	jobURL, warnings, err := client.MakeRequest(RequestParams{
		RequestName: internal.DeleteRouteRequest,
		URIParams:   internal.Params{"route_guid": routeGUID},
	})

	return jobURL, warnings, err
}

func (client Client) GetApplicationRoutes(appGUID string) ([]resources.Route, Warnings, error) {
	var routes []resources.Route

	_, warnings, err := client.MakeListRequest(RequestParams{
		RequestName:  internal.GetApplicationRoutesRequest,
		URIParams:    internal.Params{"app_guid": appGUID},
		ResponseBody: resources.Route{},
		AppendToList: func(item interface{}) error {
			routes = append(routes, item.(resources.Route))
			return nil
		},
	})

	return routes, warnings, err
}

func (client Client) GetRouteDestinations(routeGUID string) ([]resources.RouteDestination, Warnings, error) {
	var responseBody struct {
		Destinations []resources.RouteDestination `json:"destinations"`
	}

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.GetRouteDestinationsRequest,
		URIParams:    internal.Params{"route_guid": routeGUID},
		ResponseBody: &responseBody,
	})

	return responseBody.Destinations, warnings, err
}

func (client Client) GetRoutes(query ...Query) ([]resources.Route, Warnings, error) {
	var routes []resources.Route

	_, warnings, err := client.MakeListRequest(RequestParams{
		RequestName:  internal.GetRoutesRequest,
		Query:        query,
		ResponseBody: resources.Route{},
		AppendToList: func(item interface{}) error {
			routes = append(routes, item.(resources.Route))
			return nil
		},
	})

	return routes, warnings, err
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

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName: internal.MapRouteRequest,
		URIParams:   internal.Params{"route_guid": routeGUID},
		RequestBody: &requestBody,
	})

	return warnings, err
}

func (client Client) UnmapRoute(routeGUID string, destinationGUID string) (Warnings, error) {
	var responseBody Build

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.UnmapRouteRequest,
		URIParams:    internal.Params{"route_guid": routeGUID, "destination_guid": destinationGUID},
		ResponseBody: &responseBody,
	})

	return warnings, err
}
