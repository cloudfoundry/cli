package maker

import "github.com/cloudfoundry/cli/cf/models"

var serviceInstanceGuid func() string

func init() {
	serviceInstanceGuid = guidGenerator("services")
}

func NewServiceInstance(name string) (service models.ServiceInstance) {
	return models.ServiceInstance{ServiceInstanceFields: models.ServiceInstanceFields{
		Name: name,
		Guid: serviceInstanceGuid(),
	}}
}
