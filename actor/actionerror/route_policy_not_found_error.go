package actionerror

import "fmt"

type RoutePolicyNotFoundError struct {
	Source string
}

func (e RoutePolicyNotFoundError) Error() string {
	return fmt.Sprintf("Route policy with source '%s' not found.", e.Source)
}
