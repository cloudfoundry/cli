package ccv3

import (
	"bytes"
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
)

type Route struct {
	GUID       string
	SpaceGUID  string
	DomainGUID string
	Host       string
	Path       string
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
		ccR.Relationships = &Relationships{RelationshipData{Data{GUID: r.SpaceGUID}},
			RelationshipData{Data{GUID: r.DomainGUID}}}
	}
	return json.Marshal(ccR)
}

func (r *Route) UnmarshalJSON(data []byte) error {
	var alias struct {
		GUID string `json:"guid,omitempty"`
		Host string `json:"host,omitempty"`
		Path string `json:"path,omitempty"`

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

	return nil
}

func (client Client) CreateRoute(route Route) (Route, Warnings, error) {
	bodyBytes, err := json.Marshal(route)
	if err != nil {
		return Route{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PostRouteRequest,
		Body:        bytes.NewReader(bodyBytes),
	})

	if err != nil {
		return Route{}, nil, err
	}

	var ccRoute Route
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &ccRoute,
	}

	err = client.connection.Make(request, &response)

	return ccRoute, response.Warnings, err
}
