package service

import (
	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/flags"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type RenameService struct {
	ui                 terminal.UI
	config             coreconfig.Reader
	serviceRepo        api.ServiceRepository
	serviceInstanceReq requirements.ServiceInstanceRequirement
}

func init() {
	commandregistry.Register(&RenameService{})
}

func (cmd *RenameService) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "rename-service",
		Description: T("Rename a service instance"),
		Usage: []string{
			T("CF_NAME rename-service SERVICE_INSTANCE NEW_SERVICE_INSTANCE"),
		},
	}
}

func (cmd *RenameService) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	if len(fc.Args()) != 2 {
		cmd.ui.Failed(T("Incorrect Usage. Requires SERVICE_INSTANCE and NEW_SERVICE_INSTANCE as arguments\n\n") + commandregistry.Commands.CommandUsage("rename-service"))
	}

	cmd.serviceInstanceReq = requirementsFactory.NewServiceInstanceRequirement(fc.Args()[0])

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
		cmd.serviceInstanceReq,
	}

	return reqs
}

func (cmd *RenameService) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.serviceRepo = deps.RepoLocator.GetServiceRepository()
	return cmd
}

func (cmd *RenameService) Execute(c flags.FlagContext) error {
	newName := c.Args()[1]
	serviceInstance := cmd.serviceInstanceReq.GetServiceInstance()

	cmd.ui.Say(T("Renaming service {{.ServiceName}} to {{.NewServiceName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"ServiceName":    terminal.EntityNameColor(serviceInstance.Name),
			"NewServiceName": terminal.EntityNameColor(newName),
			"OrgName":        terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			"SpaceName":      terminal.EntityNameColor(cmd.config.SpaceFields().Name),
			"CurrentUser":    terminal.EntityNameColor(cmd.config.Username()),
		}))
	err := cmd.serviceRepo.RenameService(serviceInstance, newName)

	if err != nil {
		if httpError, ok := err.(errors.HTTPError); ok && httpError.ErrorCode() == errors.ServiceInstanceNameTaken {
			return errors.New(T("{{.ErrorDescription}}\nTIP: Use '{{.CFServicesCommand}}' to view all services in this org and space.",
				map[string]interface{}{
					"ErrorDescription":  httpError.Error(),
					"CFServicesCommand": cf.Name + " " + "services",
				}))
		}
		return err
	}

	cmd.ui.Ok()
	return nil
}
