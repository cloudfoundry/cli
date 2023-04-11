package v7

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
)

type CreateUserProvidedServiceCommand struct {
	BaseCommand

	RequiredArgs    flag.ServiceInstance   `positional-args:"yes"`
	SyslogDrainURL  flag.OptionalString    `short:"l" description:"URL to which logs for bound applications will be streamed"`
	Credentials     flag.CredentialsOrJSON `short:"p" description:"Credentials, provided inline or in a file, to be exposed in the VCAP_SERVICES environment variable for bound applications"`
	RouteServiceURL flag.OptionalString    `short:"r" description:"URL to which requests for bound routes will be forwarded. Scheme for this URL must be https"`
	Tags            flag.Tags              `short:"t" description:"User provided tags"`
	usage           interface{}            `usage:"CF_NAME create-user-provided-service SERVICE_INSTANCE [-p CREDENTIALS] [-l SYSLOG_DRAIN_URL] [-r ROUTE_SERVICE_URL] [-t TAGS]\n\n   Pass comma separated credential parameter names to enable interactive mode:\n   CF_NAME create-user-provided-service SERVICE_INSTANCE -p \"comma, separated, parameter, names\"\n\n   Pass credential parameters as JSON to create a service non-interactively:\n   CF_NAME create-user-provided-service SERVICE_INSTANCE -p '{\"key1\":\"value1\",\"key2\":\"value2\"}'\n\n   Specify a path to a file containing JSON:\n   CF_NAME create-user-provided-service SERVICE_INSTANCE -p PATH_TO_FILE\n\nEXAMPLES:\n   CF_NAME create-user-provided-service my-db-mine -p \"username, password\"\n   CF_NAME create-user-provided-service my-db-mine -p /path/to/credentials.json\n   CF_NAME create-user-provided-service my-db-mine -t \"list, of, tags\"\n   CF_NAME create-user-provided-service my-drain-service -l syslog://example.com\n   CF_NAME create-user-provided-service my-route-service -r https://example.com\n\n   Linux/Mac:\n      CF_NAME create-user-provided-service my-db-mine -p '{\"username\":\"admin\",\"password\":\"pa55woRD\"}'\n\n   Windows Command Line:\n      CF_NAME create-user-provided-service my-db-mine -p \"{\\\"username\\\":\\\"admin\\\",\\\"password\\\":\\\"pa55woRD\\\"}\"\n\n   Windows PowerShell:\n      CF_NAME create-user-provided-service my-db-mine -p '{\\\"username\\\":\\\"admin\\\",\\\"password\\\":\\\"pa55woRD\\\"}'"`
	relatedCommands interface{}            `related_commands:"bind-service, services"`
}

func (cmd CreateUserProvidedServiceCommand) Execute(args []string) error {
	if err := cmd.SharedActor.CheckTarget(true, true); err != nil {
		return err
	}

	if err := promptUserForCredentialsIfRequired(&cmd.Credentials, cmd.UI); err != nil {
		return err
	}

	if err := cmd.displayMessage(); err != nil {
		return err
	}

	serviceInstance := resources.ServiceInstance{
		Name:            string(cmd.RequiredArgs.ServiceInstance),
		SpaceGUID:       cmd.Config.TargetedSpace().GUID,
		Tags:            types.OptionalStringSlice(cmd.Tags),
		SyslogDrainURL:  types.OptionalString(cmd.SyslogDrainURL),
		RouteServiceURL: types.OptionalString(cmd.RouteServiceURL),
		Credentials:     cmd.Credentials.OptionalObject,
	}

	warnings, err := cmd.Actor.CreateUserProvidedServiceInstance(serviceInstance)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()
	return nil
}

func (cmd CreateUserProvidedServiceCommand) displayMessage() error {
	user, err := cmd.Actor.GetCurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Creating user provided service {{.ServiceInstance}} in org {{.Org}} / space {{.Space}} as {{.User}}...",
		map[string]interface{}{
			"ServiceInstance": cmd.RequiredArgs.ServiceInstance,
			"Org":             cmd.Config.TargetedOrganization().Name,
			"Space":           cmd.Config.TargetedSpace().Name,
			"User":            user.Name,
		},
	)

	return nil
}

func promptUserForCredentialsIfRequired(flag *flag.CredentialsOrJSON, ui command.UI) error {
	if len(flag.UserPromptCredentials) > 0 {
		flag.Value = make(map[string]interface{})

		for _, key := range flag.UserPromptCredentials {
			var err error
			flag.Value[key], err = ui.DisplayPasswordPrompt(key)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
