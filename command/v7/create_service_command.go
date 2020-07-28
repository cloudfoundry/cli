package v7

import (
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/types"
)

type CreateServiceCommand struct {
	BaseCommand

	RequiredArgs     flag.CreateServiceArgs        `positional-args:"yes"`
	ServiceBroker    string                        `short:"b" description:"Create a service instance from a particular broker. Required when service offering name is ambiguous"`
	ParametersAsJSON flag.JSONOrFileWithValidation `short:"c" description:"Valid JSON object containing service-specific configuration parameters, provided either in-line or in a file. For a list of supported configuration parameters, see documentation for the particular service offering."`
	Tags             flag.Tags                     `short:"t" description:"User provided tags"`
	relatedCommands  interface{}                   `related_commands:"bind-service, create-user-provided-service, marketplace, services"`
}

func (cmd CreateServiceCommand) Usage() string {
	return `
CF_NAME create-service SERVICE PLAN SERVICE_INSTANCE [-b SERVICE_BROKER] [-c JSON_PARAMS] [-t TAGS]

Optionally provide service-specific configuration parameters in a valid JSON object in-line:

CF_NAME create-service SERVICE PLAN SERVICE_INSTANCE -c '{"name":"value","name":"value"}'

Optionally provide a file containing service-specific configuration parameters in a valid JSON object.
The path to the parameters file can be an absolute or relative path to a file:

CF_NAME create-service SERVICE PLAN SERVICE_INSTANCE -c PATH_TO_FILE

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

	if err := cmd.displayCreatingMessage(); err != nil {
		return err
	}

	warnings, err := cmd.Actor.CreateManagedServiceInstance(
		v7action.ManagedServiceInstanceParams{
			ServiceOfferingName: cmd.RequiredArgs.Service,
			ServicePlanName:     cmd.RequiredArgs.ServicePlan,
			ServiceInstanceName: cmd.RequiredArgs.ServiceInstance,
			ServiceBrokerName:   cmd.ServiceBroker,
			SpaceGUID:           cmd.Config.TargetedSpace().GUID,
			Tags:                types.OptionalStringSlice(cmd.Tags),
			Parameters:          types.OptionalObject(cmd.ParametersAsJSON),
		},
	)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		if _, nameTakenError := err.(ccerror.ServiceInstanceNameTakenError); nameTakenError {
			cmd.UI.DisplayOK()
			cmd.UI.DisplayTextWithFlavor("Service {{.ServiceInstance}} already exists",
				map[string]interface{}{
					"ServiceInstance": cmd.RequiredArgs.ServiceInstance,
				})
			return nil
		}
		return err
	}

	cmd.UI.DisplayOK()
	cmd.displayCreateInProgressMessage()
	return nil
}

func (cmd CreateServiceCommand) displayCreateInProgressMessage() {
	cmd.UI.DisplayTextWithFlavor("Create in progress. Use 'cf services' or 'cf service {{.ServiceInstance}}' to check operation status.",
		map[string]interface{}{
			"ServiceInstance": cmd.RequiredArgs.ServiceInstance,
		})
}

func (cmd CreateServiceCommand) displayCreatingMessage() error {
	user, err := cmd.Config.CurrentUser()
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
