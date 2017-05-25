package v3

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/v3/shared"
)

//go:generate counterfeiter . V3StageActor

type V3StageActor interface {
	StagePackage(packageGUID string) (<-chan v3action.Build, <-chan v3action.Warnings, <-chan error)
}

type V3StageCommand struct {
	usage       interface{} `usage:"CF_NAME v3-create-app --name [name]"`
	AppName     string      `short:"n" long:"name" description:"The desired application name" required:"true"`
	PackageGUID string      `long:"package-guid" description:"The guid of the package to stage" required:"true"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       V3StageActor
}

func (cmd *V3StageCommand) Setup(config command.Config, ui command.UI) error {
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

func (cmd V3StageCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(cmd.Config, true, true)
	if err != nil {
		return shared.HandleError(err)
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Staging package for {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
		"AppName":   cmd.AppName,
		"OrgName":   cmd.Config.TargetedOrganization().Name,
		"SpaceName": cmd.Config.TargetedSpace().Name,
		"Username":  user.Name,
	})

	buildStream, warningsStream, errStream := cmd.Actor.StagePackage(cmd.PackageGUID)

	var closedBuildStream, closedWarningsStream, closedErrStream bool
	for {
		select {
		case build, ok := <-buildStream:
			if !ok {
				closedBuildStream = true
				break
			}
			cmd.UI.DisplayText("droplet: {{.DropletGUID}}", map[string]interface{}{"DropletGUID": build.Droplet.GUID})
		case warnings, ok := <-warningsStream:
			if !ok {
				closedWarningsStream = true
				break
			}
			cmd.UI.DisplayWarnings(warnings)
		case err, ok := <-errStream:
			if !ok {
				closedErrStream = true
				break
			}
			return shared.HandleError(err)
		}
		if closedBuildStream && closedWarningsStream && closedErrStream {
			break
		}
	}

	cmd.UI.DisplayOK()

	return nil
}
