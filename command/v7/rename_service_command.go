package v7

import (
	"strings"

	"code.cloudfoundry.org/cli/v9/actor/actionerror"
	"code.cloudfoundry.org/cli/v9/command/flag"
	"code.cloudfoundry.org/cli/v9/command/translatableerror"
)

type RenameServiceCommand struct {
	BaseCommand
	RequiredArgs    flag.RenameServiceArgs `positional-args:"yes"`
	relatedCommands interface{}            `related_commands:"services, update-service"`
}

func (cmd RenameServiceCommand) Execute(args []string) error {
	if err := cmd.SharedActor.CheckTarget(true, true); err != nil {
		return err
	}

	cmd.RequiredArgs.ServiceInstance = strings.TrimSpace(cmd.RequiredArgs.ServiceInstance)
	cmd.RequiredArgs.NewServiceInstanceName = strings.TrimSpace(cmd.RequiredArgs.NewServiceInstanceName)

	if err := cmd.displayMessage(); err != nil {
		return err
	}

	warnings, err := cmd.Actor.RenameServiceInstance(
		cmd.RequiredArgs.ServiceInstance,
		cmd.Config.TargetedSpace().GUID,
		cmd.RequiredArgs.NewServiceInstanceName,
	)
	cmd.UI.DisplayWarnings(warnings)

	switch e := err.(type) {
	case nil:
		cmd.UI.DisplayOK()
		return nil
	case actionerror.ServiceInstanceNotFoundError:
		cmd.UI.DisplayText("TIP: Use 'cf services' to view all services in this org and space.")
		return translatableerror.ServiceInstanceNotFoundError{Name: e.Name}
	default:
		return err
	}
}

func (cmd RenameServiceCommand) Usage() string {
	return "cf rename-service SERVICE_INSTANCE NEW_SERVICE_INSTANCE"
}

func (cmd RenameServiceCommand) displayMessage() error {
	user, err := cmd.Actor.GetCurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Renaming service {{.OriginalName}} to {{.NewName}} in org {{.Org}} / space {{.Space}} as {{.User}}...", map[string]interface{}{
		"OriginalName": cmd.RequiredArgs.ServiceInstance,
		"NewName":      cmd.RequiredArgs.NewServiceInstanceName,
		"Org":          cmd.Config.TargetedOrganization().Name,
		"Space":        cmd.Config.TargetedSpace().Name,
		"User":         user.Name,
	})

	return nil
}
