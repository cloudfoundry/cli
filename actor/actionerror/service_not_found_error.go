package actionerror

import "fmt"

type ServiceNotFoundError struct {
	Name, Broker string
}

func (e ServiceNotFoundError) Error() string {
	if e.Name != "" && e.Broker != "" {
		return fmt.Sprintf("Service offering '%s' for service broker '%s' not found.", e.Name, e.Broker)
	}

	if e.Name != "" {
		return fmt.Sprintf("Service offering '%s' not found.", e.Name)
	}

	if e.Broker != "" {
		return fmt.Sprintf("No service offerings found for service broker '%s'.", e.Broker)
	}

	return "No service offerings found."
}
