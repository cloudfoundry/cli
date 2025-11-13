package v7

import (
	"fmt"

	"code.cloudfoundry.org/cli/v8/command/flag"
	"code.cloudfoundry.org/cli/v8/resources"
	"code.cloudfoundry.org/cli/v8/util/ui"
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

	keys, warnings, err := cmd.Actor.GetServiceKeysByServiceInstance(
		string(cmd.RequiredArgs.ServiceInstance),
		cmd.Config.TargetedSpace().GUID,
	)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	switch len(keys) {
	case 0:
		cmd.displayEmptyResult()
	default:
		cmd.displayKeysTable(keys)
	}
	return nil
}

func (cmd ServiceKeysCommand) Usage() string {
	return `CF_NAME service-keys SERVICE_INSTANCE`
}

func (cmd ServiceKeysCommand) Examples() string {
	return `CF_NAME service-keys mydb`
}

func (cmd ServiceKeysCommand) displayIntro() error {
	user, err := cmd.Actor.GetCurrentUser()
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

func (cmd ServiceKeysCommand) displayEmptyResult() {
	cmd.UI.DisplayText("No service keys for service instance {{.ServiceInstanceName}}", map[string]interface{}{
		"ServiceInstanceName": string(cmd.RequiredArgs.ServiceInstance),
	})
}

func (cmd ServiceKeysCommand) displayKeysTable(keys []resources.ServiceCredentialBinding) {
	table := [][]string{{"name", "last operation", "message"}}
	for _, k := range keys {
		table = append(table, []string{k.Name, lastOperation(k.LastOperation), k.LastOperation.Description})
	}

	cmd.UI.DisplayTableWithHeader("", table, ui.DefaultTableSpacePadding)
}

func lastOperation(lo resources.LastOperation) string {
	if lo.Type != "" && lo.State != "" {
		return fmt.Sprintf("%s %s", lo.Type, lo.State)
	}
	return ""
}
