package v7

import (
	"encoding/json"
	"strings"

	"code.cloudfoundry.org/cli/resources"

	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/ui"
)

type ServicesCommand struct {
	BaseCommand

	Format          string      `long:"format" hidden:"yes"`
	relatedCommands interface{} `related_commands:"create-service, marketplace"`
}

func (cmd ServicesCommand) Execute(args []string) error {
	if err := cmd.SharedActor.CheckTarget(true, true); err != nil {
		return err
	}

	if err := cmd.displayMessage(); err != nil {
		return err
	}

	instances, warnings, err := cmd.Actor.GetServiceInstancesForSpace(cmd.Config.TargetedSpace().GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	switch cmd.Format {
	case "json":
		cmd.displayJSON(instances)
	case "split":
		cmd.displaySplitTable(instances)
	default:
		cmd.displayHeritageTable(instances)
	}

	return nil
}

func (cmd ServicesCommand) Usage() string {
	return "CF_NAME services"
}

func (cmd ServicesCommand) displayMessage() error {
	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Getting service instances in org {{.OrgName}} / space {{.SpaceName}} as {{.UserName}}...", map[string]interface{}{
		"OrgName":   cmd.Config.TargetedOrganization().Name,
		"SpaceName": cmd.Config.TargetedSpace().Name,
		"UserName":  user.Name,
	})
	cmd.UI.DisplayNewline()

	return nil
}

func (cmd ServicesCommand) displayHeritageTable(instances []v7action.ServiceInstance) {
	if len(instances) == 0 {
		cmd.UI.DisplayText("No service instances found.")
		return
	}

	table := [][]string{{"name", "offering", "plan", "bound apps", "last operation", "broker", "upgrade available"}}
	for _, i := range instances {
		table = append(table, heritageTableLine(i))
	}
	cmd.UI.DisplayTableWithHeader("", table, ui.DefaultTableSpacePadding)
}

func (cmd ServicesCommand) displaySplitTable(instances []v7action.ServiceInstance) {
	if len(instances) == 0 {
		cmd.UI.DisplayText("No service instances found.")
		return
	}

	var managed, user []v7action.ServiceInstance
	for _, i := range instances {
		if i.Type == resources.UserProvidedServiceInstance {
			user = append(user, i)
		} else {
			managed = append(managed, i)
		}
	}

	if len(managed) > 0 {
		table := [][]string{{"name", "offering", "plan", "bound apps", "last operation", "broker", "upgrade available"}}
		for _, i := range managed {
			table = append(table, heritageTableLine(i))
		}
		cmd.UI.DisplayText("Managed service instances:")
		cmd.UI.DisplayTableWithHeader("    ", table, ui.DefaultTableSpacePadding)

		if len(user) > 0 {
			cmd.UI.DisplayNewline()
		}
	}

	if len(user) > 0 {
		table := [][]string{{"name", "bound apps"}}
		for _, i := range user {
			table = append(table, []string{i.Name, strings.Join(i.BoundApps, ", ")})
		}
		cmd.UI.DisplayText("User-provided service instances:")
		cmd.UI.DisplayTableWithHeader("    ", table, ui.DefaultTableSpacePadding)
	}
}

func (cmd ServicesCommand) displayJSON(instances []v7action.ServiceInstance) {
	data, err := json.MarshalIndent(instances, "", "  ")
	if err != nil {
		panic(err)
	}

	cmd.UI.DisplayText(string(data))
}

func heritageTableLine(si v7action.ServiceInstance) []string {
	return []string{
		si.Name,
		serviceOfferingName(si),
		si.ServicePlanName,
		strings.Join(si.BoundApps, ", "),
		si.LastOperation,
		si.ServiceBrokerName,
		upgradeAvailableString(si.UpgradeAvailable),
	}
}

func upgradeAvailableString(u types.OptionalBoolean) string {
	switch {
	case u.IsSet && u.Value:
		return "yes"
	case u.IsSet:
		return "no"
	default:
		return ""
	}
}

func serviceOfferingName(si v7action.ServiceInstance) string {
	if si.Type == resources.UserProvidedServiceInstance {
		return "user-provided"
	}
	return si.ServiceOfferingName
}
