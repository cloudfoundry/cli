package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type CreateServiceCommand struct {
	RequiredArgs      flag.CreateServiceArgs `positional-args:"yes"`
	ConfigurationFile flag.Path              `short:"c" description:"Valid JSON object containing service-specific configuration parameters, provided either in-line or in a file. For a list of supported configuration parameters, see documentation for the particular service offering."`
	Tags              string                 `short:"t" description:"User provided tags"`
	usage             interface{}            `usage:"CF_NAME create-service SERVICE PLAN SERVICE_INSTANCE [-c PARAMETERS_AS_JSON] [-t TAGS]\n\n   Optionally provide service-specific configuration parameters in a valid JSON object in-line:\n\n   CF_NAME create-service SERVICE PLAN SERVICE_INSTANCE -c '{\"name\":\"value\",\"name\":\"value\"}'\n\n   Optionally provide a file containing service-specific configuration parameters in a valid JSON object.\n   The path to the parameters file can be an absolute or relative path to a file:\n\n   CF_NAME create-service SERVICE PLAN SERVICE_INSTANCE -c PATH_TO_FILE\n\n   Example of valid JSON object:\n   {\n      \"cluster_nodes\": {\n         \"count\": 5,\n         \"memory_mb\": 1024\n      }\n   }\n\nTIP:\n   Use 'CF_NAME create-user-provided-service' to make user-provided services available to CF apps\n\nEXAMPLES:\n   Linux/Mac:\n      CF_NAME create-service db-service silver mydb -c '{\"ram_gb\":4}'\n\n   Windows Command Line:\n      CF_NAME create-service db-service silver mydb -c \"{\\\"ram_gb\\\":4}\"\n\n   Windows PowerShell:\n      CF_NAME create-service db-service silver mydb -c '{\\\"ram_gb\\\":4}'\n\n   CF_NAME create-service db-service silver mydb -c ~/workspace/tmp/instance_config.json\n\n   CF_NAME create-service db-service silver mydb -t \"list, of, tags\""`
	relatedCommands   interface{}            `related_commands:"bind-service, create-user-provided-service, marketplace, services"`
}

func (CreateServiceCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (CreateServiceCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
