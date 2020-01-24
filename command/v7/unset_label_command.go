package v7

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/cli/types"

	//"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/clock"
)

//go:generate counterfeiter . LabelUnsetter

type LabelUnsetter interface {
	Execute(resource TargetResource, labels map[string]types.NullString) error
}

type UnsetLabelCommand struct {
	RequiredArgs    flag.UnsetLabelArgs `positional-args:"yes"`
	usage           interface{}         `usage:"CF_NAME unset-label RESOURCE RESOURCE_NAME KEY...\n\nEXAMPLES:\n   cf unset-label app dora ci_signature_sha2\n   cf unset-label org business pci public-facing\n   cf unset-label buildpack go_buildpack go -s cflinuxfs3\n\nRESOURCES:\n   app\n   buildpack\n   domain\n   org\n   route\n   service-broker\n   service-offering\n   space\n   stack"`
	relatedCommands interface{}         `related_commands:"labels, set-label"`
	BuildpackStack  string              `long:"stack" short:"s" description:"Specify stack to disambiguate buildpacks with the same name"`
	ServiceBroker   string              `long:"broker" short:"b" description:"Specify a service broker to disambiguate service offerings with the same name."`

	LabelUnsetter LabelUnsetter
}

func (cmd *UnsetLabelCommand) Setup(config command.Config, ui command.UI) error {
	sharedActor := sharedaction.NewActor(config)
	ccClient, _, err := shared.GetNewClientsAndConnectToCF(config, ui, "")
	if err != nil {
		return err
	}
	actor := v7action.NewActor(ccClient, config, nil, nil, clock.NewClock())

	cmd.LabelUnsetter = &LabelUpdater{
		UI:          ui,
		Config:      config,
		SharedActor: sharedActor,
		Actor:       actor,
		Action:      Unset,
	}
	return nil
}

func (cmd UnsetLabelCommand) Execute(args []string) error {
	labels := make(map[string]types.NullString)
	for _, value := range cmd.RequiredArgs.LabelKeys {
		labels[value] = types.NewNullString()
	}

	targetResource := TargetResource{
		ResourceType:   cmd.RequiredArgs.ResourceType,
		ResourceName:   cmd.RequiredArgs.ResourceName,
		BuildpackStack: cmd.BuildpackStack,
		ServiceBroker:  cmd.ServiceBroker,
	}
	return cmd.LabelUnsetter.Execute(targetResource, labels)
}
