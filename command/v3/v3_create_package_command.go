package v3

import (
	"os"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v3/shared"
)

//go:generate counterfeiter . V3CreatePackageActor

type V3CreatePackageActor interface {
	CloudControllerAPIVersion() string
	CreateAndUploadBitsPackageByApplicationNameAndSpace(appName string, spaceGUID string, bitsPath string) (v3action.Package, v3action.Warnings, error)
	CreateDockerPackageByApplicationNameAndSpace(appName string, spaceGUID string, dockerPath string) (v3action.Package, v3action.Warnings, error)
}

type V3CreatePackageCommand struct {
	RequiredArgs flag.AppName     `positional-args:"yes"`
	DockerImage  flag.DockerImage `long:"docker-image" short:"o" description:"Docker-image to be used (e.g. user/docker-image-name)"`
	usage        interface{}      `usage:"CF_NAME v3-create-package APP_NAME [--docker-image [REGISTRY_HOST:PORT/]IMAGE[:TAG]]"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       V3CreatePackageActor
}

func (cmd *V3CreatePackageCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor()

	client, _, err := shared.NewClients(config, ui, true)
	if err != nil {
		return err
	}
	cmd.Actor = v3action.NewActor(client, config)

	return nil
}

func (cmd V3CreatePackageCommand) Execute(args []string) error {
	cmd.UI.DisplayText(command.ExperimentalWarning)
	cmd.UI.DisplayNewline()

	err := command.MinimumAPIVersionCheck(cmd.Actor.CloudControllerAPIVersion(), command.MinVersionV3)
	if err != nil {
		return err
	}

	err = cmd.SharedActor.CheckTarget(cmd.Config, true, true)
	if err != nil {
		return shared.HandleError(err)
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return shared.HandleError(err)
	}

	var (
		pkg      v3action.Package
		warnings v3action.Warnings
	)

	if cmd.DockerImage.Path != "" {
		cmd.UI.DisplayTextWithFlavor("Creating docker package for V3 app {{.AppName}} in org {{.CurrentOrg}} / space {{.CurrentSpace}} as {{.CurrentUser}}...", map[string]interface{}{
			"AppName":      cmd.RequiredArgs.AppName,
			"CurrentSpace": cmd.Config.TargetedSpace().Name,
			"CurrentOrg":   cmd.Config.TargetedOrganization().Name,
			"CurrentUser":  user.Name,
		})

		pkg, warnings, err = cmd.Actor.CreateDockerPackageByApplicationNameAndSpace(cmd.RequiredArgs.AppName, cmd.Config.TargetedSpace().GUID, cmd.DockerImage.Path)
	} else {
		cmd.UI.DisplayTextWithFlavor("Uploading and creating bits package for V3 app {{.AppName}} in org {{.CurrentOrg}} / space {{.CurrentSpace}} as {{.CurrentUser}}...", map[string]interface{}{
			"AppName":      cmd.RequiredArgs.AppName,
			"CurrentSpace": cmd.Config.TargetedSpace().Name,
			"CurrentOrg":   cmd.Config.TargetedOrganization().Name,
			"CurrentUser":  user.Name,
		})

		pwd, osErr := os.Getwd()
		if osErr != nil {
			return shared.HandleError(osErr)
		}
		pkg, warnings, err = cmd.Actor.CreateAndUploadBitsPackageByApplicationNameAndSpace(cmd.RequiredArgs.AppName, cmd.Config.TargetedSpace().GUID, pwd)
	}

	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return shared.HandleError(err)
	}

	cmd.UI.DisplayText("package guid: {{.PackageGuid}}", map[string]interface{}{
		"PackageGuid": pkg.GUID,
	})
	cmd.UI.DisplayOK()

	return nil
}
