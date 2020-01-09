package v7

import (
	"fmt"
	"strings"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/clock"
)

//go:generate counterfeiter . LabelSetter

type LabelSetter interface {
	Execute(resource TargetResource, labels map[string]types.NullString) error
}

type SetLabelCommand struct {
	RequiredArgs    flag.SetLabelArgs `positional-args:"yes"`
	usage           interface{}       `usage:"CF_NAME set-label RESOURCE RESOURCE_NAME KEY=VALUE...\n\nEXAMPLES:\n   cf set-label app dora env=production\n   cf set-label org business pci=true public-facing=false\n   cf set-label buildpack go_buildpack go=1.12 -s cflinuxfs3\n\nRESOURCES:\n   app\n   buildpack\n   domain\n   org\n   route\n   service-broker\n   space\n   stack"`
	relatedCommands interface{}       `related_commands:"labels, unset-label"`
	BuildpackStack  string            `long:"stack" short:"s" description:"Specify stack to disambiguate buildpacks with the same name"`
	LabelSetter     LabelSetter
}

func (cmd *SetLabelCommand) Setup(config command.Config, ui command.UI) error {
	sharedActor := sharedaction.NewActor(config)
	ccClient, _, err := shared.GetNewClientsAndConnectToCF(config, ui, "")
	if err != nil {
		return err
	}
	actor := v7action.NewActor(ccClient, config, nil, nil, clock.NewClock())

	cmd.LabelSetter = &LabelUpdater{
		UI:          ui,
		Config:      config,
		SharedActor: sharedActor,
		Actor:       actor,
		Action:      Set,
	}
	return nil
}

func (cmd SetLabelCommand) Execute(args []string) error {
	targetResource := TargetResource{
		ResourceType:   cmd.RequiredArgs.ResourceType,
		ResourceName:   cmd.RequiredArgs.ResourceName,
		BuildpackStack: cmd.BuildpackStack,
	}

	labels := make(map[string]types.NullString)
	for _, label := range cmd.RequiredArgs.Labels {
		parts := strings.SplitN(label, "=", 2)
		if len(parts) < 2 {
			return fmt.Errorf("Metadata error: no value provided for label '%s'", label)
		}
		labels[parts[0]] = types.NewNullString(parts[1])
	}

	return cmd.LabelSetter.Execute(targetResource, labels)
}
