package actionerror

import "fmt"

type AccessRuleNotFoundError struct {
	Selector string
}

func (e AccessRuleNotFoundError) Error() string {
	return fmt.Sprintf("Access rule with selector '%s' not found.", e.Selector)
}
