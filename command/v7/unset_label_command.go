package v7

import (
	"code.cloudfoundry.org/cli/v8/command"
	"code.cloudfoundry.org/cli/v8/command/flag"
	"code.cloudfoundry.org/cli/v8/types"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . LabelUnsetter

type LabelUnsetter interface {
	Execute(resource TargetResource, labels map[string]types.NullString) error
}

type UnsetLabelCommand struct {
	BaseCommand

	RequiredArgs    flag.UnsetLabelArgs `positional-args:"yes"`
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

func (cmd UnsetLabelCommand) Usage() string {
	return `CF_NAME unset-label RESOURCE RESOURCE_NAME KEY...`
}

func (cmd UnsetLabelCommand) Examples() string {
	return `
cf unset-label app dora ci_signature_sha2
cf unset-label org business pci public-facing
cf unset-label buildpack go_buildpack go -s cflinuxfs4`
}

func (cmd UnsetLabelCommand) Resources() string {
	return `
app
buildpack
domain
org
route
service-broker
service-instance
service-offering
service-plan
space
stack`
}
