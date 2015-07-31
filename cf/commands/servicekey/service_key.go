package servicekey

import (
	"encoding/json"

	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
	"github.com/cloudfoundry/cli/flags/flag"

	. "github.com/cloudfoundry/cli/cf/i18n"
)

type ServiceKey struct {
	ui                         terminal.UI
	config                     core_config.Reader
	serviceRepo                api.ServiceRepository
	serviceKeyRepo             api.ServiceKeyRepository
	serviceInstanceRequirement requirements.ServiceInstanceRequirement
}

func init() {
	command_registry.Register(&ServiceKey{})
}

func (cmd *ServiceKey) MetaData() command_registry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["guid"] = &cliFlags.BoolFlag{Name: "guid", Usage: T("Retrieve and display the given service-key's guid.  All other output for the service is suppressed.")}

	return command_registry.CommandMetadata{
		Name:        "service-key",
		Description: T("Show service key info"),
		Usage: T(`CF_NAME service-key SERVICE_INSTANCE SERVICE_KEY

EXAMPLE:
   CF_NAME service-key mydb mykey`),
		Flags: fs,
	}
}

func (cmd *ServiceKey) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 2 {
		cmd.ui.Failed(T("Incorrect Usage. Requires SERVICE_INSTANCE SERVICE_KEY as arguments\n\n") + command_registry.Commands.CommandUsage("service-key"))
	}

	loginRequirement := requirementsFactory.NewLoginRequirement()
	cmd.serviceInstanceRequirement = requirementsFactory.NewServiceInstanceRequirement(fc.Args()[0])
	targetSpaceRequirement := requirementsFactory.NewTargetedSpaceRequirement()

	reqs := []requirements.Requirement{loginRequirement, cmd.serviceInstanceRequirement, targetSpaceRequirement}
	return reqs, nil
}

func (cmd *ServiceKey) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.serviceRepo = deps.RepoLocator.GetServiceRepository()
	cmd.serviceKeyRepo = deps.RepoLocator.GetServiceKeyRepository()
	return cmd
}

func (cmd *ServiceKey) Execute(c flags.FlagContext) {
	serviceInstance := cmd.serviceInstanceRequirement.GetServiceInstance()
	serviceKeyName := c.Args()[1]

	if !c.Bool("guid") {
		cmd.ui.Say(T("Getting key {{.ServiceKeyName}} for service instance {{.ServiceInstanceName}} as {{.CurrentUser}}...",
			map[string]interface{}{
				"ServiceKeyName":      terminal.EntityNameColor(serviceKeyName),
				"ServiceInstanceName": terminal.EntityNameColor(serviceInstance.Name),
				"CurrentUser":         terminal.EntityNameColor(cmd.config.Username()),
			}))
	}

	serviceKey, err := cmd.serviceKeyRepo.GetServiceKey(serviceInstance.Guid, serviceKeyName)
	if err != nil {
		switch err.(type) {
		case *errors.NotAuthorizedError:
			cmd.ui.Say(T("No service key {{.ServiceKeyName}} found for service instance {{.ServiceInstanceName}}",
				map[string]interface{}{
					"ServiceKeyName":      terminal.EntityNameColor(serviceKeyName),
					"ServiceInstanceName": terminal.EntityNameColor(serviceInstance.Name)}))
			return
		default:
			cmd.ui.Failed(err.Error())
			return
		}
	}

	if c.Bool("guid") {
		cmd.ui.Say(serviceKey.Fields.Guid)
	} else {
		if serviceKey.Fields.Name == "" {
			cmd.ui.Say(T("No service key {{.ServiceKeyName}} found for service instance {{.ServiceInstanceName}}",
				map[string]interface{}{
					"ServiceKeyName":      terminal.EntityNameColor(serviceKeyName),
					"ServiceInstanceName": terminal.EntityNameColor(serviceInstance.Name)}))
			return
		}

		jsonBytes, err := json.MarshalIndent(serviceKey.Credentials, "", " ")
		if err != nil {
			cmd.ui.Failed(err.Error())
			return
		}

		cmd.ui.Say("")
		cmd.ui.Say(string(jsonBytes))
	}
}
