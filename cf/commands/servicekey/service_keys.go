package servicekey

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"

	. "github.com/cloudfoundry/cli/cf/i18n"
)

type ServiceKeys struct {
	ui             terminal.UI
	config         core_config.Reader
	serviceRepo    api.ServiceRepository
	serviceKeyRepo api.ServiceKeyRepository
}

func NewListServiceKeys(ui terminal.UI, config core_config.Reader, serviceRepo api.ServiceRepository, serviceKeyRepo api.ServiceKeyRepository) (cmd ServiceKeys) {
	return ServiceKeys{
		ui:             ui,
		config:         config,
		serviceRepo:    serviceRepo,
		serviceKeyRepo: serviceKeyRepo,
	}
}

func (cmd ServiceKeys) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "service-keys",
		ShortName:   "sk",
		Description: T("List keys for a service instance"),
		Usage: T(`CF_NAME service-keys SERVICE_INSTANCE

EXAMPLE:
   CF_NAME service-keys mydb`),
	}
}

func (cmd ServiceKeys) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		cmd.ui.FailWithUsage(c)
	}

	loginRequirement := requirementsFactory.NewLoginRequirement()
	serviceInstanceRequirement := requirementsFactory.NewServiceInstanceRequirement(c.Args()[0])
	targetSpaceRequirement := requirementsFactory.NewTargetedSpaceRequirement()

	reqs = []requirements.Requirement{loginRequirement, serviceInstanceRequirement, targetSpaceRequirement}

	return reqs, nil
}

func (cmd ServiceKeys) Run(c *cli.Context) {
	serviceInstanceName := c.Args()[0]

	cmd.ui.Say(T("Getting keys for service instance {{.ServiceInstanceName}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"ServiceInstanceName": terminal.EntityNameColor(serviceInstanceName),
			"CurrentUser":         terminal.EntityNameColor(cmd.config.Username()),
		}))

	serviceInstance, err := cmd.serviceRepo.FindInstanceByName(serviceInstanceName)
	if err != nil {
		cmd.ui.Failed(err.Error())
		return
	}

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
			map[string]interface{}{"ServiceInstanceName": terminal.EntityNameColor(serviceInstanceName)}))
		return
	} else {
		table.Print()
	}
}
