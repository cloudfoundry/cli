package v7

import (
	"code.cloudfoundry.org/cli/v8/actor/actionerror"
	"code.cloudfoundry.org/cli/v8/command/flag"
	"code.cloudfoundry.org/cli/v8/command/translatableerror"
	"code.cloudfoundry.org/cli/v8/command/v7/shared"
)

type DeleteServiceKeyCommand struct {
	BaseCommand

	RequiredArgs    flag.ServiceInstanceKey `positional-args:"yes"`
	Force           bool                    `short:"f" description:"Force deletion without confirmation"`
	Wait            bool                    `short:"w" long:"wait" description:"Wait for the operation to complete"`
	relatedCommands interface{}             `related_commands:"service-keys"`
}

func (cmd DeleteServiceKeyCommand) Execute(args []string) error {
	if err := cmd.SharedActor.CheckTarget(true, true); err != nil {
		return err
	}

	if !cmd.Force {
		delete, err := cmd.displayPrompt()
		if err != nil {
			return err
		}

		if !delete {
			cmd.UI.DisplayText("Delete cancelled")
			return nil
		}
	}

	if err := cmd.displayIntro(); err != nil {
		return err
	}

	stream, warnings, err := cmd.Actor.DeleteServiceKeyByServiceInstanceAndName(
		cmd.RequiredArgs.ServiceInstance,
		cmd.RequiredArgs.ServiceKey,
		cmd.Config.TargetedSpace().GUID,
	)
	cmd.UI.DisplayWarnings(warnings)
	switch err.(type) {
	case nil:
	case actionerror.ServiceKeyNotFoundError:
		cmd.displayNotFound()
	case actionerror.ServiceInstanceNotFoundError:
		return translatableerror.ServiceInstanceNotFoundError{Name: cmd.RequiredArgs.ServiceInstance}
	default:
		return err
	}

	completed, err := shared.WaitForResult(stream, cmd.UI, cmd.Wait)
	switch {
	case err != nil:
		return err
	case completed:
		cmd.UI.DisplayOK()
		return nil
	default:
		cmd.UI.DisplayOK()
		cmd.UI.DisplayText("Delete in progress.")
		return nil
	}
}

func (cmd DeleteServiceKeyCommand) Usage() string {
	return `CF_NAME delete-service-key SERVICE_INSTANCE SERVICE_KEY [-f] [--wait]`
}

func (cmd DeleteServiceKeyCommand) Examples() string {
	return `CF_NAME delete-service-key mydb mykey`
}

func (cmd DeleteServiceKeyCommand) displayPrompt() (bool, error) {
	delete, err := cmd.UI.DisplayBoolPrompt(
		false,
		"Really delete the service key {{.ServiceKey}}?",
		cmd.names(),
	)
	if err != nil {
		return false, err
	}

	return delete, nil
}

func (cmd DeleteServiceKeyCommand) displayIntro() error {
	user, err := cmd.Actor.GetCurrentUser()
	if err != nil {
		return err
	}

	names := cmd.names()
	names["User"] = user.Name
	cmd.UI.DisplayTextWithFlavor(
		"Deleting key {{.ServiceKey}} for service instance {{.ServiceInstance}} as {{.User}}...",
		names,
	)

	return nil
}

func (cmd DeleteServiceKeyCommand) displayNotFound() {
	cmd.UI.DisplayNewline()
	cmd.UI.DisplayText("Service key {{.ServiceKey}} does not exist for service instance {{.ServiceInstance}}.", cmd.names())
}

func (cmd DeleteServiceKeyCommand) names() map[string]interface{} {
	return map[string]interface{}{
		"ServiceInstance": cmd.RequiredArgs.ServiceInstance,
		"ServiceKey":      cmd.RequiredArgs.ServiceKey,
	}
}
