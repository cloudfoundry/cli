package v7

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/util/ui"
)

type ServiceKeysCommand struct {
	BaseCommand

	RequiredArgs    flag.ServiceInstance `positional-args:"yes"`
	relatedCommands interface{}          `related_commands:"delete-service-key"`
}

func (cmd ServiceKeysCommand) Execute(args []string) error {
	if err := cmd.SharedActor.CheckTarget(true, true); err != nil {
		return err
	}

	if err := cmd.displayIntro(); err != nil {
		return err
	}

	names, warnings, err := cmd.Actor.GetServiceKeysByServiceInstance(
		string(cmd.RequiredArgs.ServiceInstance),
		cmd.Config.TargetedSpace().GUID,
	)
	cmd.UI.DisplayWarnings(warnings)
	switch err.(type) {
	case nil:
		cmd.displayResult(names)
		return nil
	case actionerror.ServiceInstanceTypeError:
		return translatableerror.ServiceKeysNotSupportedWithUserProvidedServiceInstances{}
	default:
		return err
	}
}

func (cmd ServiceKeysCommand) Usage() string {
	return `CF_NAME service-keys SERVICE_INSTANCE`
}

func (cmd ServiceKeysCommand) Examples() string {
	return `CF_NAME service-keys mydb`
}

func (cmd ServiceKeysCommand) displayIntro() error {
	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Getting keys for service instance {{.ServiceInstanceName}} as {{.UserName}}...", map[string]interface{}{
		"ServiceInstanceName": string(cmd.RequiredArgs.ServiceInstance),
		"UserName":            user.Name,
	})
	cmd.UI.DisplayNewline()

	return nil
}

func (cmd ServiceKeysCommand) displayResult(names []string) {
	switch len(names) {
	case 0:
		cmd.UI.DisplayText("No service keys for service instance {{.ServiceInstanceName}}", map[string]interface{}{
			"ServiceInstanceName": string(cmd.RequiredArgs.ServiceInstance),
		})
	default:
		table := [][]string{{"name"}}
		for _, n := range names {
			table = append(table, []string{n})
		}

		cmd.UI.DisplayTableWithHeader("", table, ui.DefaultTableSpacePadding)
	}
}
