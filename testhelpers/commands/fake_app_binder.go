package commands

import (
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/simonleung8/flags"
)

type FakeAppBinder struct {
	AppsToBind        []models.Application
	InstancesToBindTo []models.ServiceInstance
	Params            map[string]interface{}

	BindApplicationReturns struct {
		Error error
	}
}

func (binder *FakeAppBinder) BindApplication(app models.Application, service models.ServiceInstance, paramsMap map[string]interface{}) error {
	binder.AppsToBind = append(binder.AppsToBind, app)
	binder.InstancesToBindTo = append(binder.InstancesToBindTo, service)
	binder.Params = paramsMap

	return binder.BindApplicationReturns.Error
}

func (cmd *FakeAppBinder) MetaData() command_registry.CommandMetadata {
	return command_registry.CommandMetadata{Name: "bind-service"}
}

func (cmd *FakeAppBinder) SetDependency(_ command_registry.Dependency, _ bool) command_registry.Command {
	return cmd
}

func (cmd *FakeAppBinder) Requirements(_ requirements.Factory, _ flags.FlagContext) (reqs []requirements.Requirement, err error) {
	return
}

func (cmd *FakeAppBinder) Execute(_ flags.FlagContext) {}
