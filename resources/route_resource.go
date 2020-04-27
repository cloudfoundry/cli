package resources

import "code.cloudfoundry.org/cli/types"

type Route struct {
	GUID         string
	SpaceGUID    string
	DomainGUID   string
	Host         string
	Path         string
	DomainName   string
	SpaceName    string
	URL          string
	Destinations []RouteDestination
	Metadata     *Metadata
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

type Metadata struct {
	Labels map[string]types.NullString `json:"labels,omitempty"`
}

type ResourceMetadata struct {
	Metadata *Metadata `json:"metadata,omitempty"`
}
