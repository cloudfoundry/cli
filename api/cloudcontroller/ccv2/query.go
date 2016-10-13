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
	AppGUIDFilter             QueryFilter = "app_guid"
	ServiceInstanceGUIDFilter QueryFilter = "service_instance_guid"
	SpaceGUIDFilter           QueryFilter = "space_guid"

	NameFilter QueryFilter = "name"
)

const (
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
