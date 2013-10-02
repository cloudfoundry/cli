package requirements

import (
	"cf"
	"cf/configuration"
	"cf/terminal"
	"fmt"
)

type TargetedOrgRequirement interface {
	Requirement
	GetOrganization() cf.Organization
}

type TargetedOrgApiRequirement struct {
	ui     terminal.UI
	config configuration.Configuration
}

func NewTargetedOrgRequirement(ui terminal.UI, config configuration.Configuration) TargetedOrgRequirement {
	return TargetedOrgApiRequirement{ui, config}
}

func (req TargetedOrgApiRequirement) Execute() (success bool) {
	if !req.config.HasOrganization() {
		message := fmt.Sprintf("No org targeted. See '%s' to target an org.",
			terminal.CommandColor(cf.Name+" target -o ORGNAME"))
		req.ui.Failed(message)
		return false
	}

	return true
}

func (req TargetedOrgApiRequirement) GetOrganization() (org cf.Organization) {
	return req.config.Organization
}
