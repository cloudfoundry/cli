package ccerror

import (
	"fmt"
	"strings"
)

type ServiceOfferingNameAmbiguityError struct {
	ServiceOfferingName string
	ServiceBrokerNames  []string
}

func (e ServiceOfferingNameAmbiguityError) Error() string {
	const msg = "Service '%s' is provided by multiple service brokers%s"
	switch len(e.ServiceBrokerNames) {
	case 0:
		return fmt.Sprintf(msg, e.ServiceOfferingName, ".")
	default:
		return fmt.Sprintf(msg, e.ServiceOfferingName, ": "+strings.Join(e.ServiceBrokerNames, ", "))
	}
}
