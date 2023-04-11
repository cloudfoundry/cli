package v7

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
)

type UpdateUserProvidedServiceCommand struct {
	BaseCommand

	RequiredArgs    flag.ServiceInstance   `positional-args:"yes"`
	SyslogDrainURL  flag.OptionalString    `short:"l" description:"URL to which logs for bound applications will be streamed"`
	Credentials     flag.CredentialsOrJSON `short:"p" description:"Credentials, provided inline or in a file, to be exposed in the VCAP_SERVICES environment variable for bound applications. Provided credentials will override existing credentials."`
	RouteServiceURL flag.OptionalString    `short:"r" description:"URL to which requests for bound routes will be forwarded. Scheme for this URL must be https"`
	Tags            flag.Tags              `short:"t" description:"User provided tags"`
	usage           interface{}            `usage:"CF_NAME update-user-provided-service SERVICE_INSTANCE [-p CREDENTIALS] [-l SYSLOG_DRAIN_URL] [-r ROUTE_SERVICE_URL] [-t TAGS]\n\n   Pass comma separated credential parameter names to enable interactive mode:\n   CF_NAME update-user-provided-service SERVICE_INSTANCE -p \"comma, separated, parameter, names\"\n\n   Pass credential parameters as JSON to create a service non-interactively:\n   CF_NAME update-user-provided-service SERVICE_INSTANCE -p '{\"key1\":\"value1\",\"key2\":\"value2\"}'\n\n   Specify a path to a file containing JSON:\n   CF_NAME update-user-provided-service SERVICE_INSTANCE -p PATH_TO_FILE\n\nEXAMPLES:\n   CF_NAME update-user-provided-service my-db-mine -p '{\"username\":\"admin\", \"password\":\"pa55woRD\"}'\n   CF_NAME update-user-provided-service my-db-mine -p /path/to/credentials.json\n   CF_NAME create-user-provided-service my-db-mine -t \"list, of, tags\"\n   CF_NAME update-user-provided-service my-drain-service -l syslog://example.com\n   CF_NAME update-user-provided-service my-route-service -r https://example.com"`
	relatedCommands interface{}            `related_commands:"rename-service, services, update-service"`
}

func (cmd *UpdateUserProvidedServiceCommand) Execute(args []string) error {
	if err := cmd.SharedActor.CheckTarget(true, true); err != nil {
		return err
	}

	if err := promptUserForCredentialsIfRequired(&cmd.Credentials, cmd.UI); err != nil {
		return err
	}

	if err := cmd.displayMessage(); err != nil {
		return err
	}

	if cmd.noFlagsProvided() {
		cmd.UI.DisplayText("No flags specified. No changes were made.")
		cmd.UI.DisplayOK()
		return nil
	}

	serviceInstanceName := string(cmd.RequiredArgs.ServiceInstance)

	warnings, err := cmd.Actor.UpdateUserProvidedServiceInstance(serviceInstanceName, cmd.Config.TargetedSpace().GUID, resources.ServiceInstance{
		Tags:            types.OptionalStringSlice(cmd.Tags),
		SyslogDrainURL:  types.OptionalString(cmd.SyslogDrainURL),
		RouteServiceURL: types.OptionalString(cmd.RouteServiceURL),
		Credentials:     cmd.Credentials.OptionalObject,
	})
	cmd.UI.DisplayWarnings(warnings)
	switch err.(type) {
	case ccerror.ServiceInstanceNotFoundError:
		return actionerror.ServiceInstanceNotFoundError{Name: serviceInstanceName}
	default:
		return err
	case nil:
	}

	cmd.UI.DisplayOK()
	cmd.UI.DisplayText("TIP: Use 'cf restage' for any bound apps to ensure your env variable changes take effect")

	return nil
}

func (cmd UpdateUserProvidedServiceCommand) displayMessage() error {
	user, err := cmd.Actor.GetCurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Updating user provided service {{.ServiceInstance}} in org {{.Org}} / space {{.Space}} as {{.User}}...",
		map[string]interface{}{
			"ServiceInstance": cmd.RequiredArgs.ServiceInstance,
			"Org":             cmd.Config.TargetedOrganization().Name,
			"Space":           cmd.Config.TargetedSpace().Name,
			"User":            user.Name,
		},
	)

	return nil
}

func (cmd *UpdateUserProvidedServiceCommand) noFlagsProvided() bool {
	return !cmd.SyslogDrainURL.IsSet && !cmd.RouteServiceURL.IsSet && !cmd.Tags.IsSet && !cmd.Credentials.IsSet
}
