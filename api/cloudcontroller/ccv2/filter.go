package ccv2

import (
	"fmt"
	"net/url"
	"strings"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
)

// Filter is a type of filter that can be passed to specific request to narrow
// down the return set.
type Filter struct {
	// Type is the component that determines what the query is filtered by.
	Type constant.FilterType

	// Operator is the component that determines how the the query will be filtered.
	Operator constant.FilterOperator

	// Values is the component that determines what values are filtered.
	Values []string
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
