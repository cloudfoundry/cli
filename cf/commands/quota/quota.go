package quota

import (
	"fmt"
	. "github.com/cloudfoundry/cli/cf/i18n"

	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/formatters"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type showQuota struct {
	ui        terminal.UI
	config    configuration.Reader
	quotaRepo api.QuotaRepository
}

func NewShowQuota(ui terminal.UI, config configuration.Reader, quotaRepo api.QuotaRepository) *showQuota {
	return &showQuota{
		ui:        ui,
		config:    config,
		quotaRepo: quotaRepo,
	}
}

func (cmd *showQuota) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "quota",
		Usage:       T("CF_NAME quota QUOTA"),
		Description: T("Show quota info"),
	}
}

func (cmd *showQuota) GetRequirements(requirementsFactory requirements.Factory, context *cli.Context) ([]requirements.Requirement, error) {
	if len(context.Args()) != 1 {
		cmd.ui.FailWithUsage(context)
	}

	return []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}, nil
}

func (cmd *showQuota) Run(context *cli.Context) {
	quotaName := context.Args()[0]
	cmd.ui.Say(T("Getting quota {{.QuotaName}} info as {{.Username}}...", map[string]interface{}{"QuotaName": quotaName, "Username": cmd.config.Username()}))

	quota, err := cmd.quotaRepo.FindByName(quotaName)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	cmd.ui.Ok()

	table := terminal.NewTable(cmd.ui, []string{"", ""})
	table.Add(T("Memory"), formatters.ByteSize(quota.MemoryLimit*formatters.MEGABYTE))
	table.Add(T("Routes"), fmt.Sprintf("%d", quota.RoutesLimit))
	table.Add(T("Services"), fmt.Sprintf("%d", quota.ServicesLimit))
	table.Add(T("Paid service plans"), formatters.Allowed(quota.NonBasicServicesAllowed))
	table.Print()
}
