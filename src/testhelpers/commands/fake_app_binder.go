package commands

import (
	"cf/errors"
	"cf/models"
)

type FakeAppBinder struct {
	AppsToBind        []models.Application
	InstancesToBindTo []models.ServiceInstance
}

func (binder *FakeAppBinder) BindApplication(app models.Application, service models.ServiceInstance) (apiResponse errors.Error) {
	binder.AppsToBind = append(binder.AppsToBind, app)
	binder.InstancesToBindTo = append(binder.InstancesToBindTo, service)

	return
}
