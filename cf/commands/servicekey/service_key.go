package servicekey

import (
	"encoding/json"

	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/flags"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"

	. "github.com/cloudfoundry/cli/cf/i18n"
)

type ServiceKey struct {
	ui                         terminal.UI
	config                     coreconfig.Reader
	serviceRepo                api.ServiceRepository
	serviceKeyRepo             api.ServiceKeyRepository
	serviceInstanceRequirement requirements.ServiceInstanceRequirement
}

func init() {
	commandregistry.Register(&ServiceKey{})
}

func (cmd *ServiceKey) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["guid"] = &flags.BoolFlag{Name: "guid", Usage: T("Retrieve and display the given service-key's guid.  All other output for the service is suppressed.")}

	return commandregistry.CommandMetadata{
		Name:        "service-key",
		Description: T("Show service key info"),
		Usage: []string{
			T("CF_NAME service-key SERVICE_INSTANCE SERVICE_KEY"),
		},
		Examples: []string{
			"CF_NAME service-key mydb mykey",
		},
		Flags: fs,
	}
}

func (cmd *ServiceKey) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	if len(fc.Args()) != 2 {
		cmd.ui.Failed(T("Incorrect Usage. Requires SERVICE_INSTANCE SERVICE_KEY as arguments\n\n") + commandregistry.Commands.CommandUsage("service-key"))
	}

	loginRequirement := requirementsFactory.NewLoginRequirement()
	cmd.serviceInstanceRequirement = requirementsFactory.NewServiceInstanceRequirement(fc.Args()[0])
	targetSpaceRequirement := requirementsFactory.NewTargetedSpaceRequirement()

	reqs := []requirements.Requirement{loginRequirement, cmd.serviceInstanceRequirement, targetSpaceRequirement}
	return reqs
}

func (cmd *ServiceKey) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.serviceRepo = deps.RepoLocator.GetServiceRepository()
	cmd.serviceKeyRepo = deps.RepoLocator.GetServiceKeyRepository()
	return cmd
}

func (cmd *ServiceKey) Execute(c flags.FlagContext) error {
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

	serviceKey, err := cmd.serviceKeyRepo.GetServiceKey(serviceInstance.GUID, serviceKeyName)
	if err != nil {
		switch err.(type) {
		case *errors.NotAuthorizedError:
			cmd.ui.Say(T("No service key {{.ServiceKeyName}} found for service instance {{.ServiceInstanceName}}",
				map[string]interface{}{
					"ServiceKeyName":      terminal.EntityNameColor(serviceKeyName),
					"ServiceInstanceName": terminal.EntityNameColor(serviceInstance.Name)}))
			return nil
		default:
			return err
		}
	}

	if c.Bool("guid") {
		cmd.ui.Say(serviceKey.Fields.GUID)
	} else {
		if serviceKey.Fields.Name == "" {
			cmd.ui.Say(T("No service key {{.ServiceKeyName}} found for service instance {{.ServiceInstanceName}}",
				map[string]interface{}{
					"ServiceKeyName":      terminal.EntityNameColor(serviceKeyName),
					"ServiceInstanceName": terminal.EntityNameColor(serviceInstance.Name)}))
			return nil
		}

		jsonBytes, err := json.MarshalIndent(serviceKey.Credentials, "", " ")
		if err != nil {
			return err
		}

		cmd.ui.Say("")
		cmd.ui.Say(string(jsonBytes))
	}
	return nil
}
