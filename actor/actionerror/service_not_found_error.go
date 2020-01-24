package actionerror

import "fmt"

type ServiceNotFoundError struct {
	Name, Broker string
}

func (e ServiceNotFoundError) Error() string {
	if e.Broker == "" {
		return fmt.Sprintf("Service offering '%s' not found.", e.Name)
	}
	return fmt.Sprintf("Service offering '%s' for service broker '%s' not found.", e.Name, e.Broker)
}
