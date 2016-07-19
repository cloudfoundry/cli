package service

import (
	"fmt"

	"code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/flags"
	. "code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
)

type UnbindService struct {
	ui                 terminal.UI
	config             coreconfig.Reader
	serviceBindingRepo api.ServiceBindingRepository
	appReq             requirements.ApplicationRequirement
	serviceInstanceReq requirements.ServiceInstanceRequirement
}

func init() {
	commandregistry.Register(&UnbindService{})
}

func (cmd *UnbindService) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "unbind-service",
		ShortName:   "us",
		Description: T("Unbind a service instance from an app"),
		Usage: []string{
			T("CF_NAME unbind-service APP_NAME SERVICE_INSTANCE"),
		},
	}
}

func (cmd *UnbindService) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 2 {
		cmd.ui.Failed(T("Incorrect Usage. Requires APP SERVICE_INSTANCE as arguments\n\n") + commandregistry.Commands.CommandUsage("unbind-service"))
		return nil, fmt.Errorf("Incorrect usage: %d arguments of %d required", len(fc.Args()), 2)
	}

	serviceName := fc.Args()[1]

	cmd.appReq = requirementsFactory.NewApplicationRequirement(fc.Args()[0])
	cmd.serviceInstanceReq = requirementsFactory.NewServiceInstanceRequirement(serviceName)

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		cmd.appReq,
		cmd.serviceInstanceReq,
	}
	return reqs, nil
}

func (cmd *UnbindService) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.serviceBindingRepo = deps.RepoLocator.GetServiceBindingRepository()
	return cmd
}

func (cmd *UnbindService) Execute(c flags.FlagContext) error {
	app := cmd.appReq.GetApplication()
	instance := cmd.serviceInstanceReq.GetServiceInstance()

	cmd.ui.Say(T("Unbinding app {{.AppName}} from service {{.ServiceName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"AppName":     terminal.EntityNameColor(app.Name),
			"ServiceName": terminal.EntityNameColor(instance.Name),
			"OrgName":     terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			"SpaceName":   terminal.EntityNameColor(cmd.config.SpaceFields().Name),
			"CurrentUser": terminal.EntityNameColor(cmd.config.Username()),
		}))

	found, err := cmd.serviceBindingRepo.Delete(instance, app.GUID)
	if err != nil {
		return err
	}

	cmd.ui.Ok()

	if !found {
		cmd.ui.Warn(T("Binding between {{.InstanceName}} and {{.AppName}} did not exist",
			map[string]interface{}{"InstanceName": instance.Name, "AppName": app.Name}))
	}
	return nil
}
