package ccv2

import (
	"fmt"
	"net/url"
)

type QueryFilter string
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

type Query struct {
	Filter   QueryFilter
	Operator QueryOperator
	Value    string
}

func (query Query) format() string {
	return fmt.Sprintf("%s%s%s", query.Filter, query.Operator, query.Value)
}

func FormatQueryParameters(queries []Query) url.Values {
	params := url.Values{"q": []string{}}
	for _, query := range queries {
		params["q"] = append(params["q"], query.format())
	}

	return params
}
