package ccv2

import (
	"fmt"
	"net/url"
	"strings"
)

// FilterType is the type of filter a Filter uses.
type FilterType string

// FilterOperator is the type of operation a Filter uses.
type FilterOperator string

const (
	// AppGUIDFilter is the name of the 'app_guid' filter.
	AppGUIDFilter FilterType = "app_guid"
	// DomainGUIDFilter is the name of the 'domain_guid' filter.
	DomainGUIDFilter FilterType = "domain_guid"
	// OrganizationGUIDFilter is the name of the 'organization_guid' filter.
	OrganizationGUIDFilter FilterType = "organization_guid"
	// RouteGUIDFilter is the name of the 'route_guid' filter.
	RouteGUIDFilter FilterType = "route_guid"
	// ServiceInstanceGUIDFilter is the name of the 'service_instance_guid' filter.
	ServiceInstanceGUIDFilter FilterType = "service_instance_guid"
	// SpaceGUIDFilter is the name of the 'space_guid' filter.
	SpaceGUIDFilter FilterType = "space_guid"

	// NameFilter is the name of the 'name' filter.
	NameFilter FilterType = "name"
	// HostFilter is the name of the 'host' filter.
	HostFilter FilterType = "host"
	// PathFilter is the name of the 'path' filter.
	PathFilter FilterType = "path"
	// PortFilter is the name of the 'port' filter.
	PortFilter FilterType = "port"
)

const (
	// EqualOperator is the Filter's equal operator.
	EqualOperator FilterOperator = ":"

	// InOperator is the Filter's "IN" operator.
	InOperator FilterOperator = " IN "
)

// Filter is a type of filter that can be passed to specific request to narrow
// down the return set.
type Filter struct {
	Type     FilterType
	Operator FilterOperator
	Values   []string
}

func (filter Filter) format() string {
	return fmt.Sprintf("%s%s%s", filter.Type, filter.Operator, strings.Join(filter.Values, ","))
}

// ConvertFilterParameters converts a Filter object into a collection that
// cloudcontroller.Request can accept.
func ConvertFilterParameters(filters []Filter) url.Values {
	params := url.Values{"q": []string{}}
	for _, filter := range filters {
		params["q"] = append(params["q"], filter.format())
	}

	return params
}
