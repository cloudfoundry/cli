package testhelpers

import (
	"cf/requirements"
	"cf"
	"errors"
)

type FakeReqFactory struct {
	ApplicationName string
	Application     cf.Application

	ServiceInstanceName string
	ServiceInstance     cf.ServiceInstance

	LoginSuccess bool
	SpaceSuccess bool
}

func (f *FakeReqFactory) NewApplicationRequirement(name string) requirements.ApplicationRequirement {
	f.ApplicationName = name
	return FakeRequirement{f, true}
}

func (f *FakeReqFactory) NewServiceInstanceRequirement(name string) requirements.ServiceInstanceRequirement {
	f.ServiceInstanceName = name
	return FakeRequirement{f, true}
}

func (f *FakeReqFactory) NewLoginRequirement() requirements.Requirement {
	return FakeRequirement{ f, f.LoginSuccess }
}

func (f *FakeReqFactory) NewSpaceRequirement() requirements.Requirement {
	return FakeRequirement{ f, f.SpaceSuccess }
}

type FakeRequirement struct {
	factory *FakeReqFactory
	success bool
}

func (r FakeRequirement) Execute() (err error) {
	if !r.success {
		err = errors.New("Requirement error")
	}
	return
}

func (r FakeRequirement) GetApplication() cf.Application {
	return r.factory.Application
}

func (r FakeRequirement) GetServiceInstance() cf.ServiceInstance {
	return r.factory.ServiceInstance
}
