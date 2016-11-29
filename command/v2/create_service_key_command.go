package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type CreateServiceKeyCommand struct {
	RequiredArgs     flag.ServiceInstanceKey `positional-args:"yes"`
	ParametersAsJSON string                  `short:"c" description:"Valid JSON object containing service-specific configuration parameters, provided either in-line or in a file. For a list of supported configuration parameters, see documentation for the particular service offering."`
	usage            interface{}             `usage:"CF_NAME create-service-key SERVICE_INSTANCE SERVICE_KEY [-c PARAMETERS_AS_JSON]\n\n   Optionally provide service-specific configuration parameters in a valid JSON object in-line.\n   CF_NAME create-service-key SERVICE_INSTANCE SERVICE_KEY -c '{\"name\":\"value\",\"name\":\"value\"}'\n\n   Optionally provide a file containing service-specific configuration parameters in a valid JSON object. The path to the parameters file can be an absolute or relative path to a file.\n   CF_NAME create-service-key SERVICE_INSTANCE SERVICE_KEY -c PATH_TO_FILE\n\n   Example of valid JSON object:\n   {\n      \"permissions\": \"read-only\"\n   }\n\nEXAMPLES:\n   CF_NAME create-service-key mydb mykey -c '{\"permissions\":\"read-only\"}'\n   CF_NAME create-service-key mydb mykey -c ~/workspace/tmp/instance_config.json"`
	relatedCommands  interface{}             `related_commands:"service-key"`
}

func (_ CreateServiceKeyCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ CreateServiceKeyCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
