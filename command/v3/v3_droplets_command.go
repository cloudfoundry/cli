package v3

import (
	"net/http"
	"strings"
	"time"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v3/shared"
	"code.cloudfoundry.org/cli/util/ui"
)

//go:generate counterfeiter . V3DropletsActor

type V3DropletsActor interface {
	CloudControllerAPIVersion() string
	GetApplicationDroplets(appName string, spaceGUID string) ([]v3action.Droplet, v3action.Warnings, error)
}

type V3DropletsCommand struct {
	RequiredArgs flag.AppName `positional-args:"yes"`
	usage        interface{}  `usage:"CF_NAME v3-droplets APP_NAME"`

	UI          command.UI
	Config      command.Config
	Actor       V3DropletsActor
	SharedActor command.SharedActor
}

func (cmd *V3DropletsCommand) Setup(config command.Config, ui command.UI) error {
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

func (cmd V3DropletsCommand) Execute(args []string) error {
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

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Listing droplets of app {{.AppName}} in org {{.CurrentOrg}} / space {{.CurrentSpace}} as {{.CurrentUser}}...", map[string]interface{}{
		"AppName":      cmd.RequiredArgs.AppName,
		"CurrentSpace": cmd.Config.TargetedSpace().Name,
		"CurrentOrg":   cmd.Config.TargetedOrganization().Name,
		"CurrentUser":  user.Name,
	})
	cmd.UI.DisplayNewline()

	droplets, warnings, err := cmd.Actor.GetApplicationDroplets(cmd.RequiredArgs.AppName, cmd.Config.TargetedSpace().GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	if len(droplets) == 0 {
		cmd.UI.DisplayText("No droplets found")
		return nil
	}

	table := [][]string{
		{
			cmd.UI.TranslateText("guid"),
			cmd.UI.TranslateText("state"),
			cmd.UI.TranslateText("created"),
		},
	}

	for _, droplet := range droplets {
		t, err := time.Parse(time.RFC3339, droplet.CreatedAt)
		if err != nil {
			return err
		}

		table = append(table, []string{
			droplet.GUID,
			cmd.UI.TranslateText(strings.ToLower(string(droplet.State))),
			cmd.UI.UserFriendlyDate(t),
		})
	}

	cmd.UI.DisplayTableWithHeader("", table, ui.DefaultTableSpacePadding)

	return nil
}
