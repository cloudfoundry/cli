package v7

import (
	"code.cloudfoundry.org/cli/v8/actor/actionerror"
	"code.cloudfoundry.org/cli/v8/actor/v7action"
	"code.cloudfoundry.org/cli/v8/command/flag"
	"code.cloudfoundry.org/cli/v8/command/v7/shared"
	"code.cloudfoundry.org/cli/v8/types"
)

type BindServiceCommand struct {
	BaseCommand

	RequiredArgs     flag.BindServiceArgs          `positional-args:"yes"`
	BindingName      flag.BindingName              `long:"binding-name" description:"Name to expose service instance to app process with (Default: service instance name)"`
	ParametersAsJSON flag.JSONOrFileWithValidation `short:"c" description:"Valid JSON object containing service-specific configuration parameters, provided either in-line or in a file. For a list of supported configuration parameters, see documentation for the particular service offering."`
	Wait             bool                          `short:"w" long:"wait" description:"Wait for the operation to complete"`
	relatedCommands  interface{}                   `related_commands:"services"`
}

func (cmd BindServiceCommand) Execute(args []string) error {
	if err := cmd.SharedActor.CheckTarget(true, true); err != nil {
		return err
	}

	if err := cmd.displayIntro(); err != nil {
		return err
	}

	stream, warnings, err := cmd.Actor.CreateServiceAppBinding(v7action.CreateServiceAppBindingParams{
		SpaceGUID:           cmd.Config.TargetedSpace().GUID,
		ServiceInstanceName: cmd.RequiredArgs.ServiceInstanceName,
		AppName:             cmd.RequiredArgs.AppName,
		BindingName:         cmd.BindingName.Value,
		Parameters:          types.OptionalObject(cmd.ParametersAsJSON),
	})
	cmd.UI.DisplayWarnings(warnings)

	switch err.(type) {
	case nil:
	case actionerror.ResourceAlreadyExistsError:
		cmd.UI.DisplayText("App {{.AppName}} is already bound to service instance {{.ServiceInstanceName}}.", cmd.names())
		cmd.UI.DisplayOK()
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
		cmd.UI.DisplayText("TIP: Use 'cf restage {{.AppName}}' to ensure your env variable changes take effect", cmd.names())
		return nil
	default:
		cmd.UI.DisplayOK()
		cmd.UI.DisplayText("Binding in progress. Use 'cf service {{.ServiceInstanceName}}' to check operation status.", cmd.names())
		cmd.UI.DisplayNewline()
		cmd.UI.DisplayText("TIP: Once this operation succeeds, use 'cf restage {{.AppName}}' to ensure your env variable changes take effect", cmd.names())
		return nil
	}
}

func (cmd BindServiceCommand) Usage() string {
	return `CF_NAME bind-service APP_NAME SERVICE_INSTANCE [-c PARAMETERS_AS_JSON] [--binding-name BINDING_NAME]

Optionally provide service-specific configuration parameters in a valid JSON object in-line:

CF_NAME bind-service APP_NAME SERVICE_INSTANCE -c '{"name":"value","name":"value"}'

Optionally provide a file containing service-specific configuration parameters in a valid JSON object.
The path to the parameters file can be an absolute or relative path to a file.

CF_NAME bind-service APP_NAME SERVICE_INSTANCE -c PATH_TO_FILE

Example of valid JSON object:
{
   "permissions": "read-only"
}

Optionally provide a binding name for the association between an app and a service instance:

CF_NAME bind-service APP_NAME SERVICE_INSTANCE --binding-name BINDING_NAME`
}

func (cmd BindServiceCommand) Examples() string {
	return `
Linux/Mac:
   CF_NAME bind-service myapp mydb -c '{"permissions":"read-only"}'

Windows Command Line:
   CF_NAME bind-service myapp mydb -c "{\"permissions\":\"read-only\"}"

Windows PowerShell:
   CF_NAME bind-service myapp mydb -c '{\"permissions\":\"read-only\"}'

CF_NAME bind-service myapp mydb -c ~/workspace/tmp/instance_config.json --binding-name BINDING_NAME
`
}

func (cmd BindServiceCommand) displayIntro() error {
	user, err := cmd.Actor.GetCurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor(
		"Binding service instance {{.ServiceInstance}} to app {{.AppName}} in org {{.Org}} / space {{.Space}} as {{.User}}...",
		map[string]interface{}{
			"ServiceInstance": cmd.RequiredArgs.ServiceInstanceName,
			"AppName":         cmd.RequiredArgs.AppName,
			"User":            user.Name,
			"Space":           cmd.Config.TargetedSpace().Name,
			"Org":             cmd.Config.TargetedOrganization().Name,
		},
	)

	return nil
}

func (cmd BindServiceCommand) names() map[string]interface{} {
	return map[string]interface{}{
		"ServiceInstanceName": cmd.RequiredArgs.ServiceInstanceName,
		"AppName":             cmd.RequiredArgs.AppName,
	}
}
