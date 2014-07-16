package application

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/errors"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type Restage struct {
	ui                terminal.UI
	config            configuration.Reader
	appRepo           api.ApplicationRepository
	appStagingWatcher ApplicationStagingWatcher
}

func NewRestage(ui terminal.UI, config configuration.Reader, appRepo api.ApplicationRepository, stagingWatcher ApplicationStagingWatcher) *Restage {
	cmd := new(Restage)
	cmd.ui = ui
	cmd.config = config
	cmd.appRepo = appRepo
	cmd.appStagingWatcher = stagingWatcher
	return cmd
}

func (cmd *Restage) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "restage",
		ShortName:   "rg",
		Description: T("Restage an app"),
		Usage:       T("CF_NAME restage APP"),
	}
}

func (cmd *Restage) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		cmd.ui.FailWithUsage(c)
	}

	return []requirements.Requirement{requirementsFactory.NewLoginRequirement()}, nil
}

func (cmd *Restage) Run(c *cli.Context) {
	app, err := cmd.appRepo.Read(c.Args()[0])
	if notFound, ok := err.(*errors.ModelNotFoundError); ok {
		cmd.ui.Failed(notFound.Error())
	}

	cmd.ui.Say(T("Restaging app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"AppName":     terminal.EntityNameColor(app.Name),
			"OrgName":     terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			"SpaceName":   terminal.EntityNameColor(cmd.config.SpaceFields().Name),
			"CurrentUser": terminal.EntityNameColor(cmd.config.Username()),
		}))

	cmd.appStagingWatcher.ApplicationWatchStaging(app, func(app models.Application) (models.Application, error) {
		return app, cmd.appRepo.CreateRestageRequest(app.Guid)
	})
}
