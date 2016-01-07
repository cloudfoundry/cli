package servicekey

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"

	. "github.com/cloudfoundry/cli/cf/i18n"
)

type ServiceKeys struct {
	ui                         terminal.UI
	config                     core_config.Reader
	serviceRepo                api.ServiceRepository
	serviceKeyRepo             api.ServiceKeyRepository
	serviceInstanceRequirement requirements.ServiceInstanceRequirement
}

func init() {
	command_registry.Register(&ServiceKeys{})
}

func (cmd *ServiceKeys) MetaData() command_registry.CommandMetadata {
	return command_registry.CommandMetadata{
		Name:        "service-keys",
		ShortName:   "sk",
		Description: T("List keys for a service instance"),
		Usage: T(`CF_NAME service-keys SERVICE_INSTANCE

EXAMPLE:
   CF_NAME service-keys mydb`),
	}
}

func (cmd *ServiceKeys) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + command_registry.Commands.CommandUsage("service-keys"))
	}

	loginRequirement := requirementsFactory.NewLoginRequirement()
	cmd.serviceInstanceRequirement = requirementsFactory.NewServiceInstanceRequirement(fc.Args()[0])
	targetSpaceRequirement := requirementsFactory.NewTargetedSpaceRequirement()

	reqs := []requirements.Requirement{loginRequirement, cmd.serviceInstanceRequirement, targetSpaceRequirement}

	return reqs, nil
}

func (cmd *ServiceKeys) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.serviceRepo = deps.RepoLocator.GetServiceRepository()
	cmd.serviceKeyRepo = deps.RepoLocator.GetServiceKeyRepository()
	return cmd
}

func (cmd *ServiceKeys) Execute(c flags.FlagContext) {
	serviceInstance := cmd.serviceInstanceRequirement.GetServiceInstance()

	cmd.ui.Say(T("Getting keys for service instance {{.ServiceInstanceName}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"ServiceInstanceName": terminal.EntityNameColor(serviceInstance.Name),
			"CurrentUser":         terminal.EntityNameColor(cmd.config.Username()),
		}))

	serviceKeys, err := cmd.serviceKeyRepo.ListServiceKeys(serviceInstance.Guid)
	if err != nil {
		cmd.ui.Failed(err.Error())
		return
	}

	table := cmd.ui.Table([]string{T("name")})

	for _, serviceKey := range serviceKeys {
		table.Add(serviceKey.Fields.Name)
	}

	if len(serviceKeys) == 0 {
		cmd.ui.Say(T("No service key for service instance {{.ServiceInstanceName}}",
			map[string]interface{}{"ServiceInstanceName": terminal.EntityNameColor(serviceInstance.Name)}))
		return
	}

	cmd.ui.Say("")
	table.Print()
}
