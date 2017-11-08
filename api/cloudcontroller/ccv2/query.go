package ccv2

import (
	"fmt"
	"net/url"
	"strings"
)

// QueryFilter is the type of filter a Query uses.
type QueryFilter string

// QueryOperator is the type of operation a Query uses.
type QueryOperator string

const (
	// AppGUIDFilter is the name of the 'app_guid' filter.
	AppGUIDFilter QueryFilter = "app_guid"
	// DomainGUIDFilter is the name of the 'domain_guid' filter.
	DomainGUIDFilter QueryFilter = "domain_guid"
	// OrganizationGUIDFilter is the name of the 'organization_guid' filter.
	OrganizationGUIDFilter QueryFilter = "organization_guid"
	// RouteGUIDFilter is the name of the 'route_guid' filter.
	RouteGUIDFilter QueryFilter = "route_guid"
	// ServiceInstanceGUIDFilter is the name of the 'service_instance_guid' filter.
	ServiceInstanceGUIDFilter QueryFilter = "service_instance_guid"
	// SpaceGUIDFilter is the name of the 'space_guid' filter.
	SpaceGUIDFilter QueryFilter = "space_guid"

	// NameFilter is the name of the 'name' filter.
	NameFilter QueryFilter = "name"
	// HostFilter is the name of the 'host' filter.
	HostFilter QueryFilter = "host"
	// PathFilter is the name of the 'path' filter.
	PathFilter QueryFilter = "path"
	// PortFilter is the name of the 'port' filter.
	PortFilter QueryFilter = "port"
)

const (
	// EqualOperator is the query equal operator.
	EqualOperator QueryOperator = ":"

	// InOperator is the query "IN" operator.
	InOperator QueryOperator = " IN "
)

// Query is a type of filter that can be passed to specific request to narrow
// down the return set.
type Query struct {
	Filter   QueryFilter
	Operator QueryOperator
	Values   []string
}

func (query Query) format() string {
	return fmt.Sprintf("%s%s%s", query.Filter, query.Operator, strings.Join(query.Values, ","))
}

// FormatQueryParameters converts a Query object into a collection that
// cloudcontroller.Request can accept.
func FormatQueryParameters(queries []Query) url.Values {
	params := url.Values{"q": []string{}}
	for _, query := range queries {
		params["q"] = append(params["q"], query.format())
	}

	return params
}
