package testhelpers

import (
	"cf/requirements"
	"cf"
)

type FakeReqFactory struct {
	ApplicationName string
	Application     cf.Application

	ServiceInstanceName string
	ServiceInstance     cf.ServiceInstance

	LoginSuccess bool
	ValidAccessTokenSuccess bool
	TargetedSpaceSuccess bool
	TargetedOrgSuccess bool
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

func (f *FakeReqFactory) NewValidAccessTokenRequirement() requirements.Requirement {
	return FakeRequirement{ f, f.ValidAccessTokenSuccess }
}

func (f *FakeReqFactory) NewTargetedSpaceRequirement() requirements.Requirement {
	return FakeRequirement{ f, f.TargetedSpaceSuccess }
}

func (f *FakeReqFactory) NewTargetedOrgRequirement() requirements.Requirement {
	return FakeRequirement{ f, f.TargetedOrgSuccess }
}

type FakeRequirement struct {
	factory *FakeReqFactory
	success bool
}

func (r FakeRequirement) Execute() (success bool) {
	return r.success
}

func (r FakeRequirement) GetApplication() cf.Application {
	return r.factory.Application
}

func (r FakeRequirement) GetServiceInstance() cf.ServiceInstance {
	return r.factory.ServiceInstance
}
