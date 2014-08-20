package spacequota

import (
	"github.com/cloudfoundry/cli/cf/api/space_quotas"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"

	. "github.com/cloudfoundry/cli/cf/i18n"
)

type DeleteSpaceQuota struct {
	ui             terminal.UI
	config         configuration.Reader
	spaceQuotaRepo space_quotas.SpaceQuotaRepository
}

func NewDeleteSpaceQuota(ui terminal.UI, config configuration.Reader, spaceQuotaRepo space_quotas.SpaceQuotaRepository) DeleteSpaceQuota {
	return DeleteSpaceQuota{
		ui:             ui,
		config:         config,
		spaceQuotaRepo: spaceQuotaRepo,
	}
}
func (cmd DeleteSpaceQuota) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "delete-space-quota",
		Description: T("Delete a space quota definition and unassign the space quota from all spaces"),
		Usage:       T("CF_NAME delete-space-quota SPACE-QUOTA-NAME"),
		Flags: []cli.Flag{
			cli.BoolFlag{Name: "f", Usage: T("Force delete (do not prompt for confirmation)")},
		},
	}
}

func (cmd DeleteSpaceQuota) GetRequirements(requirementsFactory requirements.Factory, context *cli.Context) ([]requirements.Requirement, error) {
	if len(context.Args()) != 1 {
		cmd.ui.FailWithUsage(context)
	}

	return []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}, nil
}

func (cmd DeleteSpaceQuota) Run(context *cli.Context) {
	quotaName := context.Args()[0]

	if !context.Bool("f") {
		response := cmd.ui.ConfirmDelete("quota", quotaName)
		if !response {
			return
		}
	}

	cmd.ui.Say(T("Deleting space quota {{.QuotaName}} as {{.Username}}...", map[string]interface{}{
		"QuotaName": terminal.EntityNameColor(quotaName),
		"Username":  terminal.EntityNameColor(cmd.config.Username()),
	}))

	quota, apiErr := cmd.spaceQuotaRepo.FindByName(quotaName)
	switch (apiErr).(type) {
	case nil: // no error
	case *errors.ModelNotFoundError:
		cmd.ui.Ok()
		cmd.ui.Warn(T("Quota {{.QuotaName}} does not exist", map[string]interface{}{"QuotaName": quotaName}))
		return
	default:
		cmd.ui.Failed(apiErr.Error())
	}

	apiErr = cmd.spaceQuotaRepo.Delete(quota.Guid)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
	}

	cmd.ui.Ok()
}
