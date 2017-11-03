package v2

import (
	"fmt"
	"os"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v2/shared"
)

//go:generate counterfeiter . CreateAppManifestActor

type CreateAppManifestActor interface {
	CreateApplicationManifestByNameAndSpace(appName string, spaceGUID string, filePath string) (v2action.Warnings, error)
}

type CreateAppManifestCommand struct {
	RequiredArgs    flag.AppName `positional-args:"yes"`
	FilePath        flag.Path    `short:"p" description:"Specify a path for file creation. If path not specified, manifest file is created in current working directory."`
	usage           interface{}  `usage:"CF_NAME create-app-manifest APP_NAME [-p /path/to/<app-name>_manifest.yml]"`
	relatedCommands interface{}  `related_commands:"apps, push"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       CreateAppManifestActor
}

func (cmd *CreateAppManifestCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor(config)

	ccClient, uaaClient, err := shared.NewClients(config, ui, true)
	if err != nil {
		return err
	}
	cmd.Actor = v2action.NewActor(ccClient, uaaClient, config)

	return nil
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

	cmd.UI.DisplayTextWithFlavor("Creating an app manifest from current settings of app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
		"AppName":   cmd.RequiredArgs.AppName,
		"OrgName":   cmd.Config.TargetedOrganization().Name,
		"SpaceName": cmd.Config.TargetedSpace().Name,
		"Username":  user.Name,
	})

	manifestPath := cmd.FilePath.String()
	if manifestPath == "" {
		manifestPath = fmt.Sprintf(".%s%s_manifest.yml", string(os.PathSeparator), cmd.RequiredArgs.AppName)
	}
	warnings, err := cmd.Actor.CreateApplicationManifestByNameAndSpace(cmd.RequiredArgs.AppName, cmd.Config.TargetedSpace().GUID, manifestPath)

	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()
	cmd.UI.DisplayText("Manifest file created successfully at {{.FilePath}}", map[string]interface{}{
		"FilePath": manifestPath,
	})

	return nil
}
