package v3

import (
	"net/http"
	"strings"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/translatableerror"
	sharedV2 "code.cloudfoundry.org/cli/command/v2/shared"
	"code.cloudfoundry.org/cli/command/v3/shared"
	"code.cloudfoundry.org/cli/util/ui"
)

//go:generate counterfeiter . V3AppsActor

type V3AppsActor interface {
	CloudControllerAPIVersion() string
	GetApplicationsWithProcessesBySpace(spaceGUID string) ([]v3action.ApplicationWithProcessSummary, v3action.Warnings, error)
}

type V3AppsCommand struct {
	usage interface{} `usage:"CF_NAME v3-apps"`

	UI              command.UI
	Config          command.Config
	Actor           V3AppsActor
	V2AppRouteActor shared.V2AppRouteActor
	SharedActor     command.SharedActor
}

func (cmd *V3AppsCommand) Setup(config command.Config, ui command.UI) error {
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

	ccClientV2, uaaClientV2, err := sharedV2.NewClients(config, ui, true)
	if err != nil {
		return err
	}

	cmd.V2AppRouteActor = v2action.NewActor(ccClientV2, uaaClientV2, config)

	return nil
}

func (cmd V3AppsCommand) Execute(args []string) error {
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

	cmd.UI.DisplayTextWithFlavor("Getting apps in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
		"OrgName":   cmd.Config.TargetedOrganization().Name,
		"SpaceName": cmd.Config.TargetedSpace().Name,
		"Username":  user.Name,
	})
	cmd.UI.DisplayNewline()

	summaries, warnings, err := cmd.Actor.GetApplicationsWithProcessesBySpace(cmd.Config.TargetedSpace().GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	if len(summaries) == 0 {
		cmd.UI.DisplayText("No apps found")
		return nil
	}

	table := [][]string{
		{
			cmd.UI.TranslateText("name"),
			cmd.UI.TranslateText("requested state"),
			cmd.UI.TranslateText("processes"),
			cmd.UI.TranslateText("routes"),
		},
	}

	for _, summary := range summaries {
		var routesList string
		if len(summary.ProcessSummaries) > 0 {
			routes, warnings, err := cmd.V2AppRouteActor.GetApplicationRoutes(summary.GUID)
			cmd.UI.DisplayWarnings(warnings)
			if err != nil {
				return err
			}
			routesList = routes.Summary()
		}

		table = append(table, []string{
			summary.Name,
			cmd.UI.TranslateText(strings.ToLower(string(summary.State))),
			summary.ProcessSummaries.String(),
			routesList,
		})
	}

	cmd.UI.DisplayTableWithHeader("", table, ui.DefaultTableSpacePadding)

	return nil
}
