package service

import (
	"fmt"

	"code.cloudfoundry.org/cli/cf"
	"code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/flags"
	. "code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
)

type PurgeServiceOffering struct {
	ui          terminal.UI
	serviceRepo api.ServiceRepository
}

func init() {
	commandregistry.Register(&PurgeServiceOffering{})
}

func (cmd *PurgeServiceOffering) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["f"] = &flags.BoolFlag{ShortName: "f", Usage: T("Force deletion without confirmation")}
	fs["p"] = &flags.StringFlag{ShortName: "p", Usage: T("Provider")}

	return commandregistry.CommandMetadata{
		Name:        "purge-service-offering",
		Description: T("Recursively remove a service and child objects from Cloud Foundry database without making requests to a service broker"),
		Usage: []string{
			T("CF_NAME purge-service-offering SERVICE [-p PROVIDER]"),
			"\n\n",
			scaryWarningMessage(),
		},
		Flags: fs,
	}
}

func (cmd *PurgeServiceOffering) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + commandregistry.Commands.CommandUsage("purge-service-offering"))
		return nil, fmt.Errorf("Incorrect usage: %d arguments of %d required", len(fc.Args()), 1)
	}

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}

	if fc.IsSet("p") {
		reqs = append(reqs, requirementsFactory.NewMaxAPIVersionRequirement("Option '-p'", cf.ServiceAuthTokenMaximumAPIVersion))
	}

	return reqs, nil
}

func (cmd *PurgeServiceOffering) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.serviceRepo = deps.RepoLocator.GetServiceRepository()
	return cmd
}

func scaryWarningMessage() string {
	return T(`WARNING: This operation assumes that the service broker responsible for this service offering is no longer available, and all service instances have been deleted, leaving orphan records in Cloud Foundry's database. All knowledge of the service will be removed from Cloud Foundry, including service instances and service bindings. No attempt will be made to contact the service broker; running this command without destroying the service broker will cause orphan service instances. After running this command you may want to run either delete-service-auth-token or delete-service-broker to complete the cleanup.`)
}

func (cmd *PurgeServiceOffering) Execute(c flags.FlagContext) error {
	serviceName := c.Args()[0]

	var offering models.ServiceOffering
	if c.IsSet("p") {
		var err error
		offering, err = cmd.serviceRepo.FindServiceOfferingByLabelAndProvider(serviceName, c.String("p"))
		if err != nil {
			if _, ok := err.(*errors.ModelNotFoundError); ok {
				cmd.ui.Warn(T("Service offering does not exist\nTIP: If you are trying to purge a v1 service offering, you must set the -p flag."))
				return nil
			}
			return err
		}
	} else {
		offerings, err := cmd.serviceRepo.FindServiceOfferingsByLabel(serviceName)
		if err != nil {
			if _, ok := err.(*errors.ModelNotFoundError); ok {
				cmd.ui.Warn(T("Service offering does not exist\nTIP: If you are trying to purge a v1 service offering, you must set the -p flag."))
				return nil
			}
			return err
		}
		offering = offerings[0]
	}

	confirmed := c.Bool("f")
	if !confirmed {
		cmd.ui.Warn(scaryWarningMessage())
		confirmed = cmd.ui.Confirm(T("Really purge service offering {{.ServiceName}} from Cloud Foundry?",
			map[string]interface{}{"ServiceName": serviceName},
		))
	}

	if !confirmed {
		return nil
	}

	cmd.ui.Say(T("Purging service {{.ServiceName}}...", map[string]interface{}{"ServiceName": serviceName}))

	err := cmd.serviceRepo.PurgeServiceOffering(offering)
	if err != nil {
		return err
	}

	cmd.ui.Ok()
	return nil
}
