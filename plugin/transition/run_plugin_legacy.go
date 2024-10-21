package plugin_transition

import (
	"os"

	"code.cloudfoundry.org/cli/v7/cf/cmd"
	"code.cloudfoundry.org/cli/v7/command"
	"code.cloudfoundry.org/cli/v7/util/configv3"
)

func RunPlugin(plugin configv3.Plugin, _ command.UI) error {
	// ugly workaround to maintain v7 api in v7 main
	plugin = configv3.Plugin{}
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
