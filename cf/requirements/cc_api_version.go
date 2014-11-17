package requirements

import (
	"strconv"
	"strings"

	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type CCApiVersionRequirement struct {
	ui          terminal.UI
	config      core_config.Reader
	commandName string
	major       int
	minor       int
	patch       int
}

func NewCCApiVersionRequirement(ui terminal.UI, config core_config.Reader, commandName string, major, minor, patch int) CCApiVersionRequirement {
	return CCApiVersionRequirement{ui, config, commandName, major, minor, patch}
}

func (req CCApiVersionRequirement) Execute() bool {
	versions := strings.Split(req.config.ApiVersion(), ".")

	if len(versions) != 3 {
		return true
	}

	majorStr := versions[0]
	major, err := strconv.Atoi(majorStr)
	if err != nil {
		return true
	}

	minorStr := versions[1]
	minor, err := strconv.Atoi(minorStr)
	if err != nil {
		return true
	}

	patchStr := versions[2]
	patch, err := strconv.Atoi(patchStr)
	if err != nil {
		return true
	}

	if major > req.major {
		return true
	} else if major < req.major {
		return false
	}

	if minor > req.minor {
		return true
	} else if minor < req.minor {
		return false
	}

	if patch >= req.patch {
		return true
	}

	req.ui.Say(terminal.FailureColor(T("FAILED")))
	req.ui.Say(T("Current CF CLI version {{.Version}}", map[string]interface{}{"Version": cf.Version}))
	req.ui.Say(T("Current CF API version {{.ApiVersion}}", map[string]interface{}{"ApiVersion": req.config.ApiVersion()}))
	req.ui.Say(T("To use the {{.CommandName}} feature, you need to upgrade the CF API to at least {{.MinApiVersionMajor}}.{{.MinApiVersionMinor}}.{{.MinApiVersionPatch}}",
		map[string]interface{}{
			"CommandName":        req.commandName,
			"MinApiVersionMajor": req.major,
			"MinApiVersionMinor": req.minor,
			"MinApiVersionPatch": req.patch,
		}))

	return false
}
