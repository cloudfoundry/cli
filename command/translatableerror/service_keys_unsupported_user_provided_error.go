package translatableerror

type ServiceKeysNotSupportedWithUserProvidedServiceInstances struct{}

func (ServiceKeysNotSupportedWithUserProvidedServiceInstances) Error() string {
	return "Service keys are not supported for user-provided service instances"
}

func (e ServiceKeysNotSupportedWithUserProvidedServiceInstances) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error())
}
