package requirements

import (
	"fmt"
	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/configuration"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type TargetedSpaceRequirement struct {
	ui     terminal.UI
	config configuration.Reader
}

func NewTargetedSpaceRequirement(ui terminal.UI, config configuration.Reader) TargetedSpaceRequirement {
	return TargetedSpaceRequirement{ui, config}
}

func (req TargetedSpaceRequirement) Execute() (success bool) {
	if !req.config.HasOrganization() {
		message := fmt.Sprintf(T("No org and space targeted, use '{{.Command}}' to target an org and space", map[string]interface{}{"Command": terminal.CommandColor(cf.Name() + " target -o ORG -s SPACE")}))
		req.ui.Failed(message)
		return false
	}

	if !req.config.HasSpace() {
		message := fmt.Sprintf(T("No space targeted, use '{{.Command}}' to target a space", map[string]interface{}{"Command": terminal.CommandColor("cf target -s")}))
		req.ui.Failed(message)
		return false
	}

	return true
}
