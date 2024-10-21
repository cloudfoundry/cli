package v7

import (
	"fmt"
	"strings"

	"code.cloudfoundry.org/cli/v9/actor/actionerror"
	"code.cloudfoundry.org/cli/v9/actor/v7action"
	"code.cloudfoundry.org/cli/v9/command/flag"
	"code.cloudfoundry.org/cli/v9/command/v7/shared"
	"code.cloudfoundry.org/cli/v9/types"
)

type UpdateServiceCommand struct {
	BaseCommand

	RequiredArgs flag.ServiceInstance          `positional-args:"yes"`
	Parameters   flag.JSONOrFileWithValidation `short:"c" description:"Valid JSON object containing service-specific configuration parameters, provided either in-line or in a file. For a list of supported configuration parameters, see documentation for the particular service offering."`
	Plan         string                        `short:"p" description:"Change service plan for a service instance"`
	Tags         flag.Tags                     `short:"t" description:"User provided tags"`
	Wait         bool                          `short:"w" long:"wait" description:"Wait for the operation to complete"`
	Upgrade      bool                          `long:"upgrade" hidden:"true"`

	relatedCommands interface{} `related_commands:"rename-service, services, update-user-provided-service"`
}

func (cmd UpdateServiceCommand) Execute(args []string) error {
	if cmd.Upgrade {
		return fmt.Errorf(
			`Upgrading is no longer supported via updates, please run "cf upgrade-service %s" instead.`,
			cmd.RequiredArgs.ServiceInstance,
		)
	}

	if err := cmd.SharedActor.CheckTarget(true, true); err != nil {
		return err
	}

	if err := cmd.displayIntro(); err != nil {
		return err
	}

	if cmd.noFlagsProvided() {
		cmd.UI.DisplayText("No flags specified. No changes were made.")
		cmd.UI.DisplayOK()
		return nil
	}

	stream, warnings, err := cmd.Actor.UpdateManagedServiceInstance(
		v7action.UpdateManagedServiceInstanceParams{
			ServiceInstanceName: string(cmd.RequiredArgs.ServiceInstance),
			ServicePlanName:     cmd.Plan,
			SpaceGUID:           cmd.Config.TargetedSpace().GUID,
			Tags:                types.OptionalStringSlice(cmd.Tags),
			Parameters:          types.OptionalObject(cmd.Parameters),
		},
	)
	cmd.UI.DisplayWarnings(warnings)
	switch err.(type) {
	case nil:
	case actionerror.ServiceInstanceUpdateIsNoop:
		cmd.UI.DisplayText("No changes were made.")
		cmd.UI.DisplayOK()
		return nil
	default:
		return err
	}

	complete, err := shared.WaitForResult(stream, cmd.UI, cmd.Wait)
	switch {
	case err != nil:
		return err
	case complete:
		cmd.UI.DisplayTextWithFlavor("Update of service instance {{.ServiceInstance}} complete.", cmd.serviceInstanceName())
	default:
		cmd.UI.DisplayTextWithFlavor("Update in progress. Use 'cf services' or 'cf service {{.ServiceInstance}}' to check operation status.", cmd.serviceInstanceName())
	}

	cmd.UI.DisplayOK()
	return nil
}

func (cmd UpdateServiceCommand) Usage() string {
	return strings.TrimSpace(`
CF_NAME update-service SERVICE_INSTANCE [-p NEW_PLAN] [-c PARAMETERS_AS_JSON] [-t TAGS]

Optionally provide service-specific configuration parameters in a valid JSON object in-line:
CF_NAME update-service SERVICE_INSTANCE -c '{"name":"value","name":"value"}'

Optionally provide a file containing service-specific configuration parameters in a valid JSON object.
The path to the parameters file can be an absolute or relative path to a file:
CF_NAME update-service SERVICE_INSTANCE -c PATH_TO_FILE

Example of valid JSON object:
{
   "cluster_nodes": {
      "count": 5,
      "memory_mb": 1024
   }
}

Optionally provide a list of comma-delimited tags that will be written to the VCAP_SERVICES environment variable for any bound applications.
`,
	)
}

func (cmd UpdateServiceCommand) Examples() string {
	return strings.TrimSpace(`
CF_NAME update-service mydb -p gold
CF_NAME update-service mydb -c '{"ram_gb":4}'
CF_NAME update-service mydb -c ~/workspace/tmp/instance_config.json
CF_NAME update-service mydb -t "list, of, tags"
`,
	)
}

func (cmd UpdateServiceCommand) displayIntro() error {
	user, err := cmd.Actor.GetCurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor(
		"Updating service instance {{.ServiceInstanceName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...",
		map[string]interface{}{
			"ServiceInstanceName": cmd.RequiredArgs.ServiceInstance,
			"OrgName":             cmd.Config.TargetedOrganization().Name,
			"SpaceName":           cmd.Config.TargetedSpace().Name,
			"Username":            user.Name,
		},
	)
	cmd.UI.DisplayNewline()

	return nil
}

func (cmd UpdateServiceCommand) noFlagsProvided() bool {
	return !cmd.Tags.IsSet && !cmd.Parameters.IsSet && cmd.Plan == ""
}

func (cmd UpdateServiceCommand) serviceInstanceName() map[string]interface{} {
	return map[string]interface{}{
		"ServiceInstance": cmd.RequiredArgs.ServiceInstance,
	}
}
