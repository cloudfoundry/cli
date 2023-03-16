package resources

import (
	"code.cloudfoundry.org/jsonry"
)

type ServiceInstanceUsageSummaryList struct {
	UsageSummary []ServiceInstanceUsageSummary `jsonry:"usage_summary"`
}

func (s *ServiceInstanceUsageSummaryList) UnmarshalJSON(data []byte) error {
	return jsonry.Unmarshal(data, s)
}

type ServiceInstanceUsageSummary struct {
	SpaceGUID     string `jsonry:"space.guid"`
	BoundAppCount int    `jsonry:"bound_app_count"`
}

func (s *ServiceInstanceUsageSummary) UnmarshalJSON(data []byte) error {
	return jsonry.Unmarshal(data, s)
}
