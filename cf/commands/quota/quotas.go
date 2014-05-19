package quota

import (
	"fmt"
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/formatters"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"

	"github.com/cloudfoundry/cli/cf/i18n"

	goi18n "github.com/nicksnyder/go-i18n/i18n"
)

type ListQuotas struct {
	ui        terminal.UI
	config    configuration.Reader
	quotaRepo api.QuotaRepository
	T         goi18n.TranslateFunc
}

func NewListQuotas(ui terminal.UI, config configuration.Reader, quotaRepo api.QuotaRepository) (cmd *ListQuotas) {
	t, err := i18n.Init("quota", i18n.GetResourcesPath())
	if err != nil {
		ui.Failed(err.Error())
	}

	return &ListQuotas{
		ui:        ui,
		config:    config,
		quotaRepo: quotaRepo,
		T:         t,
	}
}

func (cmd *ListQuotas) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "quotas",
		Description: cmd.T("List available usage quotas"),
		Usage:       cmd.T("CF_NAME quotas"),
	}
}

func (cmd *ListQuotas) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}
	return
}

func (cmd *ListQuotas) Run(c *cli.Context) {
	cmd.ui.Say(cmd.T("Getting quotas as {{.Username}}...", map[string]interface{}{"Username": terminal.EntityNameColor(cmd.config.Username())}))

	quotas, apiErr := cmd.quotaRepo.FindAll()

	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}
	cmd.ui.Ok()
	cmd.ui.Say("")

	table := terminal.NewTable(cmd.ui, []string{cmd.T("name"), cmd.T("memory limit"), cmd.T("routes"), cmd.T("service instances"), cmd.T("paid service plans")})

	for _, quota := range quotas {
		table.Add([]string{
			quota.Name,
			formatters.ByteSize(quota.MemoryLimit * formatters.MEGABYTE),
			fmt.Sprintf("%d", quota.RoutesLimit),
			fmt.Sprintf("%d", quota.ServicesLimit),
			formatters.Allowed(quota.NonBasicServicesAllowed),
		})
	}

	table.Print()
}
