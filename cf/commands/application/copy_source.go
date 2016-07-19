package application

import (
	"errors"
	"fmt"

	"code.cloudfoundry.org/cli/cf/api/applications"
	"code.cloudfoundry.org/cli/cf/api/authentication"
	"code.cloudfoundry.org/cli/cf/api/copyapplicationsource"
	"code.cloudfoundry.org/cli/cf/api/organizations"
	"code.cloudfoundry.org/cli/cf/api/spaces"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/flags"
	. "code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
)

type CopySource struct {
	ui                terminal.UI
	config            coreconfig.Reader
	authRepo          authentication.Repository
	appRepo           applications.Repository
	orgRepo           organizations.OrganizationRepository
	spaceRepo         spaces.SpaceRepository
	copyAppSourceRepo copyapplicationsource.Repository
	appRestart        Restarter
}

func init() {
	commandregistry.Register(&CopySource{})
}

func (cmd *CopySource) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["no-restart"] = &flags.BoolFlag{Name: "no-restart", Usage: T("Override restart of the application in target environment after copy-source completes")}
	fs["o"] = &flags.StringFlag{ShortName: "o", Usage: T("Org that contains the target application")}
	fs["s"] = &flags.StringFlag{ShortName: "s", Usage: T("Space that contains the target application")}

	return commandregistry.CommandMetadata{
		Name:        "copy-source",
		Description: T("Copies the source code of an application to another existing application (and restarts that application)"),
		Usage: []string{
			T("   CF_NAME copy-source SOURCE-APP TARGET-APP [-s TARGET-SPACE [-o TARGET-ORG]] [--no-restart]\n"),
		},
		Flags: fs,
	}
}

func (cmd *CopySource) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	usageReq := requirementsFactory.NewUsageRequirement(commandregistry.CLICommandUsagePresenter(cmd),
		T("Requires SOURCE-APP TARGET-APP as arguments"),
		func() bool {
			return len(fc.Args()) != 2
		},
	)

	reqs := []requirements.Requirement{
		usageReq,
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
	}

	return reqs, nil
}

func (cmd *CopySource) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.authRepo = deps.RepoLocator.GetAuthenticationRepository()
	cmd.appRepo = deps.RepoLocator.GetApplicationRepository()
	cmd.orgRepo = deps.RepoLocator.GetOrganizationRepository()
	cmd.spaceRepo = deps.RepoLocator.GetSpaceRepository()
	cmd.copyAppSourceRepo = deps.RepoLocator.GetCopyApplicationSourceRepository()

	//get command from registry for dependency
	commandDep := commandregistry.Commands.FindCommand("restart")
	commandDep = commandDep.SetDependency(deps, false)
	cmd.appRestart = commandDep.(Restarter)

	return cmd
}

func (cmd *CopySource) Execute(c flags.FlagContext) error {
	sourceAppName := c.Args()[0]
	targetAppName := c.Args()[1]

	targetOrg := c.String("o")
	targetSpace := c.String("s")

	if targetOrg != "" && targetSpace == "" {
		return errors.New(T("Please provide the space within the organization containing the target application"))
	}

	_, err := cmd.authRepo.RefreshAuthToken()
	if err != nil {
		return err
	}

	sourceApp, err := cmd.appRepo.Read(sourceAppName)
	if err != nil {
		return err
	}

	var targetOrgName, targetSpaceName, spaceGUID, copyStr string
	if targetOrg != "" && targetSpace != "" {
		spaceGUID, err = cmd.findSpaceGUID(targetOrg, targetSpace)
		if err != nil {
			return err
		}

		targetOrgName = targetOrg
		targetSpaceName = targetSpace
	} else if targetSpace != "" {
		var space models.Space
		space, err = cmd.spaceRepo.FindByName(targetSpace)
		if err != nil {
			return err
		}
		spaceGUID = space.GUID
		targetOrgName = cmd.config.OrganizationFields().Name
		targetSpaceName = targetSpace
	} else {
		spaceGUID = cmd.config.SpaceFields().GUID
		targetOrgName = cmd.config.OrganizationFields().Name
		targetSpaceName = cmd.config.SpaceFields().Name
	}

	copyStr = buildCopyString(sourceAppName, targetAppName, targetOrgName, targetSpaceName, cmd.config.Username())

	targetApp, err := cmd.appRepo.ReadFromSpace(targetAppName, spaceGUID)
	if err != nil {
		return err
	}

	cmd.ui.Say(copyStr)
	cmd.ui.Say(T("Note: this may take some time"))
	cmd.ui.Say("")

	err = cmd.copyAppSourceRepo.CopyApplication(sourceApp.GUID, targetApp.GUID)
	if err != nil {
		return err
	}

	if !c.Bool("no-restart") {
		cmd.appRestart.ApplicationRestart(targetApp, targetOrgName, targetSpaceName)
	}

	cmd.ui.Ok()
	return nil
}

func (cmd *CopySource) findSpaceGUID(targetOrg, targetSpace string) (string, error) {
	org, err := cmd.orgRepo.FindByName(targetOrg)
	if err != nil {
		return "", err
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
		return "", fmt.Errorf(T("Could not find space {{.Space}} in organization {{.Org}}",
			map[string]interface{}{
				"Space": terminal.EntityNameColor(targetSpace),
				"Org":   terminal.EntityNameColor(targetOrg),
			},
		))
	}

	return space.GUID, nil
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
