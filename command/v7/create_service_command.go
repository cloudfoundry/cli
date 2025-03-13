package v7

import (
	"strings"

	"code.cloudfoundry.org/cli/v9/actor/v7action"
	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/v9/command/flag"
	"code.cloudfoundry.org/cli/v9/command/v7/shared"
	"code.cloudfoundry.org/cli/v9/types"
)

type CreateServiceCommand struct {
	BaseCommand

	RequiredArgs     flag.CreateServiceArgs        `positional-args:"yes"`
	ServiceBroker    string                        `short:"b" description:"Create a service instance from a particular broker. Required when service offering name is ambiguous"`
	ParametersAsJSON flag.JSONOrFileWithValidation `short:"c" description:"Valid JSON object containing service-specific configuration parameters, provided either in-line or in a file. For a list of supported configuration parameters, see documentation for the particular service offering."`
	Tags             flag.Tags                     `short:"t" description:"User provided tags"`
	Wait             bool                          `short:"w" long:"wait" description:"Wait for the operation to complete"`
	relatedCommands  interface{}                   `related_commands:"bind-service, create-user-provided-service, marketplace, services"`
}

func (cmd CreateServiceCommand) Usage() string {
	return `
CF_NAME create-service SERVICE_OFFERING PLAN SERVICE_INSTANCE [-b SERVICE_BROKER] [-c PARAMETERS_AS_JSON] [-t TAGS]

Optionally provide service-specific configuration parameters in a valid JSON object in-line:

CF_NAME create-service SERVICE_OFFERING PLAN SERVICE_INSTANCE -c '{"name":"value","name":"value"}'

Optionally provide a file containing service-specific configuration parameters in a valid JSON object.
The path to the parameters file can be an absolute or relative path to a file:

CF_NAME create-service SERVICE_OFFERING PLAN SERVICE_INSTANCE -c PATH_TO_FILE

Example of valid JSON object:
{
  "cluster_nodes": {
    "count": 5,
    "memory_mb": 1024
  }
}

TIP:
	Use 'CF_NAME create-user-provided-service' to make user-provided service instances available to CF apps

EXAMPLES:
	Linux/Mac:
	   CF_NAME create-service db-service silver mydb -c '{"ram_gb":4}'

	Windows Command Line:
	   CF_NAME create-service db-service silver mydb -c "{\"ram_gb\":4}"

	Windows PowerShell:
	   CF_NAME create-service db-service silver mydb -c '{\"ram_gb\":4}'

	   CF_NAME create-service db-service silver mydb -c ~/workspace/tmp/instance_config.json

	   CF_NAME create-service db-service silver mydb -t "list, of, tags"
`
}

func (cmd CreateServiceCommand) Execute(args []string) error {
	if err := cmd.SharedActor.CheckTarget(true, true); err != nil {
		return err
	}

	cmd.RequiredArgs.ServiceInstance = strings.TrimSpace(cmd.RequiredArgs.ServiceInstance)

	if err := cmd.displayCreatingMessage(); err != nil {
		return err
	}

	stream, warnings, err := cmd.Actor.CreateManagedServiceInstance(
		v7action.CreateManagedServiceInstanceParams{
			ServiceOfferingName: cmd.RequiredArgs.ServiceOffering,
			ServicePlanName:     cmd.RequiredArgs.ServicePlan,
			ServiceInstanceName: cmd.RequiredArgs.ServiceInstance,
			ServiceBrokerName:   cmd.ServiceBroker,
			SpaceGUID:           cmd.Config.TargetedSpace().GUID,
			Tags:                types.OptionalStringSlice(cmd.Tags),
			Parameters:          types.OptionalObject(cmd.ParametersAsJSON),
		},
	)
	cmd.UI.DisplayWarnings(warnings)
	switch err.(type) {
	case nil:
	case ccerror.ServiceInstanceNameTakenError:
		cmd.UI.DisplayOK()
		cmd.UI.DisplayTextWithFlavor("Service instance {{.ServiceInstanceName}} already exists", cmd.serviceInstanceName())
		return nil
	default:
		return err
	}

	cmd.UI.DisplayNewline()
	complete, err := shared.WaitForResult(stream, cmd.UI, cmd.Wait)
	switch {
	case err != nil:
		return err
	case complete:
		cmd.UI.DisplayTextWithFlavor("Service instance {{.ServiceInstanceName}} created.", cmd.serviceInstanceName())
	default:
		cmd.UI.DisplayTextWithFlavor("Create in progress. Use 'cf services' or 'cf service {{.ServiceInstanceName}}' to check operation status.", cmd.serviceInstanceName())
	}

	cmd.UI.DisplayOK()
	return nil
}

func (cmd CreateServiceCommand) displayCreatingMessage() error {
	user, err := cmd.Actor.GetCurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Creating service instance {{.ServiceInstance}} in org {{.Org}} / space {{.Space}} as {{.User}}...",
		map[string]interface{}{
			"ServiceInstance": cmd.RequiredArgs.ServiceInstance,
			"Org":             cmd.Config.TargetedOrganization().Name,
			"Space":           cmd.Config.TargetedSpace().Name,
			"User":            user.Name,
		},
	)

	return nil
}

func (cmd CreateServiceCommand) serviceInstanceName() map[string]interface{} {
	return map[string]interface{}{
		"ServiceInstanceName": cmd.RequiredArgs.ServiceInstance,
	}
}
