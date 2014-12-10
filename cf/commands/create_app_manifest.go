package commands

import (
	. "github.com/cloudfoundry/cli/cf/i18n"

	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/api/app_instances"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type AppManifestCreater interface {
	CreateAppManifest(app models.Application, orgName string, spaceName string)
}

type CreateAppManifest struct {
	ui               terminal.UI
	config           core_config.Reader
	appSummaryRepo   api.AppSummaryRepository
	appInstancesRepo app_instances.AppInstancesRepository
	appReq           requirements.ApplicationRequirement
}

func NewCreateAppManifest(ui terminal.UI, config core_config.Reader, appSummaryRepo api.AppSummaryRepository) (cmd *CreateAppManifest) {
	cmd = new(CreateAppManifest)
	cmd.ui = ui
	cmd.config = config
	cmd.appSummaryRepo = appSummaryRepo
	// cmd.appInstancesRepo = appInstancesRepo
	return
}

func (cmd *CreateAppManifest) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "create-app-manifest",
		Description: T("Create an app manifest for an app that has been pushed successfully."),
		Usage:       T("CF_NAME create-app-manifest [-p /path/to/<app-name>-manifest.yml ]"),
		Flags: []cli.Flag{
			cli.BoolFlag{Name: "p", Usage: T("Specify a path for file creation.  If path not specified, file is create in root directory of the application source code.")},
		},
	}
}

func (cmd *CreateAppManifest) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		cmd.ui.FailWithUsage(c)
	}

	cmd.appReq = requirementsFactory.NewApplicationRequirement(c.Args()[0])

	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
		cmd.appReq,
	}
	return
}

func (cmd *CreateAppManifest) Run(c *cli.Context) {
	// app := cmd.appReq.GetApplication()

}
