package actionerror

import (
	"fmt"
	"strings"
)

type DuplicateServiceError struct {
	Name           string
	ServiceBrokers []string
}

func (e DuplicateServiceError) Error() string {
	message := " Specify a broker by using the '-b' flag."

	if len(e.ServiceBrokers) != 0 {
		message = fmt.Sprintf("\nSpecify a broker from available brokers %s by using the '-b' flag.", e.availableBrokers())
	}

	return fmt.Sprintf(
		"Service '%s' is provided by multiple service brokers.%s", e.Name, message)
}

func (e DuplicateServiceError) availableBrokers() string {
	var quotedBrokers []string
	for _, broker := range e.ServiceBrokers {
		quotedBrokers = append(quotedBrokers, fmt.Sprintf("'%s'", broker))
	}
	availableBrokers := strings.Join(quotedBrokers, ", ")
	return availableBrokers
}
