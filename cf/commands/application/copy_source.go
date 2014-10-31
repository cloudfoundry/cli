package application

import (
	"fmt"

	"github.com/cloudfoundry/cli/cf/api/applications"
	"github.com/cloudfoundry/cli/cf/api/authentication"
	"github.com/cloudfoundry/cli/cf/api/copy_application_source"
	"github.com/cloudfoundry/cli/cf/api/organizations"
	"github.com/cloudfoundry/cli/cf/api/spaces"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/flag_helpers"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
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
	appRestart        ApplicationRestarter
}

func NewCopySource(
	ui terminal.UI,
	config core_config.Reader,
	authRepo authentication.AuthenticationRepository,
	appRepo applications.ApplicationRepository,
	orgRepo organizations.OrganizationRepository,
	spaceRepo spaces.SpaceRepository,
	copyAppSourceRepo copy_application_source.CopyApplicationSourceRepository,
	appRestart ApplicationRestarter,
) *CopySource {

	return &CopySource{
		ui:                ui,
		config:            config,
		authRepo:          authRepo,
		appRepo:           appRepo,
		orgRepo:           orgRepo,
		spaceRepo:         spaceRepo,
		copyAppSourceRepo: copyAppSourceRepo,
		appRestart:        appRestart,
	}
}

func (cmd *CopySource) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "copy-source",
		Description: T("Make a copy of app source code from one application to another.  Unless overridden, the copy-source command will restart the application."),
		Usage:       T("   CF_NAME copy-source SOURCE-APP TARGET-APP [-o TARGET-ORG] [-s TARGET-SPACE] [--no-restart]\n"),
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

	targetOrg := c.String("o")
	targetSpace := c.String("s")

	if targetOrg != "" && targetSpace == "" {
		cmd.ui.Failed(T("Please provide the space within the organization containing the target application"))
	}

	copyStr := fmt.Sprintf(T("Copying source from app {{.SourceApp}} to target app {{.TargetApp}} within currently targeted organization and space",
		map[string]interface{}{
			"SourceApp": terminal.EntityNameColor(sourceAppName),
			"TargetApp": terminal.EntityNameColor(targetAppName),
		},
	))

	_, apiErr := cmd.authRepo.RefreshAuthToken()
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
	}

	sourceApp, apiErr := cmd.appRepo.Read(sourceAppName)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
	}

	spaceGuid := cmd.config.SpaceFields().Guid

	if targetOrg != "" && targetSpace != "" {
		spaceGuid = cmd.findSpaceGuid(targetOrg, targetSpace)
		copyStr = fmt.Sprintf(T("Copying source from app {{.SourceApp}} to target app {{.TargetApp}} in organization {{.Org}} and space {{.Space}}",
			map[string]interface{}{
				"SourceApp": terminal.EntityNameColor(sourceAppName),
				"TargetApp": terminal.EntityNameColor(targetAppName),
				"Org":       terminal.EntityNameColor(targetOrg),
				"Space":     terminal.EntityNameColor(targetSpace),
			},
		))
	} else if targetSpace != "" {
		space, err := cmd.spaceRepo.FindByName(targetSpace)
		if err != nil {
			cmd.ui.Failed(err.Error())
		}
		spaceGuid = space.Guid

		copyStr = fmt.Sprintf(T("Copying source from app {{.SourceApp}} to target app {{.TargetApp}} within currently targeted organization and space {{.Space}}",
			map[string]interface{}{
				"SourceApp": terminal.EntityNameColor(sourceAppName),
				"TargetApp": terminal.EntityNameColor(targetAppName),
				"Space":     terminal.EntityNameColor(targetSpace),
			},
		))
	}

	targetApp, apiErr := cmd.appRepo.ReadFromSpace(targetAppName, spaceGuid)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
	}

	cmd.ui.Say(copyStr)
	cmd.ui.Say(T("Note: this may take some time"))
	cmd.ui.Say("")

	apiErr = cmd.copyAppSourceRepo.CopyApplication(sourceApp.Guid, targetApp.Guid)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
	}

	if !c.Bool("no-restart") {
		cmd.appRestart.ApplicationRestart(targetApp)
	}

	cmd.ui.Ok()
}

func (cmd *CopySource) findSpaceGuid(targetOrg, targetSpace string) string {
	org, err := cmd.orgRepo.FindByName(targetOrg)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	var space models.SpaceFields
	var foundSpace bool
	for _, s := range org.Spaces {
		if s.Name == targetSpace {
			space = s
			foundSpace = true
		}
	}

	if !foundSpace {
		cmd.ui.Failed(fmt.Sprintf(T("Could not find space {{.Space}} in organization {{.Org}}",
			map[string]interface{}{
				"Space": terminal.EntityNameColor(targetSpace),
				"Org":   terminal.EntityNameColor(targetOrg),
			},
		)))
	}

	return space.Guid
}
