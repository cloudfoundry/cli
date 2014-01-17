package commands

import (
	"cf"
	"cf/net"
)

type FakeAppBinder struct {
	AppsToBind        cf.ApplicationSet
	InstancesToBindTo cf.ServiceInstanceSet
}

func (binder *FakeAppBinder) BindApplication(app cf.Application, service cf.ServiceInstance) (apiResponse net.ApiResponse) {
	binder.AppsToBind = append(binder.AppsToBind, app)
	binder.InstancesToBindTo = append(binder.InstancesToBindTo, service)

	return
}
