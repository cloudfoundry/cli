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
	Type     constant.FilterType
	Operator constant.FilterOperator
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
