package v3

import (
	"os"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/v3/shared"
)

//go:generate counterfeiter . V3CreatePackageActor

type V3CreatePackageActor interface {
	CreateAndUploadPackageByApplicationNameAndSpace(appName string, spaceGUID string, bitsPath string) (v3action.Package, v3action.Warnings, error)
}

type V3CreatePackageCommand struct {
	usage   interface{} `usage:"CF_NAME v3-create-package --name [name]"`
	AppName string      `short:"n" long:"name" description:"The application name" required:"true"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       V3CreatePackageActor
}

func (cmd *V3CreatePackageCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor()

	client, err := shared.NewClients(config, ui, true)
	if err != nil {
		return err
	}
	cmd.Actor = v3action.NewActor(client, config)

	return nil
}

func (cmd V3CreatePackageCommand) Execute(args []string) error {
	cmd.UI.DisplayText(command.ExperimentalWarning)
	cmd.UI.DisplayNewline()

	err := cmd.SharedActor.CheckTarget(cmd.Config, true, true)
	if err != nil {
		return shared.HandleError(err)
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return shared.HandleError(err)
	}

	cmd.UI.DisplayTextWithFlavor("Uploading V3 app {{.AppName}} in org {{.CurrentOrg}} / space {{.CurrentSpace}} as {{.CurrentUser}}...", map[string]interface{}{
		"AppName":      cmd.AppName,
		"CurrentSpace": cmd.Config.TargetedSpace().Name,
		"CurrentOrg":   cmd.Config.TargetedOrganization().Name,
		"CurrentUser":  user.Name,
	})

	pwd, err := os.Getwd()
	if err != nil {
		return shared.HandleError(err)
	}

	_, warnings, err := cmd.Actor.CreateAndUploadPackageByApplicationNameAndSpace(cmd.AppName, cmd.Config.TargetedSpace().GUID, pwd)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return shared.HandleError(err)
	}

	cmd.UI.DisplayOK()

	return nil
}
