package servicekey

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/flags"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"

	. "github.com/cloudfoundry/cli/cf/i18n"
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

func (cmd *ServiceKeys) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + commandregistry.Commands.CommandUsage("service-keys"))
	}

	loginRequirement := requirementsFactory.NewLoginRequirement()
	cmd.serviceInstanceRequirement = requirementsFactory.NewServiceInstanceRequirement(fc.Args()[0])
	targetSpaceRequirement := requirementsFactory.NewTargetedSpaceRequirement()

	reqs := []requirements.Requirement{loginRequirement, cmd.serviceInstanceRequirement, targetSpaceRequirement}

	return reqs
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
	table.Print()
	return nil
}
