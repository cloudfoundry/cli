package strategy

import "github.com/cloudfoundry/cli/cf/api/resources"

type EventsEndpointStrategy interface {
	EventsURL(appGuid string, limit int64) string
	EventsResource() resources.EventResource
}

type eventsEndpointStrategy struct{}

func (s eventsEndpointStrategy) EventsURL(appGuid string, limit int64) string {
	return buildURL(v2("apps", appGuid, "events"), params{
		resultsPerPage: limit,
	})
}

func (s eventsEndpointStrategy) EventsResource() resources.EventResource {
	return resources.EventResourceOldV2{}
}

type globalEventsEndpointStrategy struct{}

func (s globalEventsEndpointStrategy) EventsURL(appGuid string, limit int64) string {
	return buildURL(v2("events"), params{
		resultsPerPage: limit,
		orderDirection: "desc",
		q:              map[string]string{"actee": appGuid},
	})
}

func (s globalEventsEndpointStrategy) EventsResource() resources.EventResource {
	return resources.EventResourceNewV2{}
}
