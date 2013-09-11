package requirements

import (
	"cf/configuration"
	"cf/terminal"
	"fmt"
)

type SpaceRequirement struct {
	ui     terminal.UI
	config *configuration.Configuration
}

func NewSpaceRequirement(ui terminal.UI, config *configuration.Configuration) SpaceRequirement {
	return SpaceRequirement{ui, config}
}

func (req SpaceRequirement) Execute() (success bool) {
	if !req.config.HasOrganization() {
		message := fmt.Sprintf("No org and space targeted. See '%s' to target an org and space.",
			terminal.Yellow("cf target --o ORGNAME --s SPACENAME"))
		req.ui.Failed(message, nil)
		return false
	}

	if !req.config.HasSpace() {
		message := fmt.Sprintf("No space targeted. Use '%s' to target a space.", terminal.Yellow("cf target -s"))
		req.ui.Failed(message, nil)
		return false
	}

	return true
}
