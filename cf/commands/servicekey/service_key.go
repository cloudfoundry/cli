package servicekey

import (
	"encoding/json"

	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"

	. "github.com/cloudfoundry/cli/cf/i18n"
)

type ServiceKey struct {
	ui             terminal.UI
	config         core_config.Reader
	serviceRepo    api.ServiceRepository
	serviceKeyRepo api.ServiceKeyRepository
}

func NewGetServiceKey(ui terminal.UI, config core_config.Reader, serviceRepo api.ServiceRepository, serviceKeyRepo api.ServiceKeyRepository) (cmd ServiceKey) {
	return ServiceKey{
		ui:             ui,
		config:         config,
		serviceRepo:    serviceRepo,
		serviceKeyRepo: serviceKeyRepo,
	}
}

func (cmd ServiceKey) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "service-key",
		Description: T("Show service key info"),
		Usage: T(`CF_NAME service-key SERVICE_INSTANCE SERVICE_KEY

EXAMPLE:
   CF_NAME service-key mydb mykey`),
		Flags: []cli.Flag{
			cli.BoolFlag{Name: "guid", Usage: T("Retrieve and display the given service-key's guid.  All other output for the service is suppressed.")},
		},
	}
}

func (cmd ServiceKey) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 2 {
		cmd.ui.FailWithUsage(c)
	}

	loginRequirement := requirementsFactory.NewLoginRequirement()
	serviceInstanceRequirement := requirementsFactory.NewServiceInstanceRequirement(c.Args()[0])
	targetSpaceRequirement := requirementsFactory.NewTargetedSpaceRequirement()

	reqs = []requirements.Requirement{loginRequirement, serviceInstanceRequirement, targetSpaceRequirement}

	return reqs, nil
}

func (cmd ServiceKey) Run(c *cli.Context) {
	serviceInstanceName := c.Args()[0]
	serviceKeyName := c.Args()[1]

	serviceInstance, err := cmd.serviceRepo.FindInstanceByName(serviceInstanceName)
	if err != nil {
		cmd.ui.Failed(err.Error())
		return
	}

	serviceKey, err := cmd.serviceKeyRepo.GetServiceKey(serviceInstance.Guid, serviceKeyName)
	if err != nil {
		cmd.ui.Failed(err.Error())
		return
	}

	if c.Bool("guid") {
		cmd.ui.Say(serviceKey.Fields.Guid)
	} else {
		cmd.ui.Say(T("Getting key {{.ServiceKeyName}} for service instance {{.ServiceInstanceName}} as {{.CurrentUser}}...",
			map[string]interface{}{
				"ServiceKeyName":      terminal.EntityNameColor(serviceKeyName),
				"ServiceInstanceName": terminal.EntityNameColor(serviceInstanceName),
				"CurrentUser":         terminal.EntityNameColor(cmd.config.Username()),
			}))

		if serviceKey.Fields.Name == "" {
			cmd.ui.Say(T("No service key {{.ServiceKeyName}} found for service instance {{.ServiceInstanceName}}",
				map[string]interface{}{
					"ServiceKeyName":      terminal.EntityNameColor(serviceKeyName),
					"ServiceInstanceName": terminal.EntityNameColor(serviceInstanceName)}))
			return
		}

		jsonBytes, err := json.MarshalIndent(serviceKey.Credentials, "", " ")
		if err != nil {
			cmd.ui.Failed(err.Error())
			return
		}

		cmd.ui.Say(string(jsonBytes))
		cmd.ui.Ok()
	}
}
