package spacequota

import (
	"fmt"
	"strconv"

	"code.cloudfoundry.org/cli/cf/api/spacequotas"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/flags"
	"code.cloudfoundry.org/cli/cf/formatters"
	. "code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
)

type ListSpaceQuotas struct {
	ui             terminal.UI
	config         coreconfig.Reader
	spaceQuotaRepo spacequotas.SpaceQuotaRepository
}

func init() {
	commandregistry.Register(&ListSpaceQuotas{})
}

func (cmd *ListSpaceQuotas) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "space-quotas",
		Description: T("List available space resource quotas"),
		Usage: []string{
			T("CF_NAME space-quotas"),
		},
	}
}

func (cmd *ListSpaceQuotas) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	usageReq := requirements.NewUsageRequirement(commandregistry.CLICommandUsagePresenter(cmd),
		T("No argument required"),
		func() bool {
			return len(fc.Args()) != 0
		},
	)

	reqs := []requirements.Requirement{
		usageReq,
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedOrgRequirement(),
	}

	return reqs, nil
}

func (cmd *ListSpaceQuotas) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.spaceQuotaRepo = deps.RepoLocator.GetSpaceQuotaRepository()
	return cmd
}

func (cmd *ListSpaceQuotas) Execute(c flags.FlagContext) error {
	cmd.ui.Say(T("Getting space quotas as {{.Username}}...", map[string]interface{}{"Username": terminal.EntityNameColor(cmd.config.Username())}))

	quotas, err := cmd.spaceQuotaRepo.FindByOrg(cmd.config.OrganizationFields().GUID)

	if err != nil {
		return err
	}

	cmd.ui.Ok()
	cmd.ui.Say("")

	table := cmd.ui.Table([]string{
		T("name"),
		T("total memory"),
		T("instance memory"),
		T("routes"),
		T("service instances"),
		T("paid plans"),
		T("app instances"),
		T("route ports"),
	})

	var megabytes string

	for _, quota := range quotas {
		if quota.InstanceMemoryLimit == -1 {
			megabytes = T("unlimited")
		} else {
			megabytes = formatters.ByteSize(quota.InstanceMemoryLimit * formatters.MEGABYTE)
		}

		servicesLimit := strconv.Itoa(quota.ServicesLimit)
		if servicesLimit == "-1" {
			servicesLimit = T("unlimited")
		}

		table.Add(
			quota.Name,
			formatters.ByteSize(quota.MemoryLimit*formatters.MEGABYTE),
			megabytes,
			fmt.Sprintf("%d", quota.RoutesLimit),
			T(quota.FormattedServicesLimit()),
			formatters.Allowed(quota.NonBasicServicesAllowed),
			T(quota.FormattedAppInstanceLimit()),
			T(quota.FormattedRoutePortsLimit()),
		)
	}

	err = table.Print()
	if err != nil {
		return err
	}
	return nil
}
