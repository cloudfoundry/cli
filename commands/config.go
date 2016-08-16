package commands

import "code.cloudfoundry.org/cli/utils/config"

type Config interface {
	ColorEnabled() config.ColorSetting
}
