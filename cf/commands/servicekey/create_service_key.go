package servicekey

import (
	"fmt"

	"code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/flags"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
	"code.cloudfoundry.org/cli/util/json"

	. "code.cloudfoundry.org/cli/cf/i18n"
)

type CreateServiceKey struct {
	ui                         terminal.UI
	config                     coreconfig.Reader
	serviceRepo                api.ServiceRepository
	serviceKeyRepo             api.ServiceKeyRepository
	serviceInstanceRequirement requirements.ServiceInstanceRequirement
}

func init() {
	commandregistry.Register(&CreateServiceKey{})
}

func (cmd *CreateServiceKey) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["c"] = &flags.StringFlag{ShortName: "c", Usage: T("Valid JSON object containing service-specific configuration parameters, provided either in-line or in a file. For a list of supported configuration parameters, see documentation for the particular service offering.")}

	return commandregistry.CommandMetadata{
		Name:        "create-service-key",
		ShortName:   "csk",
		Description: T("Create key for a service instance"),
		Usage: []string{
			T(`CF_NAME create-service-key SERVICE_INSTANCE SERVICE_KEY [-c PARAMETERS_AS_JSON]

   Optionally provide service-specific configuration parameters in a valid JSON object in-line.
   CF_NAME create-service-key SERVICE_INSTANCE SERVICE_KEY -c '{"name":"value","name":"value"}'

   Optionally provide a file containing service-specific configuration parameters in a valid JSON object. The path to the parameters file can be an absolute or relative path to a file.
   CF_NAME create-service-key SERVICE_INSTANCE SERVICE_KEY -c PATH_TO_FILE

   Example of valid JSON object:
   {
     "permissions": "read-only"
   }`),
		},
		Examples: []string{
			`CF_NAME create-service-key mydb mykey -c '{"permissions":"read-only"}'`,
			`CF_NAME create-service-key mydb mykey -c ~/workspace/tmp/instance_config.json`,
		},
		Flags: fs,
	}
}

func (cmd *CreateServiceKey) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 2 {
		cmd.ui.Failed(T("Incorrect Usage. Requires SERVICE_INSTANCE and SERVICE_KEY as arguments\n\n") + commandregistry.Commands.CommandUsage("create-service-key"))
		return nil, fmt.Errorf("Incorrect usage: %d arguments of %d required", len(fc.Args()), 2)
	}

	loginRequirement := requirementsFactory.NewLoginRequirement()
	cmd.serviceInstanceRequirement = requirementsFactory.NewServiceInstanceRequirement(fc.Args()[0])
	targetSpaceRequirement := requirementsFactory.NewTargetedSpaceRequirement()

	reqs := []requirements.Requirement{
		loginRequirement,
		cmd.serviceInstanceRequirement,
		targetSpaceRequirement,
	}

	return reqs, nil
}

func (cmd *CreateServiceKey) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.serviceRepo = deps.RepoLocator.GetServiceRepository()
	cmd.serviceKeyRepo = deps.RepoLocator.GetServiceKeyRepository()
	return cmd
}

func (cmd *CreateServiceKey) Execute(c flags.FlagContext) error {
	serviceInstance := cmd.serviceInstanceRequirement.GetServiceInstance()
	serviceKeyName := c.Args()[1]
	params := c.String("c")

	paramsMap, err := json.ParseJSONFromFileOrString(params)
	if err != nil {
		return errors.New(T("Invalid configuration provided for -c flag. Please provide a valid JSON object or path to a file containing a valid JSON object."))
	}

	cmd.ui.Say(T("Creating service key {{.ServiceKeyName}} for service instance {{.ServiceInstanceName}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"ServiceInstanceName": terminal.EntityNameColor(serviceInstance.Name),
			"ServiceKeyName":      terminal.EntityNameColor(serviceKeyName),
			"CurrentUser":         terminal.EntityNameColor(cmd.config.Username()),
		}))

	err = cmd.serviceKeyRepo.CreateServiceKey(serviceInstance.GUID, serviceKeyName, paramsMap)
	switch err.(type) {
	case nil:
		cmd.ui.Ok()
	case *errors.ModelAlreadyExistsError:
		cmd.ui.Ok()
		cmd.ui.Warn(err.Error())
	default:
		return err
	}
	return nil
}
