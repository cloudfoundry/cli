package actionerror

import "fmt"

type ServiceAndBrokerCombinationNotFoundError struct {
	BrokerName  string
	ServiceName string
}

func (e ServiceAndBrokerCombinationNotFoundError) Error() string {
	return fmt.Sprintf("Service '%s' provided by service broker '%s' not found.\n", e.ServiceName, e.BrokerName)
}
