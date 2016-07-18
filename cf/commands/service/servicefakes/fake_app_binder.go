package servicefakes

import (
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/flags"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
)

type OldFakeAppBinder struct {
	AppsToBind        []models.Application
	InstancesToBindTo []models.ServiceInstance
	Params            map[string]interface{}

	BindApplicationReturns struct {
		Error error
	}
}

func (binder *OldFakeAppBinder) BindApplication(app models.Application, service models.ServiceInstance, paramsMap map[string]interface{}) error {
	binder.AppsToBind = append(binder.AppsToBind, app)
	binder.InstancesToBindTo = append(binder.InstancesToBindTo, service)
	binder.Params = paramsMap

	return binder.BindApplicationReturns.Error
}

func (binder *OldFakeAppBinder) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{Name: "bind-service"}
}

func (binder *OldFakeAppBinder) SetDependency(_ commandregistry.Dependency, _ bool) commandregistry.Command {
	return binder
}

func (binder *OldFakeAppBinder) Requirements(_ requirements.Factory, _ flags.FlagContext) []requirements.Requirement {
	return []requirements.Requirement{}
}

func (binder *OldFakeAppBinder) Execute(_ flags.FlagContext) error {
	return nil
}
