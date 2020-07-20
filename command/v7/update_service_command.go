package v7

import (
	"strings"

	"code.cloudfoundry.org/cli/command/flag"
)

type UpdateServiceCommand struct {
	BaseCommand

	RequiredArgs     flag.ServiceInstance `positional-args:"yes"`
	ParametersAsJSON flag.Path            `short:"c" description:"Valid JSON object containing service-specific configuration parameters, provided either in-line or in a file. For a list of supported configuration parameters, see documentation for the particular service offering."`
	Plan             flag.OptionalString  `short:"p" description:"Change service plan for a service instance"`
	Tags             flag.OptionalString  `short:"t" description:"User provided tags"`
	Upgrade          bool                 `short:"u" long:"upgrade" description:"Upgrade the service instance to the latest version of the service plan available. It cannot be combined with flags: -c, -p, -t."`
	ForceUpgrade     bool                 `short:"f" long:"force" description:"Force the upgrade to the latest available version of the service plan. It can only be used with: -u, --upgrade."`

	relatedCommands interface{} `related_commands:"rename-service, services, update-user-provided-service"`
}

func (cmd UpdateServiceCommand) Execute(args []string) error {
	return cmd.SharedActor.CheckTarget(true, true)
}

func (cmd UpdateServiceCommand) Usage() string {
	return strings.TrimSpace(`
CF_NAME update-service SERVICE_INSTANCE [-p NEW_PLAN] [-c PARAMETERS_AS_JSON] [-t TAGS] [--upgrade]

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
CF_NAME update-service mydb --upgrade
CF_NAME update-service mydb --upgrade --force
`,
	)
}
