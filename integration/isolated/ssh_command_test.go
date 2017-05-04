package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("ssh command", func() {
	Describe("help", func() {
		Context("when --help flag is provided", func() {
			It("displays command usage to output", func() {
				session := helpers.CF("ssh", "--help")
				Eventually(session.Out).Should(Say("NAME:"))
				Eventually(session.Out).Should(Say("ssh - SSH to an application container instance"))
				Eventually(session.Out).Should(Say("USAGE:"))
				Eventually(session.Out).Should(Say("cf ssh APP_NAME \\[-i app-instance-index\\] \\[-c command\\] \\[-L \\[bind_address:\\]port:host:hostport\\] \\[--skip-host-validation\\] \\[--skip-remote-execution\\] \\[--request-pseudo-tty\\] \\[--force-pseudo-tty\\] \\[--disable-pseudo-tty\\]"))
				Eventually(session.Out).Should(Say("OPTIONS:"))
				Eventually(session.Out).Should(Say("--app-instance-index, -i\\s+Application instance index \\(Default: 0\\)"))
				Eventually(session.Out).Should(Say("--command, -c\\s+Command to run\\. This flag can be defined more than once\\."))
				Eventually(session.Out).Should(Say("--disable-pseudo-tty, -T\\s+Disable pseudo-tty allocation"))
				Eventually(session.Out).Should(Say("--force-pseudo-tty\\s+Force pseudo-tty allocation"))
				Eventually(session.Out).Should(Say("-L\\s+Local port forward specification\\. This flag can be defined more than once\\."))
				Eventually(session.Out).Should(Say("--request-pseudo-tty, -t\\s+Request pseudo-tty allocation"))
				Eventually(session.Out).Should(Say("--skip-host-validation, -k\\s+Skip host key validation"))
				Eventually(session.Out).Should(Say("--skip-remote-execution, -N\\s+Do not execute a remote command"))
				Eventually(session.Out).Should(Say("SEE ALSO:"))
				Eventually(session.Out).Should(Say("allow-space-ssh, enable-ssh, space-ssh-allowed, ssh-code, ssh-enabled"))
				Eventually(session).Should(Exit(0))
			})
		})
	})
})
