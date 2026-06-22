package v7

import (
	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/v9/command"
	"code.cloudfoundry.org/cli/v9/command/flag"
)

type RemoveRoutePolicyCommand struct {
	BaseCommand

	RequiredArgs flag.RoutePolicyArgs `positional-args:"yes"`
	RoutePolicySourceFlags
	Hostname string `long:"hostname" short:"n" required:"true" description:"Hostname for the route"`
	Path     string `long:"path" description:"Path for the route"`
	Force    bool   `short:"f" description:"Force deletion without confirmation"`

	usage           interface{} `usage:"CF_NAME remove-route-policy DOMAIN --hostname HOSTNAME [--source-app APP_NAME [--source-space SPACE_NAME] [--source-org ORG_NAME] | --source-space SPACE_NAME [--source-org ORG_NAME] | --source-org ORG_NAME | --source-any | --source SOURCE] [--path PATH] [-f]\n\nEXAMPLES:\n   # Remove by app name (mirrors add-route-policy)\n   cf remove-route-policy apps.identity --source-app frontend-app --hostname backend\n\n   # Remove by app in a different space\n   cf remove-route-policy apps.identity --source-app api-client --source-space other-space --hostname backend\n\n   # Remove a space-level policy\n   cf remove-route-policy apps.identity --source-space monitoring --hostname api --path /metrics -f\n\n   # Remove an org-level policy\n   cf remove-route-policy apps.identity --source-org platform --hostname shared-api -f\n\n   # Remove using raw source (advanced)\n   cf remove-route-policy apps.identity --source cf:app:d76446a1-f429-4444-8797-be2f78b75b08 --hostname backend\n   cf remove-route-policy apps.identity --source cf:any --hostname public-api -f"`
	relatedCommands interface{} `related_commands:"route-policies, add-route-policy"`
}

func (cmd RemoveRoutePolicyCommand) Execute(args []string) error {
	if err := command.MinimumCCAPIVersionCheck(cmd.Config.APIVersion(), ccversion.MinVersionRoutePolicies); err != nil {
		return err
	}

	if err := cmd.RoutePolicySourceFlags.validateSourceFlags(); err != nil {
		return err
	}

	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	user, err := cmd.Actor.GetCurrentUser()
	if err != nil {
		return err
	}

	source, _, warnings, err := resolveSource(cmd.RoutePolicySourceFlags, cmd.Actor, cmd.Config)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	if err := validateSource(source); err != nil {
		return err
	}

	domainName := cmd.RequiredArgs.Domain

	if !cmd.Force {
		prompt := "Really remove route policy with source {{.Source}} for route {{.Hostname}}.{{.Domain}}{{.Path}}?"
		response, promptErr := cmd.UI.DisplayBoolPrompt(false, prompt, map[string]interface{}{
			"Source":   source,
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

	warnings, err = cmd.Actor.DeleteRoutePolicyBySource(domainName, source, cmd.Hostname, cmd.Path)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()

	return nil
}
