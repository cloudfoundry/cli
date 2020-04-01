package v7

import (
	"code.cloudfoundry.org/cli/cf/appfiles"
	"code.cloudfoundry.org/cli/command/flag"
	"os"
	"path/filepath"
)

type PullCommand struct {
	BaseCommand

	RequiredArgs    flag.AppName `positional-args:"yes"`
	GUID            bool         `long:"guid" description:"Sync local directory with current code from the given app"`
	usage           interface{}  `usage:"CF_NAME app APP_NAME"`
	relatedCommands interface{}  `related_commands:"push"`
}

func (cmd PullCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Syncing local directory with {{.AppName}}, all existing files will be removed...", map[string]interface{}{
		"AppName": cmd.RequiredArgs.AppName,
	})
	cmd.UI.DisplayNewline()

	app, warnings, err := cmd.Actor.GetApplicationByNameAndSpace(cmd.RequiredArgs.AppName, cmd.Config.TargetedSpace().GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	appPackage, warnings, err := cmd.Actor.GetNewestReadyPackageForApplication(app.GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	bitsFile, warnings, err := cmd.Actor.DownloadBits(appPackage)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	appPath, err := os.Getwd()
	if err != nil {
		return err
	}

	err = cmd.cleanDir(appPath)
	if err != nil {
		return err
	}

	zipper := appfiles.ApplicationZipper{}
	err = zipper.Unzip(bitsFile.Name(), appPath)
	if err != nil {
		return err
	}

	err = os.Remove(bitsFile.Name())
	if err != nil {
		return err
	}
	return nil
}

func (cmd PullCommand) cleanDir(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}
