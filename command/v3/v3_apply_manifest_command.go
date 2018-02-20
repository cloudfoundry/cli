package v3

import (
	"net/http"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v3/shared"
	"code.cloudfoundry.org/cli/util/manifestparser"
)

//go:generate counterfeiter . ManifestParser

type ManifestParser interface {
	AppNames() []string
	Parse(manifestPath string) error
	RawManifest(name string) ([]byte, error)
}

//go:generate counterfeiter . V3ApplyManifestActor

type V3ApplyManifestActor interface {
	CloudControllerAPIVersion() string
	ApplyApplicationManifest(parser v3action.ManifestParser, spaceGUID string) (v3action.Warnings, error)
}

type V3ApplyManifestCommand struct {
	PathToManifest flag.PathWithExistenceCheck `short:"f" description:"Path to app manifest" required:"true"`
	usage          interface{}                 `usage:"CF_NAME v3-apply-manifest -f APP_MANIFESTPATH"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       V3ApplyManifestActor
	Parser      ManifestParser
}

func (cmd *V3ApplyManifestCommand) Setup(config command.Config, ui command.UI) error {
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
	cmd.Parser = manifestparser.NewParser()

	return nil
}

func (cmd V3ApplyManifestCommand) Execute(args []string) error {
	pathToManifest := string(cmd.PathToManifest)

	cmd.UI.DisplayText(command.ExperimentalWarning)
	cmd.UI.DisplayNewline()

	// TODO: Update minimum API version when apply-manifest is complete in V3 API
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

	cmd.UI.DisplayTextWithFlavor("Applying manifest {{.ManifestPath}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
		"ManifestPath": pathToManifest,
		"OrgName":      cmd.Config.TargetedOrganization().Name,
		"SpaceName":    cmd.Config.TargetedSpace().Name,
		"Username":     user.Name,
	})

	err = cmd.Parser.Parse(pathToManifest)
	if err != nil {
		return err
	}

	warnings, err := cmd.Actor.ApplyApplicationManifest(cmd.Parser, cmd.Config.TargetedSpace().GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()

	return nil
}
