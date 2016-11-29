package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type UpdateSecurityGroupCommand struct {
	RequiredArgs    flag.SecurityGroupArgs `positional-args:"yes"`
	usage           interface{}            `usage:"CF_NAME update-security-group SECURITY_GROUP PATH_TO_JSON_RULES_FILE\n\n   The provided path can be an absolute or relative path to a file.\n   It should have a single array with JSON objects inside describing the rules.\n\n   Valid json file example:\n   [\n     {\n       \"protocol\": \"tcp\",\n       \"destination\": \"10.0.11.0/24\",\n       \"ports\": \"80,443\",\n       \"description\": \"Allow http and https traffic from ZoneA\"\n     }\n   ]\n\nTIP: Changes will not apply to existing running applications until they are restarted."`
	relatedCommands interface{}            `related_commands:"restage, security-groups"`
}

func (_ UpdateSecurityGroupCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ UpdateSecurityGroupCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
