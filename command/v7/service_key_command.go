package v7

import (
	"code.cloudfoundry.org/cli/v9/command/flag"
)

type ServiceKeyCommand struct {
	BaseCommand

	RequiredArgs flag.ServiceInstanceKey `positional-args:"yes"`
	GUID         bool                    `long:"guid" description:"Retrieve and display the given service-key's guid. All other output is suppressed."`
	JSON         bool                    `long:"json" description:"Output credentials as JSON. All other output is suppressed."`
}

func (cmd ServiceKeyCommand) Execute(args []string) error {
	if err := cmd.SharedActor.CheckTarget(true, true); err != nil {
		return err
	}

	switch {
	case cmd.GUID:
		return cmd.guid()
	case cmd.JSON:
		return cmd.json()
	default:
		return cmd.details()
	}
}

func (cmd ServiceKeyCommand) Usage() string {
	return `CF_NAME service-key SERVICE_INSTANCE SERVICE_KEY`
}

func (cmd ServiceKeyCommand) Examples() string {
	return `CF_NAME service-key mydb mykey`
}

func (cmd ServiceKeyCommand) guid() error {
	key, warnings, err := cmd.Actor.GetServiceKeyByServiceInstanceAndName(
		cmd.RequiredArgs.ServiceInstance,
		cmd.RequiredArgs.ServiceKey,
		cmd.Config.TargetedSpace().GUID,
	)
	if err != nil {
		cmd.UI.DisplayWarnings(warnings)
		return err
	}

	cmd.UI.DisplayText(key.GUID)
	return nil
}

func (cmd ServiceKeyCommand) json() error {
	details, warnings, err := cmd.Actor.GetServiceKeyDetailsByServiceInstanceAndName(
		cmd.RequiredArgs.ServiceInstance,
		cmd.RequiredArgs.ServiceKey,
		cmd.Config.TargetedSpace().GUID,
	)
	if err != nil {
		cmd.UI.DisplayWarnings(warnings)
		return err
	}

	return cmd.UI.DisplayJSON("", details)
}

func (cmd ServiceKeyCommand) details() error {
	user, err := cmd.Actor.GetCurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Getting key {{.KeyName}} for service instance {{.ServiceInstanceName}} as {{.UserName}}...", map[string]interface{}{
		"ServiceInstanceName": cmd.RequiredArgs.ServiceInstance,
		"KeyName":             cmd.RequiredArgs.ServiceKey,
		"UserName":            user.Name,
	})

	details, warnings, err := cmd.Actor.GetServiceKeyDetailsByServiceInstanceAndName(
		cmd.RequiredArgs.ServiceInstance,
		cmd.RequiredArgs.ServiceKey,
		cmd.Config.TargetedSpace().GUID,
	)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayNewline()

	return cmd.UI.DisplayJSON("", details)
}
