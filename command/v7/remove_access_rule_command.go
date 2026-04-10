package v7

import (
	"code.cloudfoundry.org/cli/v9/command/flag"
)

type RemoveAccessRuleCommand struct {
	BaseCommand

	RequiredArgs flag.RemoveAccessRuleArgs `positional-args:"yes"`
	Hostname     string                    `long:"hostname" required:"true" description:"Hostname for the route"`
	Path         string                    `long:"path" description:"Path for the route"`
	Force        bool                      `short:"f" description:"Force deletion without confirmation"`
	usage        interface{}               `usage:"CF_NAME remove-access-rule RULE_NAME DOMAIN --hostname HOSTNAME [--path PATH] [-f]\n\nEXAMPLES:\n   cf remove-access-rule allow-frontend apps.identity --hostname backend\n   cf remove-access-rule allow-monitoring apps.identity --hostname api --path /metrics -f"`
	relatedCommands interface{} `related_commands:"access-rules, add-access-rule"`
}

func (cmd RemoveAccessRuleCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	user, err := cmd.Actor.GetCurrentUser()
	if err != nil {
		return err
	}

	ruleName := cmd.RequiredArgs.RuleName
	domainName := cmd.RequiredArgs.Domain

	if !cmd.Force {
		prompt := "Really remove access rule {{.RuleName}} for route {{.Hostname}}.{{.Domain}}{{.Path}}?"
		response, promptErr := cmd.UI.DisplayBoolPrompt(false, prompt, map[string]interface{}{
			"RuleName": ruleName,
			"Hostname": cmd.Hostname,
			"Domain":   domainName,
			"Path":     formatPath(cmd.Path),
		})

		if promptErr != nil {
			return promptErr
		}

		if !response {
			cmd.UI.DisplayText("Access rule '{{.RuleName}}' has not been removed.", map[string]interface{}{
				"RuleName": ruleName,
			})
			return nil
		}
	}

	cmd.UI.DisplayTextWithFlavor("Removing access rule {{.RuleName}} for route {{.Hostname}}.{{.Domain}}{{.Path}} as {{.User}}...",
		map[string]interface{}{
			"RuleName": ruleName,
			"Hostname": cmd.Hostname,
			"Domain":   domainName,
			"Path":     formatPath(cmd.Path),
			"User":     user.Name,
		})

	warnings, err := cmd.Actor.DeleteAccessRule(ruleName, domainName, cmd.Hostname, cmd.Path)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()

	return nil
}
