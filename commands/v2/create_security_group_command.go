package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
	"code.cloudfoundry.org/cli/commands/flags"
)

type CreateSecurityGroupCommand struct {
	RequiredArgs    flags.SecurityGroupArgs `positional-args:"yes"`
	usage           interface{}             `usage:"CF_NAME create-security-group SECURITY_GROUP PATH_TO_JSON_RULES_FILE\n\n   The provided path can be an absolute or relative path to a file.  The file should have\n   a single array with JSON objects inside describing the rules.  The JSON Base Object is\n   omitted and only the square brackets and associated child object are required in the file.\n\n   Valid json file example:\n   [\n     {\n       \"protocol\": \"tcp\",\n       \"destination\": \"10.244.1.18\",\n       \"ports\": \"3306\"\n     }\n   ]"`
	relatedCommands interface{}             `related_commands:"bind-security-group, bind-running-security-group, bind-staging-security-group, security-groups"`
}

func (_ CreateSecurityGroupCommand) Setup(config commands.Config, ui commands.UI) error {
	return nil
}

func (_ CreateSecurityGroupCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
