package requirements

import (
	"cf/configuration"
	"cf/terminal"
	"fmt"
)

type OrgRequirement struct {
	ui     terminal.UI
	config *configuration.Configuration
}

func NewOrgRequirement(ui terminal.UI, config *configuration.Configuration) OrgRequirement {
	return OrgRequirement{ui, config}
}

func (req OrgRequirement) Execute() (success bool) {
	if !req.config.HasOrganization() {
		message := fmt.Sprintf("No org targeted. See '%s' to target an org.",
			terminal.Yellow("cf target --o ORGNAME"))
		req.ui.Failed(message)
		return false
	}

	return true
}
