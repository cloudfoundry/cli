package application

import (
	"fmt"

	"github.com/cloudfoundry/cli/cf/api/applications"
	"github.com/cloudfoundry/cli/cf/api/authentication"
	"github.com/cloudfoundry/cli/cf/api/copy_application_source"
	"github.com/cloudfoundry/cli/cf/api/organizations"
	"github.com/cloudfoundry/cli/cf/api/spaces"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
	"github.com/cloudfoundry/cli/flags/flag"
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

func init() {
	command_registry.Register(&CopySource{})
}

func (cmd *CopySource) MetaData() command_registry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["no-restart"] = &cliFlags.BoolFlag{Name: "no-restart", Usage: T("Override restart of the application in target environment after copy-source completes")}
	fs["o"] = &cliFlags.StringFlag{ShortName: "o", Usage: T("Org that contains the target application")}
	fs["s"] = &cliFlags.StringFlag{ShortName: "s", Usage: T("Space that contains the target application")}

	return command_registry.CommandMetadata{
		Name:        "copy-source",
		Description: T("Make a copy of app source code from one application to another.  Unless overridden, the copy-source command will restart the application."),
		Usage:       T("   CF_NAME copy-source SOURCE-APP TARGET-APP [-o TARGET-ORG] [-s TARGET-SPACE] [--no-restart]\n"),
		Flags:       fs,
	}
}

func (cmd *CopySource) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(fc.Args()) != 2 {
		cmd.ui.Failed(T("Incorrect Usage. Requires SOURCE-APP TARGET-APP as arguments\n\n") + command_registry.Commands.CommandUsage("copy-source"))
	}

	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
	}
	return
}

func (cmd *CopySource) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.authRepo = deps.RepoLocator.GetAuthenticationRepository()
	cmd.appRepo = deps.RepoLocator.GetApplicationRepository()
	cmd.orgRepo = deps.RepoLocator.GetOrganizationRepository()
	cmd.spaceRepo = deps.RepoLocator.GetSpaceRepository()
	cmd.copyAppSourceRepo = deps.RepoLocator.GetCopyApplicationSourceRepository()

	//get command from registry for dependency
	commandDep := command_registry.Commands.FindCommand("restart")
	commandDep = commandDep.SetDependency(deps, false)
	cmd.appRestart = commandDep.(ApplicationRestarter)

	return cmd
}

func (cmd *CopySource) Execute(c flags.FlagContext) {
	sourceAppName := c.Args()[0]
	targetAppName := c.Args()[1]

	targetOrg := c.String("o")
	targetSpace := c.String("s")

	if targetOrg != "" && targetSpace == "" {
		cmd.ui.Failed(T("Please provide the space within the organization containing the target application"))
	}

	_, apiErr := cmd.authRepo.RefreshAuthToken()
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
	}

	sourceApp, apiErr := cmd.appRepo.Read(sourceAppName)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
	}

	var targetOrgName, targetSpaceName, spaceGuid, copyStr string
	if targetOrg != "" && targetSpace != "" {
		spaceGuid = cmd.findSpaceGuid(targetOrg, targetSpace)
		targetOrgName = targetOrg
		targetSpaceName = targetSpace
	} else if targetSpace != "" {
		space, err := cmd.spaceRepo.FindByName(targetSpace)
		if err != nil {
			cmd.ui.Failed(err.Error())
		}
		spaceGuid = space.Guid
		targetOrgName = cmd.config.OrganizationFields().Name
		targetSpaceName = targetSpace
	} else {
		spaceGuid = cmd.config.SpaceFields().Guid
		targetOrgName = cmd.config.OrganizationFields().Name
		targetSpaceName = cmd.config.SpaceFields().Name
	}

	copyStr = buildCopyString(sourceAppName, targetAppName, targetOrgName, targetSpaceName, cmd.config.Username())

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
		cmd.appRestart.ApplicationRestart(targetApp, targetOrgName, targetSpaceName)
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

func buildCopyString(sourceAppName, targetAppName, targetOrgName, targetSpaceName, username string) string {
	return fmt.Sprintf(T("Copying source from app {{.SourceApp}} to target app {{.TargetApp}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...",
		map[string]interface{}{
			"SourceApp": terminal.EntityNameColor(sourceAppName),
			"TargetApp": terminal.EntityNameColor(targetAppName),
			"OrgName":   terminal.EntityNameColor(targetOrgName),
			"SpaceName": terminal.EntityNameColor(targetSpaceName),
			"Username":  terminal.EntityNameColor(username),
		},
	))

}
