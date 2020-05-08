package v7

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/types"
)

//go:generate counterfeiter . LabelUnsetter

type LabelUnsetter interface {
	Execute(resource TargetResource, labels map[string]types.NullString) error
}

type UnsetLabelCommand struct {
	command.BaseCommand

	RequiredArgs    flag.UnsetLabelArgs `positional-args:"yes"`
	usage           interface{}         `usage:"CF_NAME unset-label RESOURCE RESOURCE_NAME KEY...\n\nEXAMPLES:\n   cf unset-label app dora ci_signature_sha2\n   cf unset-label org business pci public-facing\n   cf unset-label buildpack go_buildpack go -s cflinuxfs3\n\nRESOURCES:\n   app\n   buildpack\n   domain\n   org\n   route\n   service-broker\n   service-offering\n   service-plan\n   space\n   stack"`
	relatedCommands interface{}         `related_commands:"labels, set-label"`
	BuildpackStack  string              `long:"stack" short:"s" description:"Specify stack to disambiguate buildpacks with the same name"`
	ServiceBroker   string              `long:"broker" short:"b" description:"Specify a service broker to disambiguate service offerings or service plans with the same name."`
	ServiceOffering string              `long:"offering" short:"e" description:"Specify a service offering to disambiguate service plans with the same name."`

	LabelUnsetter LabelUnsetter
}

func (cmd *UnsetLabelCommand) Setup(config command.Config, ui command.UI) error {
	err := cmd.BaseCommand.Setup(config, ui)
	if err != nil {
		return err
	}

	cmd.LabelUnsetter = &LabelUpdater{
		UI:          ui,
		Config:      config,
		SharedActor: cmd.SharedActor,
		Actor:       cmd.Actor,
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
		ResourceType:    cmd.RequiredArgs.ResourceType,
		ResourceName:    cmd.RequiredArgs.ResourceName,
		BuildpackStack:  cmd.BuildpackStack,
		ServiceBroker:   cmd.ServiceBroker,
		ServiceOffering: cmd.ServiceOffering,
	}
	return cmd.LabelUnsetter.Execute(targetResource, labels)
}
