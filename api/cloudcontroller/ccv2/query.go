package ccv2

import (
	"fmt"
	"net/url"
)

// QueryFilter is the type of filter a Query uses.
type QueryFilter string

// QueryOperator is the type of operation a Query uses.
type QueryOperator string

const (
	// AppGUIDFilter is the name of the App GUID filter.
	AppGUIDFilter QueryFilter = "app_guid"
	// OrganizationGUIDFilter is the name of the organization GUID filter.
	OrganizationGUIDFilter QueryFilter = "organization_guid"
	// RouteGUIDFilter is the name of the route GUID filter.
	RouteGUIDFilter QueryFilter = "route_guid"
	// ServiceInstanceGUIDFilter is the name of the service instance GUID filter.
	ServiceInstanceGUIDFilter QueryFilter = "service_instance_guid"
	// SpaceGUIDFilter is the name of the space GUID filter.
	SpaceGUIDFilter QueryFilter = "space_guid"

	// NameFilter is the name of the name filter.
	NameFilter QueryFilter = "name"
)

const (
	// EqualOperator is the query equal operator.
	EqualOperator QueryOperator = ":"
)

// Query is a type of filter that can be passed to specific request to narrow
// down the return set.
type Query struct {
	Filter   QueryFilter
	Operator QueryOperator
	Value    string
}

func (query Query) format() string {
	return fmt.Sprintf("%s%s%s", query.Filter, query.Operator, query.Value)
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
