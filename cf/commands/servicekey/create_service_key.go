package servicekey

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/flag_helpers"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/json"
	"github.com/codegangsta/cli"

	. "github.com/cloudfoundry/cli/cf/i18n"
)

type CreateServiceKey struct {
	ui                         terminal.UI
	config                     core_config.Reader
	serviceRepo                api.ServiceRepository
	serviceKeyRepo             api.ServiceKeyRepository
	serviceInstanceRequirement requirements.ServiceInstanceRequirement
}

func NewCreateServiceKey(ui terminal.UI, config core_config.Reader, serviceRepo api.ServiceRepository, serviceKeyRepo api.ServiceKeyRepository) (cmd *CreateServiceKey) {
	return &CreateServiceKey{
		ui:             ui,
		config:         config,
		serviceRepo:    serviceRepo,
		serviceKeyRepo: serviceKeyRepo,
	}
}

func (cmd *CreateServiceKey) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "create-service-key",
		ShortName:   "csk",
		Description: T("Create key for a service instance"),
		Usage: T(`CF_NAME create-service-key SERVICE_INSTANCE SERVICE_KEY [-c PARAMETERS_AS_JSON]

  Optionally provide service-specific configuration parameters in a valid JSON object in-line.
  CF_NAME create-service-key SERVICE_INSTANCE SERVICE_KEY -c '{"name":"value","name":"value"}'

  Optionally provide a file containing service-specific configuration parameters in a valid JSON object. The path to the parameters file can be an absolute or relative path to a file.
  CF_NAME create-service-key SERVICE_INSTANCE SERVICE_KEY -c PATH_TO_FILE

   Example of valid JSON object:
   {
     "permissions": "read-only"
   }

EXAMPLE:
   CF_NAME create-service-key mydb mykey -c '{"permissions":"read-only"}'
   CF_NAME create-service-key mydb mykey -c ~/workspace/tmp/instance_config.json`),
		Flags: []cli.Flag{
			flag_helpers.NewStringFlag("c", T("Valid JSON object containing service-specific configuration parameters, provided either in-line or in a file. For a list of supported configuration parameters, see documentation for the particular service offering.")),
		},
	}
}

func (cmd *CreateServiceKey) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 2 {
		cmd.ui.FailWithUsage(c)
	}

	loginRequirement := requirementsFactory.NewLoginRequirement()
	cmd.serviceInstanceRequirement = requirementsFactory.NewServiceInstanceRequirement(c.Args()[0])
	targetSpaceRequirement := requirementsFactory.NewTargetedSpaceRequirement()

	reqs = []requirements.Requirement{loginRequirement, cmd.serviceInstanceRequirement, targetSpaceRequirement}

	return reqs, nil
}

func (cmd *CreateServiceKey) Run(c *cli.Context) {
	serviceInstance := cmd.serviceInstanceRequirement.GetServiceInstance()
	serviceKeyName := c.Args()[1]
	params := c.String("c")

	paramsMap, err := json.ParseJsonFromFileOrString(params)
	if err != nil {
		cmd.ui.Failed(T("Invalid configuration provided for -c flag. Please provide a valid JSON object or path to a file containing a valid JSON object."))
	}

	cmd.ui.Say(T("Creating service key {{.ServiceKeyName}} for service instance {{.ServiceInstanceName}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"ServiceInstanceName": terminal.EntityNameColor(serviceInstance.Name),
			"ServiceKeyName":      terminal.EntityNameColor(serviceKeyName),
			"CurrentUser":         terminal.EntityNameColor(cmd.config.Username()),
		}))

	err = cmd.serviceKeyRepo.CreateServiceKey(serviceInstance.Guid, serviceKeyName, paramsMap)
	switch err.(type) {
	case nil:
		cmd.ui.Ok()
	case *errors.ModelAlreadyExistsError:
		cmd.ui.Ok()
		cmd.ui.Warn(err.Error())
	default:
		cmd.ui.Failed(err.Error())
	}
}
