package v2

import (
	"os"
	"strings"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	oldCmd "code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v2/shared"
)

//go:generate counterfeiter . OrgActor

type OrgActor interface {
	GetOrganizationByName(orgName string) (v2action.Organization, v2action.Warnings, error)
	GetOrganizationSummaryByName(orgName string) (v2action.OrganizationSummary, v2action.Warnings, error)
}

type OrgCommand struct {
	RequiredArgs    flag.Organization `positional-args:"yes"`
	GUID            bool              `long:"guid" description:"Retrieve and display the given org's guid.  All other output for the org is suppressed."`
	usage           interface{}       `usage:"CF_NAME org ORG [--guid]"`
	relatedCommands interface{}       `related_commands:"org-users, orgs"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       OrgActor
}

func (cmd *OrgCommand) Setup(config command.Config, ui command.UI) error {
	cmd.Config = config
	cmd.UI = ui
	cmd.SharedActor = sharedaction.NewActor()

	ccClient, uaaClient, err := shared.NewClients(config, ui)
	if err != nil {
		return err
	}
	cmd.Actor = v2action.NewActor(ccClient, uaaClient)

	return nil
}

func (cmd OrgCommand) Execute(args []string) error {
	if cmd.Config.Experimental() == false {
		oldCmd.Main(os.Getenv("CF_TRACE"), os.Args)
		return nil
	}
	cmd.UI.DisplayText(command.ExperimentalWarning)
	cmd.UI.DisplayNewline()

	err := cmd.SharedActor.CheckTarget(cmd.Config, false, false)
	if err != nil {
		return shared.HandleError(err)
	}

	if cmd.GUID {
		return cmd.displayOrgGUID()
	} else {
		return cmd.displayOrgSummary()
	}
}

func (cmd OrgCommand) displayOrgGUID() error {
	org, warnings, err := cmd.Actor.GetOrganizationByName(cmd.RequiredArgs.Organization)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return shared.HandleError(err)
	}

	cmd.UI.DisplayText(org.GUID)

	return nil
}

func (cmd OrgCommand) displayOrgSummary() error {
	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return shared.HandleError(err)
	}

	cmd.UI.DisplayTextWithFlavor(
		"Getting info for org {{.OrgName}} as {{.Username}}...",
		map[string]interface{}{
			"OrgName":  cmd.RequiredArgs.Organization,
			"Username": user.Name,
		})
	cmd.UI.DisplayNewline()

	orgSummary, warnings, err := cmd.Actor.GetOrganizationSummaryByName(cmd.RequiredArgs.Organization)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return shared.HandleError(err)
	}

	table := [][]string{
		{cmd.UI.TranslateText("name:"), orgSummary.Name},
		{cmd.UI.TranslateText("domains:"), strings.Join(orgSummary.DomainNames, ", ")},
		{cmd.UI.TranslateText("quota:"), orgSummary.QuotaName},
		{cmd.UI.TranslateText("spaces:"), strings.Join(orgSummary.SpaceNames, ", ")},
	}
	cmd.UI.DisplayTable("", table, 3)

	return nil
}
