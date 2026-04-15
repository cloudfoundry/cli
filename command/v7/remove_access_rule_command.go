package v7

import (
	"code.cloudfoundry.org/cli/v9/command/flag"
)

type RemoveAccessRuleCommand struct {
	BaseCommand

	RequiredArgs flag.RemoveAccessRuleArgs `positional-args:"yes"`
	Selector     string                    `long:"selector" required:"true" description:"Selector to identify the access rule (cf:app:<guid>, cf:space:<guid>, cf:org:<guid>, or cf:any)"`
	Hostname     string                    `long:"hostname" required:"true" description:"Hostname for the route"`
	Path         string                    `long:"path" description:"Path for the route"`
	Force        bool                      `short:"f" description:"Force deletion without confirmation"`
	usage        interface{}               `usage:"CF_NAME remove-access-rule DOMAIN --selector SELECTOR --hostname HOSTNAME [--path PATH] [-f]\n\nEXAMPLES:\n   cf remove-access-rule apps.identity --selector cf:app:d76446a1-f429-4444-8797-be2f78b75b08 --hostname backend\n   cf remove-access-rule apps.identity --selector cf:space:2b26e210-1b48-4e60-8432-f24bc5927789 --hostname api --path /metrics -f\n   cf remove-access-rule apps.identity --selector cf:any --hostname public-api -f"`
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

	// Validate selector format
	if err := validateSelector(cmd.Selector); err != nil {
		return err
	}

	domainName := cmd.RequiredArgs.Domain

	if !cmd.Force {
		prompt := "Really remove access rule with selector {{.Selector}} for route {{.Hostname}}.{{.Domain}}{{.Path}}?"
		response, promptErr := cmd.UI.DisplayBoolPrompt(false, prompt, map[string]interface{}{
			"Selector": cmd.Selector,
			"Hostname": cmd.Hostname,
			"Domain":   domainName,
			"Path":     formatPath(cmd.Path),
		})

		if promptErr != nil {
			return promptErr
		}

		if !response {
			cmd.UI.DisplayText("Access rule has not been removed.")
			return nil
		}
	}

	cmd.UI.DisplayTextWithFlavor("Removing access rule for route {{.Hostname}}.{{.Domain}}{{.Path}} as {{.User}}...",
		map[string]interface{}{
			"Hostname": cmd.Hostname,
			"Domain":   domainName,
			"Path":     formatPath(cmd.Path),
			"User":     user.Name,
		})

	warnings, err := cmd.Actor.DeleteAccessRuleBySelector(domainName, cmd.Selector, cmd.Hostname, cmd.Path)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()

	return nil
}
