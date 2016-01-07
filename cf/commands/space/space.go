package space

import (
	"fmt"
	"strings"

	"github.com/cloudfoundry/cli/cf/command_registry"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/flags"
	"github.com/cloudfoundry/cli/flags/flag"
	"github.com/cloudfoundry/cli/plugin/models"

	"github.com/cloudfoundry/cli/cf/api/space_quotas"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/formatters"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type ShowSpace struct {
	ui          terminal.UI
	config      core_config.Reader
	spaceReq    requirements.SpaceRequirement
	quotaRepo   space_quotas.SpaceQuotaRepository
	pluginModel *plugin_models.GetSpace_Model
	pluginCall  bool
}

func init() {
	command_registry.Register(&ShowSpace{})
}

func (cmd *ShowSpace) MetaData() command_registry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["guid"] = &cliFlags.BoolFlag{Name: "guid", Usage: T("Retrieve and display the given space's guid.  All other output for the space is suppressed.")}
	fs["security-group-rules"] = &cliFlags.BoolFlag{Name: "security-group-rules", Usage: T("Retrieve the rules for all the security groups associated with the space")}
	return command_registry.CommandMetadata{
		Name:        "space",
		Description: T("Show space info"),
		Usage:       T("CF_NAME space SPACE"),
		Flags:       fs,
	}
}

func (cmd *ShowSpace) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + command_registry.Commands.CommandUsage("space"))
	}

	cmd.spaceReq = requirementsFactory.NewSpaceRequirement(fc.Args()[0])
	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedOrgRequirement(),
		cmd.spaceReq,
	}
	return
}

func (cmd *ShowSpace) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.quotaRepo = deps.RepoLocator.GetSpaceQuotaRepository()
	cmd.pluginCall = pluginCall
	cmd.pluginModel = deps.PluginModels.Space
	return cmd
}

func (cmd *ShowSpace) Execute(c flags.FlagContext) {
	space := cmd.spaceReq.GetSpace()
	if cmd.pluginCall {
		cmd.populatePluginModel(space)
		return
	}
	if c.Bool("guid") {
		cmd.ui.Say(space.Guid)
	} else {
		cmd.ui.Say(T("Getting info for space {{.TargetSpace}} in org {{.OrgName}} as {{.CurrentUser}}...",
			map[string]interface{}{
				"TargetSpace": terminal.EntityNameColor(space.Name),
				"OrgName":     terminal.EntityNameColor(space.Organization.Name),
				"CurrentUser": terminal.EntityNameColor(cmd.config.Username()),
			}))

		quotaString := cmd.quotaString(space)
		cmd.ui.Ok()
		cmd.ui.Say("")
		table := terminal.NewTable(cmd.ui, []string{terminal.EntityNameColor(space.Name), "", ""})
		table.Add("", T("Org:"), terminal.EntityNameColor(space.Organization.Name))

		apps := []string{}
		for _, app := range space.Applications {
			apps = append(apps, terminal.EntityNameColor(app.Name))
		}
		table.Add("", T("Apps:"), strings.Join(apps, ", "))

		domains := []string{}
		for _, domain := range space.Domains {
			domains = append(domains, terminal.EntityNameColor(domain.Name))
		}
		table.Add("", T("Domains:"), strings.Join(domains, ", "))

		services := []string{}
		for _, service := range space.ServiceInstances {
			services = append(services, terminal.EntityNameColor(service.Name))
		}
		table.Add("", T("Services:"), strings.Join(services, ", "))

		securityGroups := []string{}
		for _, group := range space.SecurityGroups {
			securityGroups = append(securityGroups, terminal.EntityNameColor(group.Name))
		}
		table.Add("", T("Security Groups:"), strings.Join(securityGroups, ", "))

		table.Add("", T("Space Quota:"), quotaString)

		table.Print()
	}
	if c.Bool("security-group-rules") {
		cmd.ui.Say("")
		for _, group := range space.SecurityGroups {
			cmd.ui.Say(T("Getting rules for the security group  : {{.SecurityGroupName}}...",
				map[string]interface{}{"SecurityGroupName": terminal.EntityNameColor(group.Name)}))
			table := terminal.NewTable(cmd.ui, []string{"", "", "", ""})
			for _, rules := range group.Rules {
				for ruleName, ruleValue := range rules {
					table.Add("", ruleName, ":", fmt.Sprintf("%v", ruleValue))
				}
				table.Add("", "", "", "")
			}
			table.Print()
		}
	}

}

func (cmd *ShowSpace) quotaString(space models.Space) string {
	var instance_memory string

	if space.SpaceQuotaGuid == "" {
		return ""
	}

	quota, err := cmd.quotaRepo.FindByGuid(space.SpaceQuotaGuid)
	if err != nil {
		cmd.ui.Failed(err.Error())
		return ""
	}

	if quota.InstanceMemoryLimit == -1 {
		instance_memory = "-1"
	} else {
		instance_memory = formatters.ByteSize(quota.InstanceMemoryLimit * formatters.MEGABYTE)
	}
	memory := formatters.ByteSize(quota.MemoryLimit * formatters.MEGABYTE)

	spaceQuota := fmt.Sprintf("%s (%s memory limit, %s instance memory limit, %d routes, %d services, paid services %s)", quota.Name, memory, instance_memory, quota.RoutesLimit, quota.ServicesLimit, formatters.Allowed(quota.NonBasicServicesAllowed))
	//	spaceQuota := fmt.Sprintf(T("{{.QuotaName}} ({{.MemoryLimit}} memory limit, {{.InstanceMemoryLimit}} instance memory limit, {{.RoutesLimit}} routes, {{.ServicesLimit}} services, paid services {{.NonBasicServicesAllowed}})",
	//		map[string]interface{}{
	//			"QuotaName":               quota.Name,
	//			"MemoryLimit":             memory,
	//			"InstanceMemoryLimit":     instance_memory,
	//			"RoutesLimit":             quota.RoutesLimit,
	//			"ServicesLimit":           quota.ServicesLimit,
	//			"NonBasicServicesAllowed": formatters.Allowed(quota.NonBasicServicesAllowed)}))

	return spaceQuota
}

func (cmd *ShowSpace) populatePluginModel(space models.Space) {
	cmd.pluginModel.Name = space.Name
	cmd.pluginModel.Guid = space.Guid

	cmd.pluginModel.Organization.Name = space.Organization.Name
	cmd.pluginModel.Organization.Guid = space.Organization.Guid

	for _, app := range space.Applications {
		a := plugin_models.GetSpace_Apps{
			Name: app.Name,
			Guid: app.Guid,
		}
		cmd.pluginModel.Applications = append(cmd.pluginModel.Applications, a)
	}

	for _, domain := range space.Domains {
		d := plugin_models.GetSpace_Domains{
			Name: domain.Name,
			Guid: domain.Guid,
			OwningOrganizationGuid: domain.OwningOrganizationGuid,
			Shared:                 domain.Shared,
		}
		cmd.pluginModel.Domains = append(cmd.pluginModel.Domains, d)
	}

	for _, service := range space.ServiceInstances {
		si := plugin_models.GetSpace_ServiceInstance{
			Name: service.Name,
			Guid: service.Guid,
		}
		cmd.pluginModel.ServiceInstances = append(cmd.pluginModel.ServiceInstances, si)
	}
	for _, group := range space.SecurityGroups {
		sg := plugin_models.GetSpace_SecurityGroup{
			Name:  group.Name,
			Guid:  group.Guid,
			Rules: group.Rules,
		}
		cmd.pluginModel.SecurityGroups = append(cmd.pluginModel.SecurityGroups, sg)
	}

	quota, err := cmd.quotaRepo.FindByGuid(space.SpaceQuotaGuid)
	if err == nil {
		cmd.pluginModel.SpaceQuota.Name = quota.Name
		cmd.pluginModel.SpaceQuota.Guid = quota.Guid
		cmd.pluginModel.SpaceQuota.MemoryLimit = quota.MemoryLimit
		cmd.pluginModel.SpaceQuota.InstanceMemoryLimit = quota.InstanceMemoryLimit
		cmd.pluginModel.SpaceQuota.RoutesLimit = quota.RoutesLimit
		cmd.pluginModel.SpaceQuota.ServicesLimit = quota.ServicesLimit
		cmd.pluginModel.SpaceQuota.NonBasicServicesAllowed = quota.NonBasicServicesAllowed
	}
}
