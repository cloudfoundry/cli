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
	OmitApps        bool        `long:"no-apps" description:"Do not retrieve bound apps information."`
	relatedCommands interface{} `related_commands:"create-service, marketplace"`
}

func (cmd ServicesCommand) Execute(args []string) error {
	if err := cmd.SharedActor.CheckTarget(true, true); err != nil {
		return err
	}

	if err := cmd.displayMessage(); err != nil {
		return err
	}

	instances, warnings, err := cmd.Actor.GetServiceInstancesForSpace(cmd.Config.TargetedSpace().GUID, cmd.OmitApps)
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

	table := NewServicesTable(false, cmd.OmitApps)

	for _, si := range instances {
		table.AppendRow(si)
	}
	cmd.UI.DisplayTableWithHeader("", table.table, ui.DefaultTableSpacePadding)
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
		table := NewServicesTable(false, cmd.OmitApps)
		for _, si := range managed {
			table.AppendRow(si)
		}
		cmd.UI.DisplayText("Managed service instances:")
		cmd.UI.DisplayTableWithHeader("    ", table.table, ui.DefaultTableSpacePadding)

		if len(user) > 0 {
			cmd.UI.DisplayNewline()
		}
	}

	if len(user) > 0 {
		table := NewServicesTable(true, cmd.OmitApps)
		for _, si := range user {
			table.AppendRow(si)
			//table = append(table, []string{i.Name, strings.Join(i.BoundApps, ", ")})
		}
		cmd.UI.DisplayText("User-provided service instances:")
		cmd.UI.DisplayTableWithHeader("    ", table.table, ui.DefaultTableSpacePadding)
	}
}

func (cmd ServicesCommand) displayJSON(instances []v7action.ServiceInstance) {
	data, err := json.MarshalIndent(instances, "", "  ")
	if err != nil {
		panic(err)
	}

	cmd.UI.DisplayText(string(data))
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

type ServicesTable struct {
	table    [][]string
	short    bool
	showApps bool
}

func NewServicesTable(short bool, omitApps bool) *ServicesTable {
	t := &ServicesTable{
		short:    short,
		showApps: !omitApps,
	}

	return t.withHeaders()
}

func (t *ServicesTable) withHeaders() *ServicesTable {
	headers := []string{"name"}
	if t.short {
		if t.showApps {
			headers = append(headers, "bound apps")
		}
	} else {
		headers = append(headers, "offering", "plan")
		if t.showApps {
			headers = append(headers, "bound apps")
		}
		headers = append(headers, "last operation", "broker", "upgrade available")
	}
	t.table = [][]string{headers}
	return t
}

func (t *ServicesTable) AppendRow(si v7action.ServiceInstance) {
	row := []string{si.Name}
	if t.short {
		if t.showApps {
			row = append(row, strings.Join(si.BoundApps, ", "))
		}
	} else {
		row = append(row, serviceOfferingName(si), si.ServicePlanName)
		if t.showApps {
			row = append(row, strings.Join(si.BoundApps, ", "))
		}
		row = append(row, si.LastOperation, si.ServiceBrokerName, upgradeAvailableString(si.UpgradeAvailable))
	}
	t.table = append(t.table, row)
}
