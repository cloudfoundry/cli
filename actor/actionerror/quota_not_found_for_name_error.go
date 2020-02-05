package actionerror

import "fmt"

type QuotaNotFoundForNameError struct {
	Name string
}

func (e QuotaNotFoundForNameError) Error() string {
	return fmt.Sprintf("Organization quota with name '%s' not found.", e.Name)
}
