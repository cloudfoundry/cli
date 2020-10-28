package actionerror

import "fmt"

type ServiceKeyNotFoundError struct {
	KeyName             string
	ServiceInstanceName string
}

func NewServiceKeyNotFoundError(keyName, serviceInstanceName string) error {
	return ServiceKeyNotFoundError{
		KeyName:             keyName,
		ServiceInstanceName: serviceInstanceName,
	}
}

func (e ServiceKeyNotFoundError) Error() string {
	return fmt.Sprintf("No service key %s found for service instance %s", e.KeyName, e.ServiceInstanceName)
}
