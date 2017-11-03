package v3

import (
	"net/http"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v3/shared"
)

//go:generate counterfeiter . V3DeleteActor

type V3DeleteActor interface {
	CloudControllerAPIVersion() string
	DeleteApplicationByNameAndSpace(name string, spaceGUID string) (v3action.Warnings, error)
}

type V3DeleteCommand struct {
	RequiredArgs flag.AppName `positional-args:"yes"`
	Force        bool         `short:"f" description:"Force deletion without confirmation"`
	usage        interface{}  `usage:"CF_NAME v3-delete APP_NAME [-f]"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       V3DeleteActor
}

func (cmd *V3DeleteCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor(config)

	ccClient, _, err := shared.NewClients(config, ui, true)
	if err != nil {
		if v3Err, ok := err.(ccerror.V3UnexpectedResponseError); ok && v3Err.ResponseCode == http.StatusNotFound {
			return translatableerror.MinimumAPIVersionNotMetError{MinimumVersion: ccversion.MinVersionV3}
		}

		return err
	}
	cmd.Actor = v3action.NewActor(ccClient, config, nil, nil)

	return nil
}

func (cmd V3DeleteCommand) Execute(args []string) error {
	cmd.UI.DisplayText(command.ExperimentalWarning)
	cmd.UI.DisplayNewline()

	err := command.MinimumAPIVersionCheck(cmd.Actor.CloudControllerAPIVersion(), ccversion.MinVersionV3)
	if err != nil {
		return err
	}

	err = cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	currentUser, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	if !cmd.Force {
		response, promptErr := cmd.UI.DisplayBoolPrompt(false, "Really delete the app {{.AppName}}?", map[string]interface{}{
			"AppName": cmd.RequiredArgs.AppName,
		})

		if promptErr != nil {
			return promptErr
		}

		if !response {
			cmd.UI.DisplayText("Delete cancelled")
			return nil
		}
	}

	cmd.UI.DisplayTextWithFlavor("Deleting app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
		"AppName":   cmd.RequiredArgs.AppName,
		"OrgName":   cmd.Config.TargetedOrganization().Name,
		"SpaceName": cmd.Config.TargetedSpace().Name,
		"Username":  currentUser.Name,
	})

	warnings, err := cmd.Actor.DeleteApplicationByNameAndSpace(cmd.RequiredArgs.AppName, cmd.Config.TargetedSpace().GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		switch err.(type) {
		case actionerror.ApplicationNotFoundError:
			cmd.UI.DisplayTextWithFlavor("App {{.AppName}} does not exist", map[string]interface{}{
				"AppName": cmd.RequiredArgs.AppName,
			})
		default:
			return err
		}
	}

	cmd.UI.DisplayOK()

	return nil
}
