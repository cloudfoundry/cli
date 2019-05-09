package actionerror

import "fmt"

type ServiceBrokerNotFoundKey string

const (
	KeyName ServiceBrokerNotFoundKey = "name"
	KeyGUID ServiceBrokerNotFoundKey = "GUID"
)

type ServiceBrokerNotFoundError struct {
	Key   ServiceBrokerNotFoundKey
	Value string
}

func (e ServiceBrokerNotFoundError) Error() string {
	return fmt.Sprintf("Service broker with %s '%s' not found.\nTIP: Use 'cf service-brokers' to see a list of available brokers.", e.Key, e.Value)
}
