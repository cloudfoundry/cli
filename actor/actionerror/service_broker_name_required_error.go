package actionerror

import (
	"fmt"
)

type ServiceBrokerNameRequiredError struct {
	ServiceOfferingName string
}

func (e ServiceBrokerNameRequiredError) Error() string {
	return fmt.Sprintf(
		"Service offering '%s' is provided by multiple service brokers. Specify a broker name by using the '-b' flag.",
		e.ServiceOfferingName,
	)
}
