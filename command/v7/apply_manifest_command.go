package v7

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/cli/util/manifestparser"
)

//go:generate counterfeiter . ApplyManifestActor
type ApplyManifestActor interface {
	SetSpaceManifest(spaceGUID string, rawManifest []byte, noRoute bool) (v7action.Warnings, error)
}

type ApplyManifestCommand struct {
	PathToManifest  flag.PathWithExistenceCheck `short:"f" description:"Path to app manifest" required:"true"`
	usage           interface{}                 `usage:"CF_NAME apply-manifest -f APP_MANIFEST_PATH"`
	relatedCommands interface{}                 `related_commands:"create-app, create-app-manifest, push"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       ApplyManifestActor
	Parser      ManifestParser
}

func (cmd *ApplyManifestCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor(config)

	ccClient, _, err := shared.NewClients(config, ui, true, "")
	if err != nil {
		return err
	}
	cmd.Actor = v7action.NewActor(ccClient, config, nil, nil)
	cmd.Parser = manifestparser.NewParser()

	return nil
}

func (cmd ApplyManifestCommand) Execute(args []string) error {
	pathToManifest := string(cmd.PathToManifest)

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

	err = cmd.Parser.InterpolateAndParse(pathToManifest, nil, nil, "")
	if err != nil {
		return err
	}

	warnings, err := cmd.Actor.SetSpaceManifest(cmd.Config.TargetedSpace().GUID, cmd.Parser.FullRawManifest(), false)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()

	return nil
}
