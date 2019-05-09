package actionerror

import "fmt"

type ServiceBrokerNotFoundError struct {
	Name string
}

func (e ServiceBrokerNotFoundError) Error() string {
	return fmt.Sprintf("Service broker '%s' not found.\nTIP: Use 'cf service-brokers' to see a list of available brokers.", e.Name)
}
