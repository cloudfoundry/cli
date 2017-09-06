package v3

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v3/shared"
	"code.cloudfoundry.org/cli/version"
)

//go:generate counterfeiter . V3CreatePackageActor

type V3CreatePackageActor interface {
	CloudControllerAPIVersion() string
	CreatePackageByApplicationNameAndSpace(appName string, spaceGUID string, bitsPath string, dockerImage string) (v3action.Package, v3action.Warnings, error)
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

	err := version.MinimumAPIVersionCheck(cmd.Actor.CloudControllerAPIVersion(), version.MinVersionV3)
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

	var flavorTextTemplate string
	if cmd.DockerImage.Path != "" {
		flavorTextTemplate = "Creating docker package for app {{.AppName}} in org {{.CurrentOrg}} / space {{.CurrentSpace}} as {{.CurrentUser}}..."
	} else {
		flavorTextTemplate = "Uploading and creating bits package for app {{.AppName}} in org {{.CurrentOrg}} / space {{.CurrentSpace}} as {{.CurrentUser}}..."
	}

	cmd.UI.DisplayTextWithFlavor(flavorTextTemplate, map[string]interface{}{
		"AppName":      cmd.RequiredArgs.AppName,
		"CurrentSpace": cmd.Config.TargetedSpace().Name,
		"CurrentOrg":   cmd.Config.TargetedOrganization().Name,
		"CurrentUser":  user.Name,
	})

	pkg, warnings, err := cmd.Actor.CreatePackageByApplicationNameAndSpace(cmd.RequiredArgs.AppName, cmd.Config.TargetedSpace().GUID, "", cmd.DockerImage.Path)

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
