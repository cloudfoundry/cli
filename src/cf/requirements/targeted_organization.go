package requirements

import (
	"cf"
	"cf/configuration"
	"cf/terminal"
	"fmt"
)

type TargetedOrgRequirement struct {
	ui     terminal.UI
	config configuration.Configuration
}

func NewTargetedOrgRequirement(ui terminal.UI, config configuration.Configuration) TargetedOrgRequirement {
	return TargetedOrgRequirement{ui, config}
}

func (req TargetedOrgRequirement) Execute() (success bool) {
	if !req.config.HasOrganization() {
		message := fmt.Sprintf("No org targeted. See '%s' to target an org.",
			terminal.CommandColor(cf.Name+" target -o ORGNAME"))
		req.ui.Failed(message)
		return false
	}

	return true
}
