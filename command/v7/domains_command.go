package v7

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/cli/util/sorting"
	"code.cloudfoundry.org/cli/util/ui"
	"sort"
)

//go:generate counterfeiter . DomainsActor

type DomainsActor interface {
	GetOrganizationDomains(string) ([]v7action.Domain, v7action.Warnings, error)
}

type DomainsCommand struct {
	usage           interface{} `usage:"CF_NAME domains"`
	relatedCommands interface{} `related_commands:"create-route, routes, create-shared-domain, create-private-domain"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       DomainsActor
}

func (cmd *DomainsCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor(config)

	ccClient, _, err := shared.NewClients(config, ui, true, "")
	if err != nil {
		return err
	}
	cmd.Actor = v7action.NewActor(ccClient, config, nil, nil)

	return nil
}

func (cmd DomainsCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, false)
	if err != nil {
		return err
	}

	currentUser, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	targetedOrg := cmd.Config.TargetedOrganization()
	cmd.UI.DisplayTextWithFlavor("Getting domains in org {{.CurrentOrg}} as {{.CurrentUser}}...\n", map[string]interface{}{
		"CurrentOrg":  targetedOrg.Name,
		"CurrentUser": currentUser.Name,
	})

	domains, warnings, err := cmd.Actor.GetOrganizationDomains(targetedOrg.GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	sort.Slice(domains, func(i, j int) bool { return sorting.LessIgnoreCase(domains[i].Name, domains[j].Name) })

	if len(domains) > 0 {
		cmd.displayDomainsTable(domains)
	} else {
		cmd.UI.DisplayText("No domains found.")
	}
	return nil
}

func (cmd DomainsCommand) displayDomainsTable(domains []v7action.Domain) {
	var domainsTable = [][]string{
		{
			cmd.UI.TranslateText("name"),
			cmd.UI.TranslateText("availability"),
			cmd.UI.TranslateText("internal"),
		},
	}

	for _, domain := range domains {
		var availability string
		var internal string

		if domain.Shared() {
			availability = cmd.UI.TranslateText("shared")
		} else {
			availability = cmd.UI.TranslateText("private")
		}

		if domain.Internal.IsSet && domain.Internal.Value {
			internal = cmd.UI.TranslateText("true")
		}

		domainsTable = append(domainsTable, []string{
			domain.Name,
			availability,
			internal,
		})
	}

	cmd.UI.DisplayTableWithHeader("", domainsTable, ui.DefaultTableSpacePadding)

}
