package ccv3

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
)

type Route struct {
	GUID         string
	SpaceGUID    string
	DomainGUID   string
	Host         string
	Path         string
	URL          string
	Destinations []RouteDestination
	Metadata     *Metadata
}

func (r Route) MarshalJSON() ([]byte, error) {
	type Data struct {
		GUID string `json:"guid,omitempty"`
	}

	type RelationshipData struct {
		Data Data `json:"data,omitempty"`
	}

	type Relationships struct {
		Space  RelationshipData `json:"space,omitempty"`
		Domain RelationshipData `json:"domain,omitempty"`
	}

	// Building up the request body in ccRoute
	type ccRoute struct {
		GUID          string         `json:"guid,omitempty"`
		Host          string         `json:"host,omitempty"`
		Path          string         `json:"path,omitempty"`
		Relationships *Relationships `json:"relationships,omitempty"`
	}

	ccR := ccRoute{
		GUID: r.GUID,
		Host: r.Host,
		Path: r.Path,
	}

	if r.SpaceGUID != "" {
		ccR.Relationships = &Relationships{
			Space:  RelationshipData{Data{GUID: r.SpaceGUID}},
			Domain: RelationshipData{Data{GUID: r.DomainGUID}},
		}
	}

	return json.Marshal(ccR)
}

func (r *Route) UnmarshalJSON(data []byte) error {
	var alias struct {
		GUID         string             `json:"guid,omitempty"`
		Host         string             `json:"host,omitempty"`
		Path         string             `json:"path,omitempty"`
		URL          string             `json:"url,omitempty"`
		Destinations []RouteDestination `json:"destinations,omitempty"`
		Metadata     *Metadata          `json:"metadata,omitempty"`

		Relationships struct {
			Space struct {
				Data struct {
					GUID string `json:"guid,omitempty"`
				} `json:"data,omitempty"`
			} `json:"space,omitempty"`
			Domain struct {
				Data struct {
					GUID string `json:"guid,omitempty"`
				} `json:"data,omitempty"`
			} `json:"domain,omitempty"`
		} `json:"relationships,omitempty"`
	}

	err := cloudcontroller.DecodeJSON(data, &alias)
	if err != nil {
		return err
	}

	r.GUID = alias.GUID
	r.Host = alias.Host
	r.SpaceGUID = alias.Relationships.Space.Data.GUID
	r.DomainGUID = alias.Relationships.Domain.Data.GUID
	r.Path = alias.Path
	r.URL = alias.URL
	r.Destinations = alias.Destinations
	r.Metadata = alias.Metadata

	return nil
}

type RouteDestinationApp struct {
	GUID    string
	Process struct {
		Type string
	}
}

type RouteDestination struct {
	GUID string
	App  RouteDestinationApp
}

func (client Client) CreateRoute(route Route) (Route, Warnings, error) {
	var responseBody Route

	_, warnings, err := client.makeRequest(requestParams{
		RequestName:  internal.PostRouteRequest,
		RequestBody:  route,
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

func (client Client) DeleteOrphanedRoutes(spaceGUID string) (JobURL, Warnings, error) {
	jobURL, warnings, err := client.makeRequest(requestParams{
		RequestName: internal.DeleteOrphanedRoutesRequest,
		URIParams:   internal.Params{"space_guid": spaceGUID},
		Query:       []Query{{Key: UnmappedFilter, Values: []string{"true"}}},
	})

	return jobURL, warnings, err
}

func (client Client) DeleteRoute(routeGUID string) (JobURL, Warnings, error) {
	jobURL, warnings, err := client.makeRequest(requestParams{
		RequestName: internal.DeleteRouteRequest,
		URIParams:   internal.Params{"route_guid": routeGUID},
	})

	return jobURL, warnings, err
}

func (client Client) GetApplicationRoutes(appGUID string) ([]Route, Warnings, error) {
	var resources []Route

	_, warnings, err := client.makeListRequest(requestParams{
		RequestName:  internal.GetApplicationRoutesRequest,
		URIParams:    internal.Params{"app_guid": appGUID},
		ResponseBody: Route{},
		AppendToList: func(item interface{}) error {
			resources = append(resources, item.(Route))
			return nil
		},
	})

	return resources, warnings, err
}

func (client Client) GetRouteDestinations(routeGUID string) ([]RouteDestination, Warnings, error) {
	var responseBody struct {
		Destinations []RouteDestination `json:"destinations"`
	}

	_, warnings, err := client.makeRequest(requestParams{
		RequestName:  internal.GetRouteDestinationsRequest,
		URIParams:    internal.Params{"route_guid": routeGUID},
		ResponseBody: &responseBody,
	})

	return responseBody.Destinations, warnings, err
}

func (client Client) GetRoutes(query ...Query) ([]Route, Warnings, error) {
	var resources []Route

	_, warnings, err := client.makeListRequest(requestParams{
		RequestName:  internal.GetRoutesRequest,
		Query:        query,
		ResponseBody: Route{},
		AppendToList: func(item interface{}) error {
			resources = append(resources, item.(Route))
			return nil
		},
	})

	return resources, warnings, err
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

	_, warnings, err := client.makeRequest(requestParams{
		RequestName: internal.MapRouteRequest,
		URIParams:   internal.Params{"route_guid": routeGUID},
		RequestBody: &requestBody,
	})

	return warnings, err
}

func (client Client) UnmapRoute(routeGUID string, destinationGUID string) (Warnings, error) {
	var responseBody Build

	_, warnings, err := client.makeRequest(requestParams{
		RequestName:  internal.UnmapRouteRequest,
		URIParams:    internal.Params{"route_guid": routeGUID, "destination_guid": destinationGUID},
		ResponseBody: &responseBody,
	})

	return warnings, err
}
