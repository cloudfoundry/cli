package servicefakes

import (
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/flags"
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

func (binder *OldFakeAppBinder) MetaData() command_registry.CommandMetadata {
	return command_registry.CommandMetadata{Name: "bind-service"}
}

func (binder *OldFakeAppBinder) SetDependency(_ command_registry.Dependency, _ bool) command_registry.Command {
	return binder
}

func (binder *OldFakeAppBinder) Requirements(_ requirements.Factory, _ flags.FlagContext) []requirements.Requirement {
	return []requirements.Requirement{}
}

func (binder *OldFakeAppBinder) Execute(_ flags.FlagContext) {}
