package v7

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
)

//go:generate counterfeiter . UnsharePrivateDomainActor

type UnsharePrivateDomainActor interface {
	UnsharePrivateDomain(domainName string, orgName string) (v7action.Warnings, error)
}

type UnsharePrivateDomainCommand struct {
	RequiredArgs    flag.OrgDomain `positional-args:"yes"`
	usage           interface{}    `usage:"CF_NAME unshare-private-domain ORG DOMAIN"`
	relatedCommands interface{}    `related_commands:"delete-private-domain, domains"`

	UI          command.UI
	Config      command.Config
	Actor       UnsharePrivateDomainActor
	SharedActor command.SharedActor
}

func (cmd *UnsharePrivateDomainCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	sharedActor := sharedaction.NewActor(config)
	cmd.SharedActor = sharedActor

	ccClient, uaaClient, err := shared.NewClients(config, ui, true, "")
	if err != nil {
		return err
	}
	cmd.Actor = v7action.NewActor(ccClient, config, sharedActor, uaaClient)
	return nil
}

func (cmd UnsharePrivateDomainCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	domain := cmd.RequiredArgs.Domain
	orgName := cmd.RequiredArgs.Organization

	cmd.UI.DisplayTextWithFlavor("Warning: org {{.Organization}} will no longer be able to access private domain {{.Domain}}",
		map[string]interface{}{
			"Domain":       domain,
			"Organization": orgName,
		})

	response, err := cmd.UI.DisplayBoolPrompt(false,
		"Really unshare private domain {{.Domain}}?",
		map[string]interface{}{
			"Domain": domain,
		},
	)

	if err != nil {
		return err
	}

	if !response {
		cmd.UI.DisplayTextWithFlavor("Domain {{.Domain}} has not been unshared from organization {{.Organization}}",
			map[string]interface{}{
				"Domain":       domain,
				"Organization": orgName,
			})
		return nil
	}

	cmd.UI.DisplayTextWithFlavor("Unsharing domain {{.Domain}} from org {{.Organization}} as {{.User}}...",
		map[string]interface{}{
			"Domain":       domain,
			"User":         user.Name,
			"Organization": orgName,
		})

	warnings, err := cmd.Actor.UnsharePrivateDomain(domain, orgName)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()

	return nil
}
