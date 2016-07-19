package servicekey

import (
	"fmt"

	"code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/flags"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"

	. "code.cloudfoundry.org/cli/cf/i18n"
)

type ServiceKeys struct {
	ui                         terminal.UI
	config                     coreconfig.Reader
	serviceRepo                api.ServiceRepository
	serviceKeyRepo             api.ServiceKeyRepository
	serviceInstanceRequirement requirements.ServiceInstanceRequirement
}

func init() {
	commandregistry.Register(&ServiceKeys{})
}

func (cmd *ServiceKeys) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "service-keys",
		ShortName:   "sk",
		Description: T("List keys for a service instance"),
		Usage: []string{
			T("CF_NAME service-keys SERVICE_INSTANCE"),
		},
		Examples: []string{
			"CF_NAME service-keys mydb",
		},
	}
}

func (cmd *ServiceKeys) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + commandregistry.Commands.CommandUsage("service-keys"))
		return nil, fmt.Errorf("Incorrect usage: %d arguments of %d required", len(fc.Args()), 1)
	}

	loginRequirement := requirementsFactory.NewLoginRequirement()
	cmd.serviceInstanceRequirement = requirementsFactory.NewServiceInstanceRequirement(fc.Args()[0])
	targetSpaceRequirement := requirementsFactory.NewTargetedSpaceRequirement()

	reqs := []requirements.Requirement{loginRequirement, cmd.serviceInstanceRequirement, targetSpaceRequirement}

	return reqs, nil
}

func (cmd *ServiceKeys) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.serviceRepo = deps.RepoLocator.GetServiceRepository()
	cmd.serviceKeyRepo = deps.RepoLocator.GetServiceKeyRepository()
	return cmd
}

func (cmd *ServiceKeys) Execute(c flags.FlagContext) error {
	serviceInstance := cmd.serviceInstanceRequirement.GetServiceInstance()

	cmd.ui.Say(T("Getting keys for service instance {{.ServiceInstanceName}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"ServiceInstanceName": terminal.EntityNameColor(serviceInstance.Name),
			"CurrentUser":         terminal.EntityNameColor(cmd.config.Username()),
		}))

	serviceKeys, err := cmd.serviceKeyRepo.ListServiceKeys(serviceInstance.GUID)
	if err != nil {
		return err
	}

	table := cmd.ui.Table([]string{T("name")})

	for _, serviceKey := range serviceKeys {
		table.Add(serviceKey.Fields.Name)
	}

	if len(serviceKeys) == 0 {
		cmd.ui.Say(T("No service key for service instance {{.ServiceInstanceName}}",
			map[string]interface{}{"ServiceInstanceName": terminal.EntityNameColor(serviceInstance.Name)}))
		return nil
	}

	cmd.ui.Say("")
	err = table.Print()
	if err != nil {
		return err
	}
	return nil
}
