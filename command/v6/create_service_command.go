package v6

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v6/shared"
)

//go:generate counterfeiter . CreateServiceActor

type CreateServiceActor interface {
	CreateServiceInstance(spaceGUID, serviceName, servicePlanName, serviceInstanceName string, params map[string]interface{}, tags []string) (v2action.ServiceInstance, v2action.Warnings, error)
}

type CreateServiceCommand struct {
	RequiredArgs     flag.CreateServiceArgs        `positional-args:"yes"`
	ParametersAsJSON flag.JSONOrFileWithValidation `short:"c" description:"Valid JSON object containing service-specific configuration parameters, provided either in-line or in a file. For a list of supported configuration parameters, see documentation for the particular service offering."`
	Tags             flag.Tags                     `short:"t" description:"User provided tags"`
	usage            interface{}                   `usage:"CF_NAME create-service SERVICE PLAN SERVICE_INSTANCE [-c PARAMETERS_AS_JSON] [-t TAGS]\n\n   Optionally provide service-specific configuration parameters in a valid JSON object in-line:\n\n   CF_NAME create-service SERVICE PLAN SERVICE_INSTANCE -c '{\"name\":\"value\",\"name\":\"value\"}'\n\n   Optionally provide a file containing service-specific configuration parameters in a valid JSON object.\n   The path to the parameters file can be an absolute or relative path to a file:\n\n   CF_NAME create-service SERVICE PLAN SERVICE_INSTANCE -c PATH_TO_FILE\n\n   Example of valid JSON object:\n   {\n      \"cluster_nodes\": {\n         \"count\": 5,\n         \"memory_mb\": 1024\n      }\n   }\n\nTIP:\n   Use 'CF_NAME create-user-provided-service' to make user-provided services available to CF apps\n\nEXAMPLES:\n   Linux/Mac:\n      CF_NAME create-service db-service silver mydb -c '{\"ram_gb\":4}'\n\n   Windows Command Line:\n      CF_NAME create-service db-service silver mydb -c \"{\\\"ram_gb\\\":4}\"\n\n   Windows PowerShell:\n      CF_NAME create-service db-service silver mydb -c '{\\\"ram_gb\\\":4}'\n\n   CF_NAME create-service db-service silver mydb -c ~/workspace/tmp/instance_config.json\n\n   CF_NAME create-service db-service silver mydb -t \"list, of, tags\""`
	relatedCommands  interface{}                   `related_commands:"bind-service, create-user-provided-service, marketplace, services"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       CreateServiceActor
}

func (cmd *CreateServiceCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config

	cmd.SharedActor = sharedaction.NewActor(config)

	ccClient, uaaClient, err := shared.NewClients(config, ui, true)
	if err != nil {
		return err
	}

	cmd.Actor = v2action.NewActor(ccClient, uaaClient, config)

	return nil
}

func (cmd CreateServiceCommand) Execute(args []string) error {
	if len(args) > 0 {
		return translatableerror.TooManyArgumentsError{
			ExtraArgument: args[0],
		}
	}

	if err := cmd.SharedActor.CheckTarget(true, true); err != nil {
		return err
	}

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
		})

	instance, warnings, err := cmd.Actor.CreateServiceInstance(
		cmd.Config.TargetedSpace().GUID,
		cmd.RequiredArgs.Service,
		cmd.RequiredArgs.ServicePlan,
		cmd.RequiredArgs.ServiceInstance,
		cmd.ParametersAsJSON,
		cmd.Tags,
	)

	cmd.UI.DisplayWarnings(warnings)

	if _, nameTakenError := err.(ccerror.ServiceInstanceNameTakenError); nameTakenError {
		cmd.UI.DisplayOK()
		cmd.UI.DisplayTextWithFlavor("Service {{.ServiceInstance}} already exists",
			map[string]interface{}{
				"ServiceInstance": cmd.RequiredArgs.ServiceInstance,
			})
		return nil
	}
	if err != nil {
		return err
	}

	if instance.LastOperation.State == constant.LastOperationInProgress {
		cmd.UI.DisplayOK()
		cmd.UI.DisplayNewline()
		cmd.UI.DisplayTextWithFlavor("Create in progress. Use 'cf services' or 'cf service {{.ServiceInstance}}' to check operation status.",
			map[string]interface{}{
				"ServiceInstance": cmd.RequiredArgs.ServiceInstance,
			})
		return nil
	}

	cmd.UI.DisplayOK()
	return nil
}
