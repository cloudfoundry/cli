package requirements

import (
	"cf/configuration"
	"cf/terminal"
	"errors"
	"fmt"
)

type SpaceRequirement struct {
	ui     terminal.UI
	config *configuration.Configuration
}

func NewSpaceRequirement(ui terminal.UI, config *configuration.Configuration) SpaceRequirement {
	return SpaceRequirement{ui, config}
}

func (req SpaceRequirement) Execute() (err error) {
	if !req.config.HasOrganization() {
		message := fmt.Sprintf("No org and space targeted. Use '%s' to target an org and space.", terminal.Yellow("cf target -o"))
		req.ui.Failed(message, nil)
		err = errors.New("No org and space targeted")
		return
	}

	if !req.config.HasSpace() {
		message := fmt.Sprintf("No space targeted. Use '%s' to target a space.", terminal.Yellow("cf target -s"))
		req.ui.Failed(message, nil)
		err = errors.New("No space targeted")
		return
	}

	return
}
