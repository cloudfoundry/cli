package strategy

import "cf/api/resources"

type EventsEndpointStrategy interface {
	EventsURL(appGuid string, limit uint64) string
	EventsResource() resources.EventResource
}

type eventsEndpointStrategy struct{}

func (_ eventsEndpointStrategy) EventsURL(appGuid string, limit uint64) string {
	return buildURL(v2("apps", appGuid, "events"), query{
		resultsPerPage: limit,
	})
}

func (_ eventsEndpointStrategy) EventsResource() resources.EventResource {
	return resources.EventResourceOldV2{}
}

type globalEventsEndpointStrategy struct{}

func (strategy globalEventsEndpointStrategy) EventsURL(appGuid string, limit uint64) string {
	return buildURL(v2("events"), query{
		resultsPerPage: limit,
		orderDirection: "desc",
		q:              map[string]string{"actee": appGuid},
	})
}

func (_ globalEventsEndpointStrategy) EventsResource() resources.EventResource {
	return resources.EventResourceNewV2{}
}
