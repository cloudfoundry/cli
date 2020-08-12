package v7

import (
	"fmt"
	"strings"

	"code.cloudfoundry.org/cli/actor/v7action"

	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/types"
)

type UpdateServiceCommand struct {
	BaseCommand

	RequiredArgs flag.ServiceInstance          `positional-args:"yes"`
	Parameters   flag.JSONOrFileWithValidation `short:"c" description:"Valid JSON object containing service-specific configuration parameters, provided either in-line or in a file. For a list of supported configuration parameters, see documentation for the particular service offering."`
	Plan         flag.OptionalString           `short:"p" description:"Change service plan for a service instance"`
	Tags         flag.Tags                     `short:"t" description:"User provided tags"`
	Upgrade      bool                          `long:"upgrade" hidden:"true"`

	relatedCommands interface{} `related_commands:"rename-service, services, update-user-provided-service"`
}

func (cmd UpdateServiceCommand) Execute(args []string) error {
	if err := cmd.SharedActor.CheckTarget(true, true); err != nil {
		return err
	}

	if err := cmd.displayIntro(); err != nil {
		return err
	}

	if cmd.Upgrade {
		return fmt.Errorf(
			`Upgrading is no longer supported via updates, please run "cf upgrade-service %s" instead.`,
			cmd.RequiredArgs.ServiceInstance,
		)
	}

	if cmd.noFlagsProvided() {
		cmd.UI.DisplayText("No flags specified. No changes were made.")
		cmd.UI.DisplayOK()
		return nil
	}
  
	noop, warnings, err := cmd.Actor.UpdateManagedServiceInstance(
    string(cmd.RequiredArgs.ServiceInstance),
		cmd.Config.TargetedSpace().GUID,
		v7action.ServiceInstanceUpdateManagedParams{
			Tags:            types.OptionalStringSlice(cmd.Tags),
			Parameters:      types.OptionalObject(cmd.Parameters),
			ServicePlanName: types.OptionalString(cmd.Plan),
		},
	)

	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()
	if noop {
		cmd.UI.DisplayText("No changes were made.")
	}

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
	user, err := cmd.Config.CurrentUser()
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
	return !(cmd.Tags.IsSet || cmd.Parameters.IsSet || cmd.Plan.IsSet)
}
