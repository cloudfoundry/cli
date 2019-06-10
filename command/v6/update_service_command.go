package v6

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/composite"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v6/shared"
)

//go:generate counterfeiter . UpdateServiceActor

type UpdateServiceActor interface {
	CloudControllerAPIVersion() string
	GetServiceInstanceByNameAndSpace(name string, spaceGUID string) (v2action.ServiceInstance, v2action.Warnings, error)
	UpgradeServiceInstance(serviceInstanceGUID, servicePlanGUID string) (v2action.Warnings, error)
}

type textData map[string]interface{}

type UpdateServiceCommand struct {
	RequiredArgs     flag.ServiceInstance `positional-args:"yes"`
	ParametersAsJSON flag.Path            `short:"c" description:"Valid JSON object containing service-specific configuration parameters, provided either in-line or in a file. For a list of supported configuration parameters, see documentation for the particular service offering."`
	Plan             string               `short:"p" description:"Change service plan for a service instance"`
	Tags             string               `short:"t" description:"User provided tags"`
	usage            interface{}          `usage:"CF_NAME update-service SERVICE_INSTANCE [-p NEW_PLAN] [-c PARAMETERS_AS_JSON] [-t TAGS] [--upgrade]\n\n   Optionally provide service-specific configuration parameters in a valid JSON object in-line:\n   CF_NAME update-service SERVICE_INSTANCE -c '{\"name\":\"value\",\"name\":\"value\"}'\n\n   Optionally provide a file containing service-specific configuration parameters in a valid JSON object. \n   The path to the parameters file can be an absolute or relative path to a file:\n   CF_NAME update-service SERVICE_INSTANCE -c PATH_TO_FILE\n\n   Example of valid JSON object:\n   {\n      \"cluster_nodes\": {\n         \"count\": 5,\n         \"memory_mb\": 1024\n      }\n   }\n\n   Optionally provide a list of comma-delimited tags that will be written to the VCAP_SERVICES environment variable for any bound applications.\n\nEXAMPLES:\n   CF_NAME update-service mydb -p gold\n   CF_NAME update-service mydb -c '{\"ram_gb\":4}'\n   CF_NAME update-service mydb -c ~/workspace/tmp/instance_config.json\n   CF_NAME update-service mydb -t \"list, of, tags\"\n   CF_NAME update-service mydb --upgrade"`
	relatedCommands  interface{}          `related_commands:"rename-service, services, update-user-provided-service"`
	Upgrade          bool                 `short:"u" long:"upgrade" description:"Upgrade the service instance to the latest version of the service plan available. This flag is in EXPERIMENTAL stage and may change without notice. It cannot be combined with other flags."`

	UI          command.UI
	Actor       UpdateServiceActor
	SharedActor command.SharedActor
	Config      command.Config
}

func (cmd *UpdateServiceCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config

	cmd.SharedActor = sharedaction.NewActor(config)

	ccClient, uaaClient, err := shared.NewClients(config, ui, true)
	if err != nil {
		return err
	}

	baseActor := v2action.NewActor(ccClient, uaaClient, config)
	cmd.Actor = &composite.UpdateServiceInstanceCompositeActor{
		GetAPIVersionActor:                        baseActor,
		GetServicePlanActor:                       baseActor,
		GetServiceInstanceActor:                   baseActor,
		UpdateServiceInstanceMaintenanceInfoActor: baseActor,
	}

	return nil
}

func (cmd *UpdateServiceCommand) Execute(args []string) error {
	if !cmd.Upgrade {
		return translatableerror.UnrefactoredCommandError{}
	}

	if len(args) > 0 {
		return translatableerror.TooManyArgumentsError{
			ExtraArgument: args[0],
		}
	}

	if err := cmd.validateArgumentCombination(); err != nil {
		return err
	}

	if err := command.MinimumCCAPIVersionCheck(cmd.Actor.CloudControllerAPIVersion(), ccversion.MinVersionUpdateServiceInstanceMaintenanceInfoV2, "Option '--upgrade'"); err != nil {
		return err
	}

	if err := cmd.SharedActor.CheckTarget(true, true); err != nil {
		return err
	}

	instance, warnings, err := cmd.Actor.GetServiceInstanceByNameAndSpace(cmd.RequiredArgs.ServiceInstance, cmd.Config.TargetedSpace().GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	proceed, err := cmd.promptForUpgrade()

	if err != nil {
		return err
	}

	if !proceed {
		cmd.UI.DisplayText("Update cancelled")
		return nil
	}

	return cmd.performUpgrade(instance)
}

func (cmd *UpdateServiceCommand) promptForUpgrade() (bool, error) {
	var serviceName = textData{"ServiceName": cmd.RequiredArgs.ServiceInstance}

	cmd.UI.DisplayText("This command is in EXPERIMENTAL stage and may change without notice.")
	cmd.UI.DisplayTextWithFlavor("You are about to update {{.ServiceName}}.", serviceName)
	cmd.UI.DisplayText("Warning: This operation may be long running and will block further operations on the service until complete.")

	return cmd.UI.DisplayBoolPrompt(false, "Really update service {{.ServiceName}}?", serviceName)
}

func (cmd *UpdateServiceCommand) performUpgrade(instance v2action.ServiceInstance) error {
	warnings, err := cmd.Actor.UpgradeServiceInstance(instance.GUID, instance.ServicePlanGUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()
	return nil
}

func (cmd *UpdateServiceCommand) validateArgumentCombination() error {
	if cmd.Tags != "" || cmd.ParametersAsJSON != "" || cmd.Plan != "" {
		return translatableerror.ArgumentCombinationError{
			Args: []string{"--upgrade", "-t", "-c", "-p"},
		}
	}

	return nil
}
