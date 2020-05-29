package v7

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type DownloadDropletCommand struct {
	BaseCommand

	RequiredArgs    flag.AppName `positional-args:"yes"`
	usage           interface{}  `usage:"CF_NAME download-droplet APP_NAME"`
	relatedCommands interface{}  `related_commands:"apps, droplets, push, set-droplet"`

	// field for setting current working dir for ease of testing
	CWD string
}

func (cmd DownloadDropletCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Downloading current droplet for app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
		"AppName":   cmd.RequiredArgs.AppName,
		"OrgName":   cmd.Config.TargetedOrganization().Name,
		"SpaceName": cmd.Config.TargetedSpace().Name,
		"Username":  user.Name,
	})

	rawDropletBytes, dropletGUID, warnings, err := cmd.Actor.DownloadDropletByAppName(cmd.RequiredArgs.AppName, cmd.Config.TargetedSpace().GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		if _, ok := err.(actionerror.DropletNotFoundError); ok {
			return translatableerror.NoCurrentDropletForAppError{AppName: cmd.RequiredArgs.AppName}
		}
		return err
	}

	var pathToDroplet string
	if cmd.CWD == "" {
		currentDir, err := os.Getwd()
		if err != nil {
			return err
		}
		cmd.CWD = currentDir
	}
	pathToDroplet = filepath.Join(cmd.CWD, fmt.Sprintf("droplet_%s.tgz", dropletGUID))

	err = ioutil.WriteFile(pathToDroplet, rawDropletBytes, 0666)
	if err != nil {
		return translatableerror.DropletFileError{Err: err}
	}

	cmd.UI.DisplayText("Droplet downloaded successfully at {{.FilePath}}", map[string]interface{}{
		"FilePath": pathToDroplet,
	})
	cmd.UI.DisplayOK()

	return nil
}
