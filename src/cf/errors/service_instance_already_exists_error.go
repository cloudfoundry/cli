package errors

type ServiceInstanceAlreadyExistsError struct {
	ServiceInstanceName string
}

func NewServiceInstanceAlreadyExistsError(name string) *ServiceInstanceAlreadyExistsError {
	return &ServiceInstanceAlreadyExistsError{
		ServiceInstanceName: name,
	}
}

func (err *ServiceInstanceAlreadyExistsError) Error() string {
	return "Service " + err.ServiceInstanceName + " already exists"
}
