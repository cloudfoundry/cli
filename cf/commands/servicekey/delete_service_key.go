package servicekey

import (
	"fmt"

	"code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/flags"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"

	. "code.cloudfoundry.org/cli/cf/i18n"
)

type DeleteServiceKey struct {
	ui             terminal.UI
	config         coreconfig.Reader
	serviceRepo    api.ServiceRepository
	serviceKeyRepo api.ServiceKeyRepository
}

func init() {
	commandregistry.Register(&DeleteServiceKey{})
}

func (cmd *DeleteServiceKey) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["f"] = &flags.BoolFlag{ShortName: "f", Usage: T("Force deletion without confirmation")}

	return commandregistry.CommandMetadata{
		Name:        "delete-service-key",
		ShortName:   "dsk",
		Description: T("Delete a service key"),
		Usage: []string{
			T("CF_NAME delete-service-key SERVICE_INSTANCE SERVICE_KEY [-f]"),
		},
		Examples: []string{
			"CF_NAME delete-service-key mydb mykey",
		},
		Flags: fs,
	}
}

func (cmd *DeleteServiceKey) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 2 {
		cmd.ui.Failed(T("Incorrect Usage. Requires SERVICE_INSTANCE SERVICE_KEY as arguments\n\n") + commandregistry.Commands.CommandUsage("delete-service-key"))
		return nil, fmt.Errorf("Incorrect usage: %d arguments of %d required", len(fc.Args()), 2)
	}

	loginRequirement := requirementsFactory.NewLoginRequirement()
	targetSpaceRequirement := requirementsFactory.NewTargetedSpaceRequirement()

	reqs := []requirements.Requirement{loginRequirement, targetSpaceRequirement}
	return reqs, nil
}

func (cmd *DeleteServiceKey) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.serviceRepo = deps.RepoLocator.GetServiceRepository()
	cmd.serviceKeyRepo = deps.RepoLocator.GetServiceKeyRepository()
	return cmd
}

func (cmd *DeleteServiceKey) Execute(c flags.FlagContext) error {
	serviceInstanceName := c.Args()[0]
	serviceKeyName := c.Args()[1]

	if !c.Bool("f") {
		if !cmd.ui.ConfirmDelete(T("service key"), serviceKeyName) {
			return nil
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
		return nil
	}

	serviceKey, err := cmd.serviceKeyRepo.GetServiceKey(serviceInstance.GUID, serviceKeyName)
	if err != nil || serviceKey.Fields.GUID == "" {
		switch err.(type) {
		case *errors.NotAuthorizedError:
			cmd.ui.Say(T("No service key {{.ServiceKeyName}} found for service instance {{.ServiceInstanceName}}",
				map[string]interface{}{
					"ServiceKeyName":      terminal.EntityNameColor(serviceKeyName),
					"ServiceInstanceName": terminal.EntityNameColor(serviceInstanceName)}))
			return nil
		default:
			cmd.ui.Ok()
			cmd.ui.Warn(T("Service key {{.ServiceKeyName}} does not exist for service instance {{.ServiceInstanceName}}.",
				map[string]interface{}{
					"ServiceKeyName":      serviceKeyName,
					"ServiceInstanceName": serviceInstanceName,
				}))
			return nil
		}
	}

	err = cmd.serviceKeyRepo.DeleteServiceKey(serviceKey.Fields.GUID)
	if err != nil {
		return err
	}

	cmd.ui.Ok()
	return nil
}
