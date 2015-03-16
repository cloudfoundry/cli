package servicekey

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"

	. "github.com/cloudfoundry/cli/cf/i18n"
)

type CreateServiceKey struct {
	ui             terminal.UI
	config         core_config.Reader
	serviceRepo    api.ServiceRepository
	serviceKeyRepo api.ServiceKeyRepository
}

func NewCreateServiceKey(ui terminal.UI, config core_config.Reader, serviceRepo api.ServiceRepository, serviceKeyRepo api.ServiceKeyRepository) (cmd CreateServiceKey) {
	return CreateServiceKey{
		ui:             ui,
		config:         config,
		serviceRepo:    serviceRepo,
		serviceKeyRepo: serviceKeyRepo,
	}
}

func (cmd CreateServiceKey) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "create-service-key",
		ShortName:   "csk",
		Description: T("Create key for a service instance"),
		Usage: T(`CF_NAME create-service-key SERVICE_INSTANCE SERVICE_KEY

EXAMPLE:
   CF_NAME create-service-key mydb mykey`),
	}
}

func (cmd CreateServiceKey) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 2 {
		cmd.ui.FailWithUsage(c)
	}

	loginRequirement := requirementsFactory.NewLoginRequirement()
	serviceInstanceRequirement := requirementsFactory.NewServiceInstanceRequirement(c.Args()[0])
	targetSpaceRequirement := requirementsFactory.NewTargetedSpaceRequirement()

	reqs = []requirements.Requirement{loginRequirement, serviceInstanceRequirement, targetSpaceRequirement}

	return reqs, nil
}

func (cmd CreateServiceKey) Run(c *cli.Context) {
	serviceInstanceName := c.Args()[0]
	serviceKeyName := c.Args()[1]

	cmd.ui.Say(T("Creating service key {{.ServiceKeyName}} for service instance {{.ServiceInstanceName}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"ServiceInstanceName": terminal.EntityNameColor(serviceInstanceName),
			"ServiceKeyName":      terminal.EntityNameColor(serviceKeyName),
			"CurrentUser":         terminal.EntityNameColor(cmd.config.Username()),
		}))

	serviceInstance, err := cmd.serviceRepo.FindInstanceByName(serviceInstanceName)
	if err != nil {
		cmd.ui.Failed(err.Error())
		return
	}

	err = cmd.serviceKeyRepo.CreateServiceKey(serviceInstance.Guid, serviceKeyName)
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
