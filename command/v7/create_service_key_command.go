package v7

import (
	"code.cloudfoundry.org/cli/v9/actor/actionerror"
	"code.cloudfoundry.org/cli/v9/actor/v7action"
	"code.cloudfoundry.org/cli/v9/command/flag"
	"code.cloudfoundry.org/cli/v9/command/v7/shared"
	"code.cloudfoundry.org/cli/v9/types"
)

type CreateServiceKeyCommand struct {
	BaseCommand

	RequiredArgs     flag.ServiceInstanceKey       `positional-args:"yes"`
	ParametersAsJSON flag.JSONOrFileWithValidation `short:"c" description:"Valid JSON object containing service-specific configuration parameters, provided either in-line or in a file. For a list of supported configuration parameters, see documentation for the particular service offering."`
	Wait             bool                          `short:"w" long:"wait" description:"Wait for the operation to complete"`
	relatedCommands  interface{}                   `related_commands:"service-key"`
}

func (cmd CreateServiceKeyCommand) Execute(args []string) error {
	if err := cmd.SharedActor.CheckTarget(true, true); err != nil {
		return err
	}

	if err := cmd.displayIntro(); err != nil {
		return err
	}

	stream, warnings, err := cmd.Actor.CreateServiceKey(v7action.CreateServiceKeyParams{
		SpaceGUID:           cmd.Config.TargetedSpace().GUID,
		ServiceInstanceName: cmd.RequiredArgs.ServiceInstance,
		ServiceKeyName:      cmd.RequiredArgs.ServiceKey,
		Parameters:          types.OptionalObject(cmd.ParametersAsJSON),
	})
	cmd.UI.DisplayWarnings(warnings)
	switch err.(type) {
	case nil:
	case actionerror.ResourceAlreadyExistsError:
		cmd.displayAlreadyExists()
		return nil
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
		cmd.UI.DisplayText("Create in progress.")
		return nil
	}
}

func (cmd CreateServiceKeyCommand) Usage() string {
	return `
CF_NAME create-service-key SERVICE_INSTANCE SERVICE_KEY [-c PARAMETERS_AS_JSON] [--wait]

Optionally provide service-specific configuration parameters in a valid JSON object in-line.
CF_NAME create-service-key SERVICE_INSTANCE SERVICE_KEY -c '{"name":"value","name":"value"}'

Optionally provide a file containing service-specific configuration parameters in a valid JSON object. The path to the parameters file can be an absolute or relative path to a file.
CF_NAME create-service-key SERVICE_INSTANCE SERVICE_KEY -c PATH_TO_FILE

Example of valid JSON object:
{
  "permissions": "read-only"
}
`
}

func (cmd CreateServiceKeyCommand) Examples() string {
	return `
CF_NAME create-service-key mydb mykey -c '{"permissions":"read-only"}'
CF_NAME create-service-key mydb mykey -c ~/workspace/tmp/instance_config.json
`
}

func (cmd CreateServiceKeyCommand) displayIntro() error {
	user, err := cmd.Actor.GetCurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor(
		"Creating service key {{.ServiceKey}} for service instance {{.ServiceInstance}} as {{.User}}...",
		map[string]interface{}{
			"ServiceInstance": cmd.RequiredArgs.ServiceInstance,
			"ServiceKey":      cmd.RequiredArgs.ServiceKey,
			"User":            user.Name,
		},
	)

	return nil
}

func (cmd CreateServiceKeyCommand) displayAlreadyExists() {
	cmd.UI.DisplayNewline()
	cmd.UI.DisplayText(
		"Service key {{.ServiceKey}} already exists",
		map[string]interface{}{"ServiceKey": cmd.RequiredArgs.ServiceKey},
	)
	cmd.UI.DisplayOK()
}
