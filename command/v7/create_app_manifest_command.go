package v7

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type CreateAppManifestCommand struct {
	BaseCommand

	RequiredArgs    flag.AppName `positional-args:"yes"`
	FilePath        flag.Path    `short:"p" description:"Specify a path for file creation. If path not specified, manifest file is created in current working directory."`
	usage           interface{}  `usage:"CF_NAME create-app-manifest APP_NAME [-p /path/to/<app-name>_manifest.yml]"`
	relatedCommands interface{}  `related_commands:"apps, push"`

	PWD string
}

func (cmd *CreateAppManifestCommand) Setup(config command.Config, ui command.UI) error {
	err := cmd.BaseCommand.Setup(config, ui)
	if err != nil {
		return err
	}
	currentDir, err := os.Getwd()
	cmd.PWD = currentDir

	return err
}

func (cmd CreateAppManifestCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	appName := cmd.RequiredArgs.AppName
	cmd.UI.DisplayTextWithFlavor("Creating an app manifest from current settings of app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
		"AppName":   appName,
		"OrgName":   cmd.Config.TargetedOrganization().Name,
		"SpaceName": cmd.Config.TargetedSpace().Name,
		"Username":  user.Name,
	})

	spaceGUID := cmd.Config.TargetedSpace().GUID
	manifestBytes, warnings, err := cmd.Actor.GetRawApplicationManifestByNameAndSpace(appName, spaceGUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	var pathToYAMLFile string
	if len(cmd.FilePath) > 0 {
		pathToYAMLFile = cmd.FilePath.String()
	} else {
		pathToYAMLFile = filepath.Join(cmd.PWD, fmt.Sprintf("%s_manifest.yml", appName))
	}

	err = ioutil.WriteFile(pathToYAMLFile, manifestBytes, 0666)
	if err != nil {
		return translatableerror.ManifestCreationError{Err: err}
	}

	cmd.UI.DisplayText("Manifest file created successfully at {{.FilePath}}", map[string]interface{}{
		"FilePath": pathToYAMLFile,
	})
	cmd.UI.DisplayOK()

	return nil
}
