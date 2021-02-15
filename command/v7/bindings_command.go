package v7

import (
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/util/ui"
)

type BindingsCommand struct {
	BaseCommand

	RequiredArgs    flag.ServiceInstance `positional-args:"yes"`
	relatedCommands interface{}          `related_commands:"service-keys"`
}

func (cmd BindingsCommand) Execute(args []string) error {
	if err := cmd.SharedActor.CheckTarget(true, true); err != nil {
		return err
	}

	if err := cmd.displayIntro(); err != nil {
		return err
	}

	bindings, warnings, err := cmd.Actor.GetBindingsByServiceInstance(v7action.BindingListParameters{
		ServiceInstanceName: string(cmd.RequiredArgs.ServiceInstance),
		SpaceGUID:           cmd.Config.TargetedSpace().GUID,
		GetAppBindings:      true,
		GetServiceKeys:      true,
		GetRouteBindings:    true,
	})
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.displayAppBindings(bindings.App)
	cmd.displayServiceKeys(bindings.Key)
	cmd.displayRouteBindings(bindings.Route)
	return nil
}

func (cmd BindingsCommand) Usage() string {
	return `CF_NAME bindings SERVICE_INSTANCE`
}

func (cmd BindingsCommand) Examples() string {
	return `CF_NAME bindings mydb`
}

func (cmd BindingsCommand) displayIntro() error {
	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Getting bindings for service instance {{.ServiceInstanceName}} as {{.UserName}}...", map[string]interface{}{
		"ServiceInstanceName": string(cmd.RequiredArgs.ServiceInstance),
		"UserName":            user.Name,
	})
	cmd.UI.DisplayNewline()

	return nil
}

func (cmd BindingsCommand) displayAppBindings(bindings []resources.ServiceCredentialBinding) {
	table := [][]string{{"app name", "binding name", "last operation"}}
	for _, b := range bindings {
		table = append(table, []string{b.AppName, b.Name, lastOperation(b.LastOperation)})
	}

	cmd.displayTable("App bindings", table)
}

func (cmd BindingsCommand) displayServiceKeys(keys []resources.ServiceCredentialBinding) {
	table := [][]string{{"name", "last operation"}}
	for _, k := range keys {
		table = append(table, []string{k.Name, lastOperation(k.LastOperation)})
	}

	cmd.displayTable("Service keys", table)
}

func (cmd BindingsCommand) displayRouteBindings(bindings []v7action.RouteBindingSummary) {
	table := [][]string{{"URL", "last operation"}}
	for _, b := range bindings {
		table = append(table, []string{b.URL, lastOperation(b.LastOperation)})
	}

	cmd.displayTable("Route bindings", table)
}

func (cmd BindingsCommand) displayTable(heading string, table [][]string) {
	cmd.UI.DisplayText("{{.Heading}}:", map[string]interface{}{"Heading": heading})

	switch len(table) {
	case 1:
		cmd.UI.DisplayText("  none")
	default:
		cmd.UI.DisplayTableWithHeader("  ", table, ui.DefaultTableSpacePadding)
	}

	cmd.UI.DisplayNewline()
}
