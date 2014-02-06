package commands

import (
"cf/models"
	"cf/net"
)

type FakeAppBinder struct {
	AppsToBind        []models.Application
	InstancesToBindTo []models.ServiceInstance
}

func (binder *FakeAppBinder) BindApplication(app models.Application, service models.ServiceInstance) (apiResponse net.ApiResponse) {
	binder.AppsToBind = append(binder.AppsToBind, app)
	binder.InstancesToBindTo = append(binder.InstancesToBindTo, service)

	return
}
