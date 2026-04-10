package actionerror

import "fmt"

type AccessRuleNotFoundError struct {
	Name string
}

func (e AccessRuleNotFoundError) Error() string {
	return fmt.Sprintf("Access rule '%s' not found.", e.Name)
}
