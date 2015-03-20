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

type DeleteServiceKey struct {
	ui             terminal.UI
	config         core_config.Reader
	serviceRepo    api.ServiceRepository
	serviceKeyRepo api.ServiceKeyRepository
}

func NewDeleteServiceKey(ui terminal.UI, config core_config.Reader, serviceRepo api.ServiceRepository, serviceKeyRepo api.ServiceKeyRepository) (cmd DeleteServiceKey) {
	return DeleteServiceKey{
		ui:             ui,
		config:         config,
		serviceRepo:    serviceRepo,
		serviceKeyRepo: serviceKeyRepo,
	}
}

func (cmd DeleteServiceKey) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "delete-service-key",
		ShortName:   "dsk",
		Description: T("Delete a service key"),
		Usage: T(`CF_NAME delete-service-key SERVICE_INSTANCE SERVICE_KEY [-f]

EXAMPLE:
   CF_NAME delete-service-key mydb mykey`),
		Flags: []cli.Flag{
			cli.BoolFlag{Name: "f", Usage: T("Force deletion without confirmation")},
		},
	}
}

func (cmd DeleteServiceKey) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 2 {
		cmd.ui.FailWithUsage(c)
	}

	loginRequirement := requirementsFactory.NewLoginRequirement()
	serviceInstanceRequirement := requirementsFactory.NewServiceInstanceRequirement(c.Args()[0])
	targetSpaceRequirement := requirementsFactory.NewTargetedSpaceRequirement()

	reqs = []requirements.Requirement{loginRequirement, serviceInstanceRequirement, targetSpaceRequirement}

	return reqs, nil
}

func (cmd DeleteServiceKey) Run(c *cli.Context) {
	serviceInstanceName := c.Args()[0]
	serviceKeyName := c.Args()[1]

	if !c.Bool("f") {
		if !cmd.ui.ConfirmDelete(T("service key"), serviceKeyName) {
			return
		}
	}

	cmd.ui.Say(T("Deleting key {{.ServiceKeyName}} for service instance {{.ServiceInstanceName}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"ServiceKeyName":      terminal.EntityNameColor(serviceKeyName),
			"ServiceInstanceName": terminal.EntityNameColor(serviceInstanceName),
			"CurrentUser":         terminal.EntityNameColor(cmd.config.Username()),
		}))

	serviceInstance, err := cmd.serviceRepo.FindInstanceByName(serviceInstanceName)
	if err != nil {
		cmd.ui.Ok()

		cmd.ui.Say(T("Service instance {{.ServiceInstanceName}} does not exist.",
			map[string]interface{}{
				"ServiceInstanceName": terminal.EntityNameColor(serviceInstanceName),
			}))

		return
	}

	serviceKey, err := cmd.serviceKeyRepo.GetServiceKey(serviceInstance.Guid, serviceKeyName)
	if err != nil {
		cmd.ui.Ok()

		cmd.ui.Say(T("Service key {{.ServiceKeyName}} does not exist for service instance {{.ServiceInstanceName}}.",
			map[string]interface{}{
				"ServiceKeyName":      terminal.EntityNameColor(serviceKeyName),
				"ServiceInstanceName": terminal.EntityNameColor(serviceInstanceName),
			}))

		return
	}

	err = cmd.serviceKeyRepo.DeleteServiceKey(serviceKey.Fields.Guid)
	if err != nil {
		cmd.ui.Failed(err.Error())
		return
	}

	cmd.ui.Ok()
}
