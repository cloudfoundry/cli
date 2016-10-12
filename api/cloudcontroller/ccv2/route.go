package ccv2

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
)

type Route struct {
	GUID         string
	Host         string
	Domain       string
	Path         string
	Port         int
	DomainFields Domain
}

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
	route.DomainFields.GUID = ccRoute.Entity.DomainGUID
	return nil
}

func (client *CloudControllerClient) GetSpaceRoutes(spaceGUID string) ([]Route, Warnings, error) {
	request := cloudcontroller.Request{
		RequestName: internal.RoutesFromSpaceRequest,
		Params:      map[string]string{"space_guid": spaceGUID},
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
		request = cloudcontroller.Request{
			URI:    wrapper.NextURL,
			Method: "GET",
		}
	}

	return fullRoutesList, fullWarningsList, nil
}

func (client *CloudControllerClient) DeleteRoute(routeGUID string) (Warnings, error) {
	request := cloudcontroller.Request{
		RequestName: internal.DeleteRouteRequest,
		Params:      map[string]string{"route_guid": routeGUID},
	}

	var response cloudcontroller.Response
	err := client.connection.Make(request, &response)
	return response.Warnings, err
}
