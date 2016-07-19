package organization

import (
	"fmt"
	"strconv"
	"strings"

	"code.cloudfoundry.org/cli/cf/api/resources"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/flags"
	"code.cloudfoundry.org/cli/cf/formatters"
	. "code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
	"code.cloudfoundry.org/cli/plugin/models"
)

type ShowOrg struct {
	ui          terminal.UI
	config      coreconfig.Reader
	orgReq      requirements.OrganizationRequirement
	pluginModel *plugin_models.GetOrg_Model
	pluginCall  bool
}

func init() {
	commandregistry.Register(&ShowOrg{})
}

func (cmd *ShowOrg) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["guid"] = &flags.BoolFlag{Name: "guid", Usage: T("Retrieve and display the given org's guid.  All other output for the org is suppressed.")}
	return commandregistry.CommandMetadata{
		Name:        "org",
		Description: T("Show org info"),
		Usage: []string{
			T("CF_NAME org ORG"),
		},
		Flags: fs,
	}
}

func (cmd *ShowOrg) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + commandregistry.Commands.CommandUsage("org"))
		return nil, fmt.Errorf("Incorrect usage: %d arguments of %d required", len(fc.Args()), 1)
	}

	cmd.orgReq = requirementsFactory.NewOrganizationRequirement(fc.Args()[0])

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		cmd.orgReq,
	}

	return reqs, nil
}

func (cmd *ShowOrg) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.pluginCall = pluginCall
	cmd.pluginModel = deps.PluginModels.Organization
	return cmd
}

func (cmd *ShowOrg) Execute(c flags.FlagContext) error {
	org := cmd.orgReq.GetOrganization()

	if c.Bool("guid") {
		cmd.ui.Say(org.GUID)
	} else {
		cmd.ui.Say(T("Getting info for org {{.OrgName}} as {{.Username}}...",
			map[string]interface{}{
				"OrgName":  terminal.EntityNameColor(org.Name),
				"Username": terminal.EntityNameColor(cmd.config.Username())}))
		cmd.ui.Ok()
		cmd.ui.Say("")

		table := cmd.ui.Table([]string{terminal.EntityNameColor(org.Name) + ":", "", ""})

		domains := []string{}
		for _, domain := range org.Domains {
			domains = append(domains, domain.Name)
		}

		spaces := []string{}
		for _, space := range org.Spaces {
			spaces = append(spaces, space.Name)
		}

		spaceQuotas := []string{}
		for _, spaceQuota := range org.SpaceQuotas {
			spaceQuotas = append(spaceQuotas, spaceQuota.Name)
		}

		quota := org.QuotaDefinition

		var reservedPortLimit string
		switch string(quota.ReservedRoutePorts) {
		case "":
			break
		case resources.UnlimitedReservedRoutePorts:
			reservedPortLimit = T("unlimited")
		default:
			if _, err := quota.ReservedRoutePorts.Int64(); err != nil {
				return err
			}
			reservedPortLimit = string(quota.ReservedRoutePorts)
		}

		appInstanceLimit := strconv.Itoa(quota.AppInstanceLimit)
		if quota.AppInstanceLimit == resources.UnlimitedAppInstances {
			appInstanceLimit = T("unlimited")
		}

		orgQuotaFields := []string{}
		orgQuotaFields = append(orgQuotaFields, T("{{.MemoryLimit}}M memory limit", map[string]interface{}{"MemoryLimit": quota.MemoryLimit}))
		orgQuotaFields = append(orgQuotaFields, T("{{.InstanceMemoryLimit}} instance memory limit", map[string]interface{}{"InstanceMemoryLimit": formatters.InstanceMemoryLimit(quota.InstanceMemoryLimit)}))
		orgQuotaFields = append(orgQuotaFields, T("{{.RoutesLimit}} routes", map[string]interface{}{"RoutesLimit": quota.RoutesLimit}))
		orgQuotaFields = append(orgQuotaFields, T("{{.ServicesLimit}} services", map[string]interface{}{"ServicesLimit": quota.ServicesLimit}))
		orgQuotaFields = append(orgQuotaFields, T("paid services {{.NonBasicServicesAllowed}}", map[string]interface{}{"NonBasicServicesAllowed": formatters.Allowed(quota.NonBasicServicesAllowed)}))
		orgQuotaFields = append(orgQuotaFields, T("{{.AppInstanceLimit}} app instance limit", map[string]interface{}{"AppInstanceLimit": appInstanceLimit}))

		if reservedPortLimit != "" {
			orgQuotaFields = append(orgQuotaFields, T("{{.ReservedRoutePorts}} route ports", map[string]interface{}{"ReservedRoutePorts": reservedPortLimit}))
		}

		orgQuota := fmt.Sprintf("%s (%s)", quota.Name, strings.Join(orgQuotaFields, ", "))

		if cmd.pluginCall {
			cmd.populatePluginModel(org, quota)
		} else {
			table.Add("", T("domains:"), terminal.EntityNameColor(strings.Join(domains, ", ")))
			table.Add("", T("quota:"), terminal.EntityNameColor(orgQuota))
			table.Add("", T("spaces:"), terminal.EntityNameColor(strings.Join(spaces, ", ")))
			table.Add("", T("space quotas:"), terminal.EntityNameColor(strings.Join(spaceQuotas, ", ")))

			err := table.Print()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (cmd *ShowOrg) populatePluginModel(org models.Organization, quota models.QuotaFields) {
	cmd.pluginModel.Name = org.Name
	cmd.pluginModel.Guid = org.GUID
	cmd.pluginModel.QuotaDefinition.Name = quota.Name
	cmd.pluginModel.QuotaDefinition.MemoryLimit = quota.MemoryLimit
	cmd.pluginModel.QuotaDefinition.InstanceMemoryLimit = quota.InstanceMemoryLimit
	cmd.pluginModel.QuotaDefinition.RoutesLimit = quota.RoutesLimit
	cmd.pluginModel.QuotaDefinition.ServicesLimit = quota.ServicesLimit
	cmd.pluginModel.QuotaDefinition.NonBasicServicesAllowed = quota.NonBasicServicesAllowed

	for _, domain := range org.Domains {
		d := plugin_models.GetOrg_Domains{
			Name: domain.Name,
			Guid: domain.GUID,
			OwningOrganizationGuid: domain.OwningOrganizationGUID,
			Shared:                 domain.Shared,
		}
		cmd.pluginModel.Domains = append(cmd.pluginModel.Domains, d)
	}

	for _, space := range org.Spaces {
		s := plugin_models.GetOrg_Space{
			Name: space.Name,
			Guid: space.GUID,
		}
		cmd.pluginModel.Spaces = append(cmd.pluginModel.Spaces, s)
	}

	for _, spaceQuota := range org.SpaceQuotas {
		sq := plugin_models.GetOrg_SpaceQuota{
			Name:                spaceQuota.Name,
			Guid:                spaceQuota.GUID,
			MemoryLimit:         spaceQuota.MemoryLimit,
			InstanceMemoryLimit: spaceQuota.InstanceMemoryLimit,
		}
		cmd.pluginModel.SpaceQuotas = append(cmd.pluginModel.SpaceQuotas, sq)
	}
}
