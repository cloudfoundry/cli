package service

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/errors"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/simonleung8/flags"
	"github.com/simonleung8/flags/flag"
)

type PurgeServiceInstance struct {
	ui          terminal.UI
	serviceRepo api.ServiceRepository
}

func init() {
	command_registry.Register(&PurgeServiceInstance{})
}

func (cmd *PurgeServiceInstance) MetaData() command_registry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["f"] = &cliFlags.BoolFlag{Name: "f", Usage: T("Force deletion without confirmation")}

	return command_registry.CommandMetadata{
		Name:        "purge-service-instance",
		Description: T("Recursively remove a service instance and child objects from Cloud Foundry database without making requests to a service broker"),
		Usage:       T("CF_NAME purge-service-instance SERVICE_INSTANCE") + "\n\n" + cmd.scaryWarningMessage(),
		Flags:       fs,
	}
}

func (cmd *PurgeServiceInstance) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + command_registry.Commands.CommandUsage("purge-service-instance"))
	}

	reqs = []requirements.Requirement{requirementsFactory.NewLoginRequirement()}
	return
}

func (cmd *PurgeServiceInstance) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.serviceRepo = deps.RepoLocator.GetServiceRepository()
	return cmd
}

func (cmd *PurgeServiceInstance) scaryWarningMessage() string {
	return T(`WARNING: This operation assumes that the service broker responsible for this service instance is no longer available or is not responding with a 200 or 410, and the service instance has been deleted, leaving orphan records in Cloud Foundry's database. All knowledge of the service instance will be removed from Cloud Foundry, including service bindings and service keys.`)
}

func (cmd *PurgeServiceInstance) Execute(c flags.FlagContext) {
	instanceName := c.Args()[0]

	instance, apiErr := cmd.serviceRepo.FindInstanceByName(instanceName)

	switch apiErr.(type) {
	case nil:
	case *errors.ModelNotFoundError:
		cmd.ui.Warn(T("Service instance {{.InstanceName}} not found",
			map[string]interface{}{"InstanceName": instanceName},
		))
		return
	default:
		cmd.ui.Failed(apiErr.Error())
	}

	confirmed := c.Bool("f")
	if !confirmed {
		cmd.ui.Warn(cmd.scaryWarningMessage())
		confirmed = cmd.ui.Confirm(T("Really purge service instance {{.InstanceName}} from Cloud Foundry?",
			map[string]interface{}{"InstanceName": instanceName},
		))
	}

	if !confirmed {
		return
	}
	cmd.ui.Say(T("Purging service {{.InstanceName}}...", map[string]interface{}{"InstanceName": instanceName}))
	err := cmd.serviceRepo.PurgeServiceInstance(instance)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	cmd.ui.Ok()
}
