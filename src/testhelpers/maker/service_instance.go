package maker

import "cf"

var serviceInstanceGuid func() string

func init() {
	serviceInstanceGuid = guidGenerator("services")
}

func NewServiceInstance(name string) (service cf.ServiceInstance) {
	service.Name = name
	service.Guid = serviceInstanceGuid()
	return
}
