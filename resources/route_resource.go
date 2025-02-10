package resources

import (
	"encoding/json"
	"strings"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
)

type RouteDestinationApp struct {
	GUID    string
	Process struct {
		Type string
	}
}

type RouteDestination struct {
	GUID     string
	App      RouteDestinationApp
	Port     int
	Protocol string
}

type Route struct {
	GUID         string
	SpaceGUID    string
	DomainGUID   string
	Host         string
	Path         string
	Protocol     string
	Port         int
	URL          string
	Destinations []RouteDestination
	Metadata     *Metadata
	Options      map[string]*string
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
		GUID          string             `json:"guid,omitempty"`
		Host          string             `json:"host,omitempty"`
		Path          string             `json:"path,omitempty"`
		Protocol      string             `json:"protocol,omitempty"`
		Port          int                `json:"port,omitempty"`
		Relationships *Relationships     `json:"relationships,omitempty"`
		Options       map[string]*string `json:"options,omitempty"`
	}

	ccR := ccRoute{
		GUID:     r.GUID,
		Host:     r.Host,
		Path:     r.Path,
		Protocol: r.Protocol,
		Port:     r.Port,
		Options:  r.Options,
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
		Protocol     string             `json:"protocol,omitempty"`
		Host         string             `json:"host,omitempty"`
		Path         string             `json:"path,omitempty"`
		Port         int                `json:"port,omitempty"`
		URL          string             `json:"url,omitempty"`
		Destinations []RouteDestination `json:"destinations,omitempty"`
		Metadata     *Metadata          `json:"metadata,omitempty"`
		Options      map[string]*string `json:"options,omitempty"`

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
	r.Protocol = alias.Protocol
	r.Host = alias.Host
	r.SpaceGUID = alias.Relationships.Space.Data.GUID
	r.DomainGUID = alias.Relationships.Domain.Data.GUID
	r.Path = alias.Path
	r.Port = alias.Port
	r.URL = alias.URL
	r.Destinations = alias.Destinations
	r.Metadata = alias.Metadata
	r.Options = alias.Options

	return nil
}

func (r *Route) FormattedOptions() string {
	var routeOpts = []string{}
	formattedOptions := ""
	if r.Options != nil && len(r.Options) > 0 {
		for optKey, optVal := range r.Options {
			routeOpts = append(routeOpts, optKey+"="+*optVal)
		}
		formattedOptions = " {" + strings.Join(routeOpts, ", ") + "}"
	}
	return formattedOptions
}
