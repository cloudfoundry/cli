package testhelpers

import (
	"cf/requirements"
	"cf"
)

type FakeReqFactory struct {
	ApplicationName string
	Application cf.Application

	ServiceInstanceName string
	ServiceInstance cf.ServiceInstance
}

func (f *FakeReqFactory) NewApplicationRequirement(name string) requirements.ApplicationRequirement {
	f.ApplicationName = name
	return requirements.ApplicationRequirement{Application: f.Application}
}

func (f *FakeReqFactory) NewServiceInstanceRequirement(name string) requirements.ServiceInstanceRequirement {
	f.ServiceInstanceName = name
	return requirements.ServiceInstanceRequirement{ServiceInstance: f.ServiceInstance}
}
