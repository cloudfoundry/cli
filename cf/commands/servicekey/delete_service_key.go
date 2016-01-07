package servicekey

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
	"github.com/cloudfoundry/cli/flags/flag"

	. "github.com/cloudfoundry/cli/cf/i18n"
)

type DeleteServiceKey struct {
	ui             terminal.UI
	config         core_config.Reader
	serviceRepo    api.ServiceRepository
	serviceKeyRepo api.ServiceKeyRepository
}

func init() {
	command_registry.Register(&DeleteServiceKey{})
}

func (cmd *DeleteServiceKey) MetaData() command_registry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["f"] = &cliFlags.BoolFlag{ShortName: "f", Usage: T("Force deletion without confirmation")}

	return command_registry.CommandMetadata{
		Name:        "delete-service-key",
		ShortName:   "dsk",
		Description: T("Delete a service key"),
		Usage: T(`CF_NAME delete-service-key SERVICE_INSTANCE SERVICE_KEY [-f]

EXAMPLE:
   CF_NAME delete-service-key mydb mykey`),
		Flags: fs,
	}
}

func (cmd *DeleteServiceKey) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 2 {
		cmd.ui.Failed(T("Incorrect Usage. Requires SERVICE_INSTANCE SERVICE_KEY as arguments\n\n") + command_registry.Commands.CommandUsage("delete-service-key"))
	}

	loginRequirement := requirementsFactory.NewLoginRequirement()
	targetSpaceRequirement := requirementsFactory.NewTargetedSpaceRequirement()

	reqs := []requirements.Requirement{loginRequirement, targetSpaceRequirement}
	return reqs, nil
}

func (cmd *DeleteServiceKey) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.serviceRepo = deps.RepoLocator.GetServiceRepository()
	cmd.serviceKeyRepo = deps.RepoLocator.GetServiceKeyRepository()
	return cmd
}

func (cmd *DeleteServiceKey) Execute(c flags.FlagContext) {
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
		cmd.ui.Warn(T("Service instance {{.ServiceInstanceName}} does not exist.",
			map[string]interface{}{
				"ServiceInstanceName": serviceInstanceName,
			}))
		return
	}

	serviceKey, err := cmd.serviceKeyRepo.GetServiceKey(serviceInstance.Guid, serviceKeyName)
	if err != nil || serviceKey.Fields.Guid == "" {
		switch err.(type) {
		case *errors.NotAuthorizedError:
			cmd.ui.Say(T("No service key {{.ServiceKeyName}} found for service instance {{.ServiceInstanceName}}",
				map[string]interface{}{
					"ServiceKeyName":      terminal.EntityNameColor(serviceKeyName),
					"ServiceInstanceName": terminal.EntityNameColor(serviceInstanceName)}))
			return
		default:
			cmd.ui.Ok()
			cmd.ui.Warn(T("Service key {{.ServiceKeyName}} does not exist for service instance {{.ServiceInstanceName}}.",
				map[string]interface{}{
					"ServiceKeyName":      serviceKeyName,
					"ServiceInstanceName": serviceInstanceName,
				}))
			return
		}
	}

	err = cmd.serviceKeyRepo.DeleteServiceKey(serviceKey.Fields.Guid)
	if err != nil {
		cmd.ui.Failed(err.Error())
		return
	}

	cmd.ui.Ok()
}
