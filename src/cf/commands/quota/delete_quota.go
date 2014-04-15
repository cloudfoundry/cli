package quota

import (
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	"cf/terminal"
	//	"errors"
	"cf/errors"
	"fmt"
	"github.com/codegangsta/cli"
)

type DeleteQuota struct {
	ui        terminal.UI
	config    configuration.Reader
	quotaRepo api.QuotaRepository
	orgReq    requirements.OrganizationRequirement
}

func NewDeleteQuota(ui terminal.UI, config configuration.Reader, quotaRepo api.QuotaRepository) (cmd *DeleteQuota) {
	cmd = new(DeleteQuota)
	cmd.ui = ui
	cmd.config = config
	cmd.quotaRepo = quotaRepo
	return
}

func (cmd *DeleteQuota) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "delete-quota")
		return
	}

	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}
	return
}

func (cmd *DeleteQuota) Run(c *cli.Context) {
	quotaName := c.Args()[0]

	if !c.Bool("f") {
		response := cmd.ui.ConfirmDelete("quota", quotaName)
		if !response {
			return
		}
	}

	cmd.ui.Say("Deleting quota %s as %s...",
		terminal.EntityNameColor(quotaName),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	quota, apiErr := cmd.quotaRepo.FindByName(quotaName)

	switch (apiErr).(type) {
	case nil: // no error
	case *errors.ModelNotFoundError:
		cmd.ui.Ok()
		cmd.ui.Warn("Quota %s does not exist", quotaName)
		return
	default:
		cmd.ui.Failed(apiErr.Error())
	}

	apiErr = cmd.quotaRepo.Delete(quota.Guid)
	if apiErr != nil {
		fmt.Println("Failed to delete")
		cmd.ui.Failed(apiErr.Error())
	}

	cmd.ui.Ok()
}
