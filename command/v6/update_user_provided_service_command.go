package v6

import (
	"fmt"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v6/shared"
)

//go:generate counterfeiter . UpdateUserProvidedServiceActor

type UpdateUserProvidedServiceActor interface {
	GetServiceInstanceByNameAndSpace(name string, spaceGUID string) (v2action.ServiceInstance, v2action.Warnings, error)
	UpdateUserProvidedServiceInstance(guid string, instance v2action.UserProvidedServiceInstance) (v2action.Warnings, error)
}

type UpdateUserProvidedServiceCommand struct {
	RequiredArgs    flag.ServiceInstance   `positional-args:"yes"`
	SyslogDrainURL  flag.OptionalString    `short:"l" description:"URL to which logs for bound applications will be streamed"`
	Credentials     flag.CredentialsOrJSON `short:"p" description:"Credentials, provided inline or in a file, to be exposed in the VCAP_SERVICES environment variable for bound applications. Provided credentials will override existing credentials."`
	RouteServiceURL flag.OptionalString    `short:"r" description:"URL to which requests for bound routes will be forwarded. Scheme for this URL must be https"`
	Tags            flag.Tags              `short:"t" description:"User provided tags"`
	usage           interface{}            `usage:"CF_NAME update-user-provided-service SERVICE_INSTANCE [-p CREDENTIALS] [-l SYSLOG_DRAIN_URL] [-r ROUTE_SERVICE_URL] [-t TAGS]\n\n   Pass comma separated credential parameter names to enable interactive mode:\n   CF_NAME update-user-provided-service SERVICE_INSTANCE -p \"comma, separated, parameter, names\"\n\n   Pass credential parameters as JSON to create a service non-interactively:\n   CF_NAME update-user-provided-service SERVICE_INSTANCE -p '{\"key1\":\"value1\",\"key2\":\"value2\"}'\n\n   Specify a path to a file containing JSON:\n   CF_NAME update-user-provided-service SERVICE_INSTANCE -p PATH_TO_FILE\n\nEXAMPLES:\n   CF_NAME update-user-provided-service my-db-mine -p '{\"username\":\"admin\", \"password\":\"pa55woRD\"}'\n   CF_NAME update-user-provided-service my-db-mine -p /path/to/credentials.json\n   CF_NAME create-user-provided-service my-db-mine -t \"list, of, tags\"\n   CF_NAME update-user-provided-service my-drain-service -l syslog://example.com\n   CF_NAME update-user-provided-service my-route-service -r https://example.com"`
	relatedCommands interface{}            `related_commands:"rename-service, services, update-service"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       UpdateUserProvidedServiceActor
}

func (cmd *UpdateUserProvidedServiceCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor(config)

	ccClient, uaaClient, err := shared.NewClients(config, ui, true)
	if err != nil {
		return err
	}
	cmd.Actor = v2action.NewActor(ccClient, uaaClient, config)

	return nil
}

func (cmd *UpdateUserProvidedServiceCommand) Execute(args []string) error {
	if len(args) > 0 {
		return translatableerror.TooManyArgumentsError{
			ExtraArgument: args[0],
		}
	}

	if err := cmd.SharedActor.CheckTarget(true, true); err != nil {
		return err
	}

	serviceInstance, err := cmd.findServiceInstance()
	if err != nil {
		return err
	}

	if ok := cmd.atLeastOneFlagProvided(); !ok {
		return nil
	}

	if err := cmd.maybePromptUserForCredentials(); err != nil {
		return err
	}

	if err := cmd.displayPreUpdateMessage(); err != nil {
		return err
	}

	warnings, err := cmd.Actor.UpdateUserProvidedServiceInstance(serviceInstance.GUID, cmd.prepareUpdate())
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()
	cmd.UI.DisplayText("TIP: Use 'cf restage' for any bound apps to ensure your env variable changes take effect")

	return nil
}

func (cmd *UpdateUserProvidedServiceCommand) findServiceInstance() (v2action.ServiceInstance, error) {
	serviceInstance, warnings, err := cmd.Actor.GetServiceInstanceByNameAndSpace(cmd.RequiredArgs.ServiceInstance, cmd.Config.TargetedSpace().GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return v2action.ServiceInstance{}, err
	}

	if !serviceInstance.IsUserProvided() {
		return v2action.ServiceInstance{}, fmt.Errorf("The service instance '%s' is not user-provided", cmd.RequiredArgs.ServiceInstance)
	}

	return serviceInstance, nil
}

func (cmd *UpdateUserProvidedServiceCommand) displayPreUpdateMessage() error {
	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Updating user provided service {{.ServiceInstance}} in org {{.Org}} / space {{.Space}} as {{.User}}...",
		map[string]interface{}{
			"ServiceInstance": cmd.RequiredArgs.ServiceInstance,
			"Org":             cmd.Config.TargetedOrganization().Name,
			"Space":           cmd.Config.TargetedSpace().Name,
			"User":            user.Name,
		})

	return nil
}

func (cmd *UpdateUserProvidedServiceCommand) prepareUpdate() v2action.UserProvidedServiceInstance {
	var instanceChanges v2action.UserProvidedServiceInstance
	if cmd.Tags.IsSet {
		instanceChanges = instanceChanges.WithTags(cmd.Tags.Value)
	}
	if cmd.RouteServiceURL.IsSet {
		instanceChanges = instanceChanges.WithRouteServiceURL(cmd.RouteServiceURL.Value)
	}
	if cmd.SyslogDrainURL.IsSet {
		instanceChanges = instanceChanges.WithSyslogDrainURL(cmd.SyslogDrainURL.Value)
	}
	if cmd.Credentials.IsSet {
		instanceChanges = instanceChanges.WithCredentials(cmd.Credentials.Value)
	}

	return instanceChanges
}

func (cmd *UpdateUserProvidedServiceCommand) atLeastOneFlagProvided() bool {
	if !cmd.SyslogDrainURL.IsSet && !cmd.RouteServiceURL.IsSet && !cmd.Tags.IsSet && !cmd.Credentials.IsSet {
		cmd.UI.DisplayText("No flags specified. No changes were made.")
		cmd.UI.DisplayOK()
		return false
	}

	return true
}

func (cmd *UpdateUserProvidedServiceCommand) maybePromptUserForCredentials() error {
	if len(cmd.Credentials.UserPromptCredentials) > 0 {
		cmd.Credentials.Value = make(map[string]interface{})

		for _, key := range cmd.Credentials.UserPromptCredentials {
			var err error
			cmd.Credentials.Value[key], err = cmd.UI.DisplayPasswordPrompt(key)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
