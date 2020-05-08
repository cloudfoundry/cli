package v7

import (
	"fmt"
	"sort"
	"strings"

	"code.cloudfoundry.org/cli/command/flag"
)

type OrgCommand struct {
	command.BaseCommand

	RequiredArgs    flag.Organization `positional-args:"yes"`
	GUID            bool              `long:"guid" description:"Retrieve and display the given org's guid.  All other output for the org is suppressed."`
	usage           interface{}       `usage:"CF_NAME org ORG [--guid]"`
	relatedCommands interface{}       `related_commands:"org-users, orgs"`
}

func (cmd OrgCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	if cmd.GUID {
		return cmd.displayOrgGUID()
	}

	return cmd.displayOrgSummary()
}

func (cmd OrgCommand) displayOrgGUID() error {
	org, warnings, err := cmd.Actor.GetOrganizationByName(cmd.RequiredArgs.Organization)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayText(org.GUID)

	return nil
}

func (cmd OrgCommand) displayOrgSummary() error {
	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
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
		return err
	}

	table := [][]string{
		{cmd.UI.TranslateText("name:"), orgSummary.Name},
		{cmd.UI.TranslateText("domains:"), strings.Join(orgSummary.DomainNames, ", ")},
		{cmd.UI.TranslateText("quota:"), orgSummary.QuotaName},
		{cmd.UI.TranslateText("spaces:"), strings.Join(orgSummary.SpaceNames, ", ")},
	}

	isolationSegments, v7Warnings, err := cmd.Actor.GetIsolationSegmentsByOrganization(orgSummary.GUID)
	cmd.UI.DisplayWarnings(v7Warnings)
	if err != nil {
		return err
	}

	isolationSegmentNames := []string{}
	for _, iso := range isolationSegments {
		if iso.GUID == orgSummary.DefaultIsolationSegmentGUID {
			isolationSegmentNames = append(isolationSegmentNames, fmt.Sprintf("%s (%s)", iso.Name, cmd.UI.TranslateText("default")))
		} else {
			isolationSegmentNames = append(isolationSegmentNames, iso.Name)
		}
	}

	sort.Strings(isolationSegmentNames)
	table = append(table, []string{cmd.UI.TranslateText("isolation segments:"), strings.Join(isolationSegmentNames, ", ")})

	cmd.UI.DisplayKeyValueTable("", table, 3)

	return nil
}
