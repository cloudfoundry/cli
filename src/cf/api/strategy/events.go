package strategy

import (
	"cf/api/resources"
	"fmt"
	"net/url"
	"strconv"
)

type EventsEndpointStrategy interface {
	EventsURL(appGuid string, limit uint64) string
	EventsResource() resources.EventResource
}

type globalEventsEndpointStrategy struct{}
type appScopedEventsEndpointStrategy struct{}

func (strategy globalEventsEndpointStrategy) EventsURL(appGuid string, limit uint64) string {
	queryParams := url.Values{
		"results-per-page": []string{strconv.FormatUint(limit, 10)},
		"order-direction":  []string{"desc"},
		"q":                []string{"actee:" + appGuid},
	}

	return fmt.Sprintf("/v2/events?%s", queryParams.Encode())
}

func (_ globalEventsEndpointStrategy) EventsResource() resources.EventResource {
	return resources.EventResourceNewV2{}
}

func (_ appScopedEventsEndpointStrategy) EventsURL(appGuid string, limit uint64) string {
	queryParams := url.Values{
		"results-per-page": []string{strconv.FormatUint(limit, 10)},
	}
	return fmt.Sprintf("/v2/apps/%s/events?%s", appGuid, queryParams.Encode())
}

func (_ appScopedEventsEndpointStrategy) EventsResource() resources.EventResource {
	return resources.EventResourceOldV2{}
}
