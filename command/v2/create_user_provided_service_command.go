package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type CreateUserProvidedServiceCommand struct {
	RequiredArgs    flag.ServiceInstance `positional-args:"yes"`
	SyslogDrainURL  string               `short:"l" description:"URL to which logs for bound applications will be streamed"`
	Credentials     string               `short:"p" description:"Credentials, provided inline or in a file, to be exposed in the VCAP_SERVICES environment variable for bound applications"`
	RouteServiceURL string               `short:"r" description:"URL to which requests for bound routes will be forwarded. Scheme for this URL must be https"`
	usage           interface{}          `usage:"CF_NAME create-user-provided-service SERVICE_INSTANCE [-p CREDENTIALS] [-l SYSLOG_DRAIN_URL] [-r ROUTE_SERVICE_URL]\n\n   Pass comma separated credential parameter names to enable interactive mode:\n   CF_NAME create-user-provided-service SERVICE_INSTANCE -p \"comma, separated, parameter, names\"\n\n   Pass credential parameters as JSON to create a service non-interactively:\n   CF_NAME create-user-provided-service SERVICE_INSTANCE -p '{\"key1\":\"value1\",\"key2\":\"value2\"}'\n\n   Specify a path to a file containing JSON:\n   CF_NAME create-user-provided-service SERVICE_INSTANCE -p PATH_TO_FILE\n\nEXAMPLES:\n   CF_NAME create-user-provided-service my-db-mine -p \"username, password\"\n   CF_NAME create-user-provided-service my-db-mine -p /path/to/credentials.json\n   CF_NAME create-user-provided-service my-drain-service -l syslog://example.com\n   CF_NAME create-user-provided-service my-route-service -r https://example.com\n\n   Linux/Mac:\n      CF_NAME create-user-provided-service my-db-mine -p '{\"username\":\"admin\",\"password\":\"pa55woRD\"}'\n\n   Windows Command Line:\n      CF_NAME create-user-provided-service my-db-mine -p \"{\\\"username\\\":\\\"admin\\\",\\\"password\\\":\\\"pa55woRD\\\"}\"\n\n   Windows PowerShell:\n      CF_NAME create-user-provided-service my-db-mine -p '{\\\"username\\\":\\\"admin\\\",\\\"password\\\":\\\"pa55woRD\\\"}'"`
	relatedCommands interface{}          `related_commands:"bind-service, services"`
}

func (_ CreateUserProvidedServiceCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ CreateUserProvidedServiceCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
