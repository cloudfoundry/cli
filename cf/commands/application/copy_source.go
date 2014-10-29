package application

import (
	. "github.com/cloudfoundry/cli/cf/i18n"

	"github.com/cloudfoundry/cli/cf/api/applications"
	"github.com/cloudfoundry/cli/cf/api/authentication"
	"github.com/cloudfoundry/cli/cf/api/copy_application_source"
	"github.com/cloudfoundry/cli/cf/api/organizations"
	"github.com/cloudfoundry/cli/cf/api/spaces"

	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/flag_helpers"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type CopySource struct {
	ui                terminal.UI
	config            core_config.Reader
	authRepo          authentication.AuthenticationRepository
	appRepo           applications.ApplicationRepository
	orgRepo           organizations.OrganizationRepository
	spaceRepo         spaces.SpaceRepository
	copyAppSourceRepo copy_application_source.CopyApplicationSourceRepository
}

func NewCopySource(
	ui terminal.UI,
	config core_config.Reader,
	authRepo authentication.AuthenticationRepository,
	appRepo applications.ApplicationRepository,
	orgRepo organizations.OrganizationRepository,
	spaceRepo spaces.SpaceRepository,
	copyAppSourceRepo copy_application_source.CopyApplicationSourceRepository,
) *CopySource {

	return &CopySource{
		ui:                ui,
		config:            config,
		authRepo:          authRepo,
		appRepo:           appRepo,
		orgRepo:           orgRepo,
		spaceRepo:         spaceRepo,
		copyAppSourceRepo: copyAppSourceRepo,
	}
}

func (cmd *CopySource) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "copy-source",
		Description: T("Make a copy of app source code from one application to another.  Unless overridden, the copy-source command will restart the application."),
		Usage:       T("Copy an app\n") + T("   CF_NAME copy-source SOURCE-APP TARGET-APP [-o TARGET-ORG] [-s TARGET-SPACE] [--no-restart]\n"),
		Flags: []cli.Flag{
			flag_helpers.NewStringFlag("o", T("Org that contains the target application")),
			flag_helpers.NewStringFlag("s", T("Space that contains the target application")),
			cli.BoolFlag{Name: "no-restart", Usage: T("Override restart of the application in target environment after copy-source completes")},
		},
	}
}

func (cmd *CopySource) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 2 {
		cmd.ui.FailWithUsage(c)
	}

	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
	}
	return
}

func (cmd *CopySource) Run(c *cli.Context) {
	sourceAppName := c.Args()[0]
	targetAppName := c.Args()[1]

	_, apiErr := cmd.authRepo.RefreshAuthToken()
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	sourceApp, _ := cmd.appRepo.Read(sourceAppName)
	spaceGuid := cmd.config.SpaceFields().Guid
	targetApp, _ := cmd.appRepo.ReadFromSpace(targetAppName, spaceGuid)
	cmd.copyAppSourceRepo.CopyApplication(sourceApp.Guid, targetApp.Guid)
	//obtain source app guid
	//
	//
	//
	////obtain target app guid from space guid
}
