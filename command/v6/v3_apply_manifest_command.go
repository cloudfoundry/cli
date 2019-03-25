package v6

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v6/shared"
	"code.cloudfoundry.org/cli/util/manifestparser"
	"github.com/cloudfoundry/bosh-cli/director/template"
)

//go:generate counterfeiter . ManifestParser

type ManifestParser interface {
	v3action.ManifestParser
	InterpolateAndParse(pathToManifest string, pathsToVarsFiles []string, vars []template.VarKV) error
}

//go:generate counterfeiter . V3ApplyManifestActor

type V3ApplyManifestActor interface {
	ApplyApplicationManifest(parser v3action.ManifestParser, spaceGUID string) (v3action.Warnings, error)
}

type V3ApplyManifestCommand struct {
	PathToManifest flag.PathWithExistenceCheck `short:"f" description:"Path to app manifest" required:"true"`
	usage          interface{}                 `usage:"CF_NAME v3-apply-manifest -f APP_MANIFEST_PATH"`

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

	ccClient, _, err := shared.NewV3BasedClients(config, ui, true, "")
	if err != nil {
		return err
	}
	cmd.Actor = v3action.NewActor(ccClient, config, nil, nil)
	cmd.Parser = manifestparser.NewParser()

	return nil
}

func (cmd V3ApplyManifestCommand) Execute(args []string) error {
	pathToManifest := string(cmd.PathToManifest)

	cmd.UI.DisplayWarning(command.ExperimentalWarning)

	err := cmd.SharedActor.CheckTarget(true, true)
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

	err = cmd.Parser.InterpolateAndParse(pathToManifest, nil, nil)
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
