package spacequota

import (
	"fmt"

	"code.cloudfoundry.org/cli/cf/api/spacequotas"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/flags"
	"code.cloudfoundry.org/cli/cf/formatters"
	. "code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
)

type SpaceQuota struct {
	ui             terminal.UI
	config         coreconfig.Reader
	spaceQuotaRepo spacequotas.SpaceQuotaRepository
}

func init() {
	commandregistry.Register(&SpaceQuota{})
}

func (cmd *SpaceQuota) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "space-quota",
		Description: T("Show space quota info"),
		Usage: []string{
			T("CF_NAME space-quota SPACE_QUOTA_NAME"),
		},
	}
}

func (cmd *SpaceQuota) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + commandregistry.Commands.CommandUsage("space-quota"))
		return nil, fmt.Errorf("Incorrect usage: %d arguments of %d required", len(fc.Args()), 1)
	}

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedOrgRequirement(),
	}

	return reqs, nil
}

func (cmd *SpaceQuota) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.spaceQuotaRepo = deps.RepoLocator.GetSpaceQuotaRepository()
	return cmd
}

func (cmd *SpaceQuota) Execute(c flags.FlagContext) error {
	name := c.Args()[0]

	cmd.ui.Say(T("Getting space quota {{.Quota}} info as {{.Username}}...",
		map[string]interface{}{
			"Quota":    terminal.EntityNameColor(name),
			"Username": terminal.EntityNameColor(cmd.config.Username()),
		}))

	spaceQuota, err := cmd.spaceQuotaRepo.FindByName(name)

	if err != nil {
		return err
	}

	cmd.ui.Ok()
	cmd.ui.Say("")
	var megabytes string

	table := cmd.ui.Table([]string{"", ""})
	table.Add(T("total memory limit"), formatters.ByteSize(spaceQuota.MemoryLimit*formatters.MEGABYTE))
	if spaceQuota.InstanceMemoryLimit == -1 {
		megabytes = T("unlimited")
	} else {
		megabytes = formatters.ByteSize(spaceQuota.InstanceMemoryLimit * formatters.MEGABYTE)
	}

	table.Add(T("instance memory limit"), megabytes)
	table.Add(T("routes"), fmt.Sprintf("%d", spaceQuota.RoutesLimit))
	table.Add(T("services"), T(spaceQuota.FormattedServicesLimit()))
	table.Add(T("non basic services"), formatters.Allowed(spaceQuota.NonBasicServicesAllowed))
	table.Add(T("app instance limit"), T(spaceQuota.FormattedAppInstanceLimit()))
	table.Add(T("reserved route ports"), T(spaceQuota.FormattedRoutePortsLimit()))

	err = table.Print()
	if err != nil {
		return err
	}
	return nil
}
