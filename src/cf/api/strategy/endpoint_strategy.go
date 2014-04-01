package strategy

import (
	"cf/api/resources"
	"fmt"
	"net/url"
	"strconv"
)

type EndpointStrategy struct {
	EventsURL      func(appGuid string, limit uint64) string
	EventsResource func() resources.EventResource
	DomainsURL     func(orgGuid string) string
}

func NewEndpointStrategy(versionString string) (EndpointStrategy, error) {
	version, err := ParseVersion(versionString)
	if err != nil {
		return EndpointStrategy{}, err
	}

	strategy := EndpointStrategy{}
	setupEvents(&strategy, version)
	setupDomains(&strategy, version)

	return strategy, nil
}

func setupEvents(strategy *EndpointStrategy, version Version) {
	if version.GreaterThanOrEqualTo(Version{2, 2, 0}) {
		strategy.EventsResource = func() resources.EventResource {
			return resources.EventResourceNewV2{}
		}

		strategy.EventsURL = func(appGuid string, limit uint64) string {
			queryParams := url.Values{
				"results-per-page": []string{strconv.FormatUint(limit, 10)},
				"order-direction":  []string{"desc"},
				"q":                []string{"actee:" + appGuid},
			}

			return fmt.Sprintf("/v2/events?%s", queryParams.Encode())
		}
	} else {
		strategy.EventsResource = func() resources.EventResource {
			return resources.EventResourceOldV2{}
		}

		strategy.EventsURL = func(appGuid string, limit uint64) string {
			queryParams := url.Values{
				"results-per-page": []string{strconv.FormatUint(limit, 10)},
			}
			return fmt.Sprintf("/v2/apps/%s/events?%s", appGuid, queryParams.Encode())
		}
	}
}

func setupDomains(strategy *EndpointStrategy, version Version) {
	if version.GreaterThanOrEqualTo(Version{2, 2, 0}) {
		strategy.DomainsURL = func(orgGuid string) string {
			return fmt.Sprintf("/v2/organizations/%s/domains", orgGuid)
		}
	} else {
		strategy.DomainsURL = func(_ string) string {
			return "/v2/domains"
		}
	}
}
