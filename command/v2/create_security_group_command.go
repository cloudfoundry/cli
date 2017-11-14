package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type CreateSecurityGroupCommand struct {
	RequiredArgs    flag.SecurityGroupArgs `positional-args:"yes"`
	usage           interface{}            `usage:"CF_NAME create-security-group SECURITY_GROUP PATH_TO_JSON_RULES_FILE\n\n   The provided path can be an absolute or relative path to a file.  The file should have\n   a single array with JSON objects inside describing the rules.  The JSON Base Object is\n   omitted and only the square brackets and associated child object are required in the file.\n\n   Valid json file example:\n   [\n     {\n       \"protocol\": \"tcp\",\n       \"destination\": \"10.0.11.0/24\",\n       \"ports\": \"80,443\",\n       \"description\": \"Allow http and https traffic from ZoneA\"\n     }\n   ]"`
	relatedCommands interface{}            `related_commands:"bind-security-group, bind-running-security-group, bind-staging-security-group, security-groups"`
}

func (CreateSecurityGroupCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (CreateSecurityGroupCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
