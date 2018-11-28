package ccerror

type ServiceBrokerCatalogInvalidError struct {
	Message string
}

func (e ServiceBrokerCatalogInvalidError) Error() string {
	return e.Message
}
