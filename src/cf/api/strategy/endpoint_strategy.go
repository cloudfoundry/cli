package strategy

import (
	"cf/api/resources"
	"fmt"
	"net/url"
	"strconv"
)

type EndpointStrategy interface {
	EventsURL(appGuid string, limit uint64) string
	EventsResource() resources.EventResource
}

type endpointStrategy struct {
	version Version
}

func NewEndpointStrategy(versionString string) (EndpointStrategy, error) {
	version, err := ParseVersion(versionString)
	return endpointStrategy{version: version}, err
}

func (strategy endpointStrategy) EventsURL(appGuid string, limit uint64) string {
	if strategy.version.GreaterThanOrEqualTo(Version{2, 2, 0}) {
		queryParams := url.Values{
			"results-per-page": []string{strconv.FormatUint(limit, 10)},
			"order-direction":  []string{"desc"},
			"q":                []string{"actee:" + appGuid},
		}

		return fmt.Sprintf("/v2/events?%s", queryParams.Encode())
	} else {
		queryParams := url.Values{
			"results-per-page": []string{strconv.FormatUint(limit, 10)},
		}
		return fmt.Sprintf("/v2/apps/%s/events?%s", appGuid, queryParams.Encode())
	}
}

func (strategy endpointStrategy) EventsResource() resources.EventResource {
	if strategy.version.GreaterThanOrEqualTo(Version{2, 2, 0}) {
		return resources.EventResourceNewV2{}
	} else {
		return resources.EventResourceOldV2{}
	}
}
