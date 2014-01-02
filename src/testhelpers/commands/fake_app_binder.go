package commands

import (
	"cf"
	"cf/net"
)

type FakeAppBinder struct {
	AppToBind cf.Application
	InstanceToBindTo cf.ServiceInstance
}

func (binder *FakeAppBinder) BindApplication(app cf.Application, service cf.ServiceInstance) (apiResponse net.ApiResponse) {
	binder.AppToBind = app
	binder.InstanceToBindTo = service

	return
}
