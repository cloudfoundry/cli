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
	"path/filepath"

	"github.com/cloudfoundry/cli/cf/i18n"

	goi18n "github.com/nicksnyder/go-i18n/i18n"
)

type showQuota struct {
	ui        terminal.UI
	config    configuration.Reader
	quotaRepo api.QuotaRepository
	T         goi18n.TranslateFunc
}

func NewShowQuota(ui terminal.UI, config configuration.Reader, quotaRepo api.QuotaRepository) *showQuota {
	t, err := i18n.Init("quota", filepath.Join("cf", "i18n", "resources"))
	if err != nil {
		ui.Failed(err.Error())
	}

	return &showQuota{
		ui:        ui,
		config:    config,
		quotaRepo: quotaRepo,
		T:         t,
	}
}

func (cmd *showQuota) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "quota",
		Usage:       cmd.T("CF_NAME quota QUOTA"),
		Description: cmd.T("Show quota info"),
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
	cmd.ui.Say(cmd.T("Getting quota {{.QuotaName}} info as {{.Username}}...", map[string]interface{}{"QuotaName": quotaName, "Username": cmd.config.Username()}))

	quota, err := cmd.quotaRepo.FindByName(quotaName)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	cmd.ui.Ok()

	table := terminal.NewTable(cmd.ui, []string{"", ""})
	table.Add([]string{cmd.T("Memory"), formatters.ByteSize(quota.MemoryLimit * formatters.MEGABYTE)})
	table.Add([]string{cmd.T("Routes"), fmt.Sprintf("%d", quota.RoutesLimit)})
	table.Add([]string{cmd.T("Services"), fmt.Sprintf("%d", quota.ServicesLimit)})
	table.Add([]string{cmd.T("Paid service plans"), formatters.Allowed(quota.NonBasicServicesAllowed)})
	table.Print()
}
