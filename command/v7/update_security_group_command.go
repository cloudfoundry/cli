package v7

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/command/flag"
)

type UpdateSecurityGroupCommand struct {
	BaseCommand

	RequiredArgs    flag.SecurityGroupArgs `positional-args:"yes"`
	usage           interface{}            `usage:"CF_NAME update-security-group SECURITY_GROUP PATH_TO_JSON_RULES_FILE\n\n   The provided path can be an absolute or relative path to a file. The file should have\n   a single array with JSON objects inside describing the rules. The JSON Base Object is\n   omitted and only the square brackets and associated child object are required in the file.\n\n   Valid json file example:\n   [\n     {\n       \"protocol\": \"tcp\",\n       \"destination\": \"10.0.11.0/24\",\n       \"ports\": \"80,443\",\n       \"description\": \"Allow http and https traffic from ZoneA\"\n     }\n   ]\n\nTIP: Changes require an app restart (for running) or restage (for staging) to apply to existing applications."`
	relatedCommands interface{}            `related_commands:"restage, security-groups"`
}

func (cmd UpdateSecurityGroupCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Updating security group {{.Name}} as {{.Username}}...", map[string]interface{}{
		"Name":     cmd.RequiredArgs.SecurityGroup,
		"Username": user.Name,
	})
	cmd.UI.DisplayNewline()

	warnings, err := cmd.Actor.UpdateSecurityGroup(cmd.RequiredArgs.SecurityGroup, string(cmd.RequiredArgs.PathToJSONRules))
	cmd.UI.DisplayWarnings(warnings)
	if _, ok := err.(*json.SyntaxError); ok {
		return actionerror.SecurityGroupJsonSyntaxError{Path: string(cmd.RequiredArgs.PathToJSONRules)}
	}

	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()
	cmd.UI.DisplayNewline()
	cmd.UI.DisplayText("TIP: Changes require an app restart (for running) or restage (for staging) to apply to existing applications.")
	return nil
}
