package v6

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v6/shared"
	"code.cloudfoundry.org/cli/util/ui"
)

//go:generate counterfeiter . DomainsActor

type DomainsActor interface {
	GetDomains(orgGUID string) ([]v2action.Domain, v2action.Warnings, error)
}

type DomainsCommand struct {
	usage           interface{} `usage:"CF_NAME domains"`
	relatedCommands interface{} `related_commands:"router-groups, create-route, routes"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       DomainsActor
}

func (cmd *DomainsCommand) Setup(config command.Config, ui command.UI) error {
	cmd.Config = config
	cmd.UI = ui
	cmd.SharedActor = sharedaction.NewActor(config)
	ccClient, uaaClient, err := shared.NewClients(config, ui, true)
	if err != nil {
		return err
	}
	cmd.Actor = v2action.NewActor(ccClient, uaaClient, config)

	return nil
}

func (cmd DomainsCommand) Execute(args []string) error {
	if len(args) > 0 {
		return translatableerror.TooManyArgumentsError{
			ExtraArgument: args[0],
		}
	}

	err := cmd.SharedActor.CheckTarget(true, false)
	if err != nil {
		return err
	}

	org := cmd.Config.TargetedOrganization()

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Getting domains in org {{.CurrentOrg}} as {{.CurrentUser}}...", map[string]interface{}{
		"CurrentUser": user.Name,
		"CurrentOrg":  org.Name,
	})

	domains, warnings, err := cmd.Actor.GetDomains(org.GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	table := [][]string{
		{
			cmd.UI.TranslateText("name"),
			cmd.UI.TranslateText("status"),
			cmd.UI.TranslateText("type"),
			cmd.UI.TranslateText("details"),
		},
	}

	for _, domain := range domains {
		internalMark := ""
		if domain.Internal {
			internalMark = "internal"
		}
		table = append(
			table,
			[]string{
				domain.Name,
				string(domain.Type),
				string(domain.RouterGroupType),
				internalMark,
			},
		)
	}

	cmd.UI.DisplayTableWithHeader("", table, ui.DefaultTableSpacePadding)

	return err
}
