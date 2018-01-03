package ccv2

import (
	"fmt"
	"net/url"
	"strings"
)

// QQueryFilter is the type of filter a QQuery uses.
type QQueryFilter string

// QQueryOperator is the type of operation a QQuery uses.
type QQueryOperator string

const (
	// AppGUIDFilter is the name of the 'app_guid' filter.
	AppGUIDFilter QQueryFilter = "app_guid"
	// DomainGUIDFilter is the name of the 'domain_guid' filter.
	DomainGUIDFilter QQueryFilter = "domain_guid"
	// OrganizationGUIDFilter is the name of the 'organization_guid' filter.
	OrganizationGUIDFilter QQueryFilter = "organization_guid"
	// RouteGUIDFilter is the name of the 'route_guid' filter.
	RouteGUIDFilter QQueryFilter = "route_guid"
	// ServiceInstanceGUIDFilter is the name of the 'service_instance_guid' filter.
	ServiceInstanceGUIDFilter QQueryFilter = "service_instance_guid"
	// SpaceGUIDFilter is the name of the 'space_guid' filter.
	SpaceGUIDFilter QQueryFilter = "space_guid"

	// NameFilter is the name of the 'name' filter.
	NameFilter QQueryFilter = "name"
	// HostFilter is the name of the 'host' filter.
	HostFilter QQueryFilter = "host"
	// PathFilter is the name of the 'path' filter.
	PathFilter QQueryFilter = "path"
	// PortFilter is the name of the 'port' filter.
	PortFilter QQueryFilter = "port"
)

const (
	// EqualOperator is the query equal operator.
	EqualOperator QQueryOperator = ":"

	// InOperator is the query "IN" operator.
	InOperator QQueryOperator = " IN "
)

// QQuery is a type of filter that can be passed to specific request to narrow
// down the return set.
type QQuery struct {
	Filter   QQueryFilter
	Operator QQueryOperator
	Values   []string
}

func (query QQuery) format() string {
	return fmt.Sprintf("%s%s%s", query.Filter, query.Operator, strings.Join(query.Values, ","))
}

// FormatQueryParameters converts a Query object into a collection that
// cloudcontroller.Request can accept.
func FormatQueryParameters(queries []QQuery) url.Values {
	params := url.Values{"q": []string{}}
	for _, query := range queries {
		params["q"] = append(params["q"], query.format())
	}

	return params
}
