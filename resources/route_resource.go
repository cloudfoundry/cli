package resources

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
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
