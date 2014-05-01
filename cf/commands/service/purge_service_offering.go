package service

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/flag_helpers"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type PurgeServiceOffering struct {
	ui          terminal.UI
	serviceRepo api.ServiceRepository
}

func (cmd PurgeServiceOffering) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		err = errors.New("incorrect usage")
		cmd.ui.FailWithUsage(c, "purge-service-offering")
	}

	reqs = []requirements.Requirement{requirementsFactory.NewLoginRequirement()}
	return
}

func (command PurgeServiceOffering) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "purge-service-offering",
		Description: "Recursively remove a service and child objects from Cloud Foundry database without making requests to a service broker",
		Usage: "CF_NAME purge-service-offering SERVICE [-p PROVIDER]" +
			"\n\nWARNING:\n" +
			"This operation assumes that the service broker responsible for this service offering is no longer available, and all service instances have been deleted, leaving orphan records in Cloud Foundry's database. All knowledge of the service will be removed from Cloud Foundry, including service instances and service bindings. No attempt will be made to contact the service broker; running this command without destroying the service broker will cause orphan service instances. After running this command you may want to run either delete-service-auth-token or delete-service-broker to complete the cleanup.",
		Flags: []cli.Flag{
			flag_helpers.NewStringFlag("p", "Provider"),
			cli.BoolFlag{Name: "f", Usage: "Force deletion without confirmation"},
		},
	}
}

func (cmd PurgeServiceOffering) Run(c *cli.Context) {
	serviceName := c.Args()[0]

	offering, apiErr := cmd.serviceRepo.FindServiceOfferingByLabelAndProvider(serviceName, c.String("p"))

	switch apiErr.(type) {
	case nil:
	case *errors.ModelNotFoundError:
		cmd.ui.Warn("Service offering does not exist\nTIP: If you are trying to purge a v1 service offering, you must set the -p flag.")
		return
	default:
		cmd.ui.Failed(apiErr.Error())
	}

	confirmed := c.Bool("f")
	if !confirmed {
		cmd.ui.Warn(`Warning: This operation assumes that the service broker responsible for this service offering is no longer available, and all service instances have been deleted, leaving orphan records in Cloud Foundry's database. All knowledge of the service will be removed from Cloud Foundry, including service instances and service bindings. No attempt will be made to contact the service broker; running this command without destroying the service broker will cause orphan service instances. After running this command you may want to run either delete-service-auth-token or delete-service-broker to complete the cleanup.`)
		confirmed = cmd.ui.Confirm("Really purge service offering %s from Cloud Foundry?", serviceName)
	}

	if !confirmed {
		return
	}
	cmd.ui.Say("Purging service %s...", serviceName)
	cmd.serviceRepo.PurgeServiceOffering(offering)
	cmd.ui.Ok()

	return
}

func NewPurgeServiceOffering(ui terminal.UI, config configuration.Reader, serviceRepo api.ServiceRepository) (cmd PurgeServiceOffering) {
	cmd.ui = ui
	cmd.serviceRepo = serviceRepo
	return
}
