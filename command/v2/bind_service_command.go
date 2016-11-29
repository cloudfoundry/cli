package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type BindServiceCommand struct {
	RequiredArgs     flag.BindServiceArgs `positional-args:"yes"`
	ParametersAsJSON string               `short:"c" description:"Valid JSON object containing service-specific configuration parameters, provided either in-line or in a file. For a list of supported configuration parameters, see documentation for the particular service offering."`
	usage            interface{}          `usage:"CF_NAME bind-service APP_NAME SERVICE_INSTANCE [-c PARAMETERS_AS_JSON]\n\n   Optionally provide service-specific configuration parameters in a valid JSON object in-line:\n\n   CF_NAME bind-service APP_NAME SERVICE_INSTANCE -c '{\"name\":\"value\",\"name\":\"value\"}'\n\n   Optionally provide a file containing service-specific configuration parameters in a valid JSON object. \n   The path to the parameters file can be an absolute or relative path to a file.\n   CF_NAME bind-service APP_NAME SERVICE_INSTANCE -c PATH_TO_FILE\n\n   Example of valid JSON object:\n   {\n      \"permissions\": \"read-only\"\n   }\n\nEXAMPLES:\n   Linux/Mac:\n      CF_NAME bind-service myapp mydb -c '{\"permissions\":\"read-only\"}'\n\n   Windows Command Line:\n      CF_NAME bind-service myapp mydb -c \"{\\\"permissions\\\":\\\"read-only\\\"}\"\n\n   Windows PowerShell:\n      CF_NAME bind-service myapp mydb -c '{\\\"permissions\\\":\\\"read-only\\\"}'\n\n   CF_NAME bind-service myapp mydb -c ~/workspace/tmp/instance_config.json"`
	relatedCommands  interface{}          `related_commands:"services"`
}

func (_ BindServiceCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ BindServiceCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
