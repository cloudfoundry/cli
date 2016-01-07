package service

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/errors"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
	"github.com/cloudfoundry/cli/flags/flag"
)

type PurgeServiceOffering struct {
	ui          terminal.UI
	serviceRepo api.ServiceRepository
}

func init() {
	command_registry.Register(&PurgeServiceOffering{})
}

func (cmd *PurgeServiceOffering) MetaData() command_registry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["f"] = &cliFlags.BoolFlag{ShortName: "f", Usage: T("Force deletion without confirmation")}
	fs["p"] = &cliFlags.StringFlag{ShortName: "p", Usage: T("Provider")}

	return command_registry.CommandMetadata{
		Name:        "purge-service-offering",
		Description: T("Recursively remove a service and child objects from Cloud Foundry database without making requests to a service broker"),
		Usage:       T("CF_NAME purge-service-offering SERVICE [-p PROVIDER]") + "\n\n" + scaryWarningMessage(),
		Flags:       fs,
	}
}

func (cmd *PurgeServiceOffering) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + command_registry.Commands.CommandUsage("purge-service-offering"))
	}

	reqs = []requirements.Requirement{requirementsFactory.NewLoginRequirement()}
	return
}

func (cmd *PurgeServiceOffering) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.serviceRepo = deps.RepoLocator.GetServiceRepository()
	return cmd
}

func scaryWarningMessage() string {
	return T(`WARNING: This operation assumes that the service broker responsible for this service offering is no longer available, and all service instances have been deleted, leaving orphan records in Cloud Foundry's database. All knowledge of the service will be removed from Cloud Foundry, including service instances and service bindings. No attempt will be made to contact the service broker; running this command without destroying the service broker will cause orphan service instances. After running this command you may want to run either delete-service-auth-token or delete-service-broker to complete the cleanup.`)
}

func (cmd *PurgeServiceOffering) Execute(c flags.FlagContext) {
	serviceName := c.Args()[0]

	offering, apiErr := cmd.serviceRepo.FindServiceOfferingByLabelAndProvider(serviceName, c.String("p"))

	switch apiErr.(type) {
	case nil:
	case *errors.ModelNotFoundError:
		cmd.ui.Warn(T("Service offering does not exist\nTIP: If you are trying to purge a v1 service offering, you must set the -p flag."))
		return
	default:
		cmd.ui.Failed(apiErr.Error())
	}

	confirmed := c.Bool("f")
	if !confirmed {
		cmd.ui.Warn(scaryWarningMessage())
		confirmed = cmd.ui.Confirm(T("Really purge service offering {{.ServiceName}} from Cloud Foundry?",
			map[string]interface{}{"ServiceName": serviceName},
		))
	}

	if !confirmed {
		return
	}
	cmd.ui.Say(T("Purging service {{.ServiceName}}...", map[string]interface{}{"ServiceName": serviceName}))
	err := cmd.serviceRepo.PurgeServiceOffering(offering)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	cmd.ui.Ok()
}
