package commands

import "github.com/cloudfoundry/cli/cf/models"

type FakeAppBinder struct {
	AppsToBind        []models.Application
	InstancesToBindTo []models.ServiceInstance

	BindApplicationReturns struct {
		Error error
	}
}

func (binder *FakeAppBinder) BindApplication(app models.Application, service models.ServiceInstance) error {
	binder.AppsToBind = append(binder.AppsToBind, app)
	binder.InstancesToBindTo = append(binder.InstancesToBindTo, service)

	return binder.BindApplicationReturns.Error
}
