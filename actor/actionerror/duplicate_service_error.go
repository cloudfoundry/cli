package actionerror

import "fmt"

type DuplicateServiceError struct {
	Name string
}

func (e DuplicateServiceError) Error() string {
	return fmt.Sprintf("Service '%s' is provided by multiple service brokers. Specify a broker by using the '-b' flag.", e.Name)
}
