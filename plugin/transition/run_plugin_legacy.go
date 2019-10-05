// +build !V7

package plugin_transition

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
)

func RunPlugin() {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
}
