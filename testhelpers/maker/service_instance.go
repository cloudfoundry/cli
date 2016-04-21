package maker

import "github.com/cloudfoundry/cli/cf/models"

var serviceInstanceGUID func() string

func init() {
	serviceInstanceGUID = guidGenerator("services")
}

func NewServiceInstance(name string) (service models.ServiceInstance) {
	return models.ServiceInstance{ServiceInstanceFields: models.ServiceInstanceFields{
		Name: name,
		GUID: serviceInstanceGUID(),
	}}
}
