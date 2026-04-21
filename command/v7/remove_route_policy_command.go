package v7

import (
	"code.cloudfoundry.org/cli/v9/command/flag"
)

type RemoveRoutePolicyCommand struct {
	BaseCommand

	RequiredArgs    flag.RemoveRoutePolicyArgs `positional-args:"yes"`
	Source          string                     `long:"source" required:"true" description:"Source to identify the route policy (cf:app:<guid>, cf:space:<guid>, cf:org:<guid>, or cf:any)"`
	Hostname        string                     `long:"hostname" required:"true" description:"Hostname for the route"`
	Path            string                     `long:"path" description:"Path for the route"`
	Force           bool                       `short:"f" description:"Force deletion without confirmation"`
	usage           interface{}                `usage:"CF_NAME remove-route-policy DOMAIN --source SOURCE --hostname HOSTNAME [--path PATH] [-f]\n\nEXAMPLES:\n   cf remove-route-policy apps.identity --source cf:app:d76446a1-f429-4444-8797-be2f78b75b08 --hostname backend\n   cf remove-route-policy apps.identity --source cf:space:2b26e210-1b48-4e60-8432-f24bc5927789 --hostname api --path /metrics -f\n   cf remove-route-policy apps.identity --source cf:any --hostname public-api -f"`
	relatedCommands interface{}                `related_commands:"route-policies, add-route-policy"`
}

func (cmd RemoveRoutePolicyCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	user, err := cmd.Actor.GetCurrentUser()
	if err != nil {
		return err
	}

	// Validate source format
	if err := validateSource(cmd.Source); err != nil {
		return err
	}

	domainName := cmd.RequiredArgs.Domain

	if !cmd.Force {
		prompt := "Really remove route policy with source {{.Source}} for route {{.Hostname}}.{{.Domain}}{{.Path}}?"
		response, promptErr := cmd.UI.DisplayBoolPrompt(false, prompt, map[string]interface{}{
			"Source":   cmd.Source,
			"Hostname": cmd.Hostname,
			"Domain":   domainName,
			"Path":     formatPath(cmd.Path),
		})

		if promptErr != nil {
			return promptErr
		}

		if !response {
			cmd.UI.DisplayText("Route policy has not been removed.")
			return nil
		}
	}

	cmd.UI.DisplayTextWithFlavor("Removing route policy for route {{.Hostname}}.{{.Domain}}{{.Path}} as {{.User}}...",
		map[string]interface{}{
			"Hostname": cmd.Hostname,
			"Domain":   domainName,
			"Path":     formatPath(cmd.Path),
			"User":     user.Name,
		})

	warnings, err := cmd.Actor.DeleteRoutePolicyBySource(domainName, cmd.Source, cmd.Hostname, cmd.Path)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()

	return nil
}
