package v2

import (
	"fmt"
	"sort"
	"strings"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v2/shared"
	sharedV3 "code.cloudfoundry.org/cli/command/v3/shared"
)

//go:generate counterfeiter . OrgActor

type OrgActor interface {
	GetOrganizationByName(orgName string) (v2action.Organization, v2action.Warnings, error)
	GetOrganizationSummaryByName(orgName string) (v2action.OrganizationSummary, v2action.Warnings, error)
}

//go:generate counterfeiter . OrgActorV3

type OrgActorV3 interface {
	GetIsolationSegmentsByOrganization(orgName string) ([]v3action.IsolationSegment, v3action.Warnings, error)
	CloudControllerAPIVersion() string
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
	ActorV3     OrgActorV3
}

func (cmd *OrgCommand) Setup(config command.Config, ui command.UI) error {
	cmd.Config = config
	cmd.UI = ui
	cmd.SharedActor = sharedaction.NewActor(config)

	ccClient, uaaClient, err := shared.NewClients(config, ui, true)
	if err != nil {
		return err
	}
	cmd.Actor = v2action.NewActor(ccClient, uaaClient, config)

	ccClientV3, _, err := sharedV3.NewClients(config, ui, true)
	if err != nil {
		if _, ok := err.(translatableerror.V3APIDoesNotExistError); !ok {
			return err
		}
	} else {
		cmd.ActorV3 = v3action.NewActor(ccClientV3, config, nil, nil)
	}

	return nil
}

func (cmd OrgCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
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

	if cmd.ActorV3 != nil {
		apiCheck := command.MinimumAPIVersionCheck(cmd.ActorV3.CloudControllerAPIVersion(), ccversion.MinVersionIsolationSegmentV3)
		if apiCheck == nil {
			isolationSegments, v3Warnings, err := cmd.ActorV3.GetIsolationSegmentsByOrganization(orgSummary.GUID)
			cmd.UI.DisplayWarnings(v3Warnings)
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
		}
	}

	cmd.UI.DisplayKeyValueTable("", table, 3)

	return nil
}
