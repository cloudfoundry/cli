package push

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("help text", func() {
	Context("when --help flag is set", func() {
		It("Displays command usage to output", func() {
			session := helpers.CF(PushCommandName, "--help")
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Say("%s - Push a new app or sync changes to an existing app", PushCommandName))
			Eventually(session).Should(Say("USAGE:"))
			Eventually(session).Should(Say("cf %s APP_NAME \\[-b BUILDPACK_NAME\\] \\[-c COMMAND\\] \\[-f MANIFEST_PATH \\| --no-manifest\\] \\[--no-start\\]", PushCommandName))
			Eventually(session).Should(Say("cf %s APP_NAME --docker-image \\[REGISTRY_HOST:PORT/\\]IMAGE\\[:TAG\\] \\[--docker-username USERNAME\\]", PushCommandName))
			Eventually(session).Should(Say("cf %s -f MANIFEST_WITH_MULTIPLE_APPS_PATH \\[APP_NAME\\] \\[--no-start\\]", PushCommandName))
			Eventually(session).Should(Say("OPTIONS:"))
			Eventually(session).Should(Say("ENVIRONMENT:"))
			Eventually(session).Should(Say("CF_STAGING_TIMEOUT=15        Max wait time for buildpack staging, in minutes"))
			Eventually(session).Should(Say("CF_STARTUP_TIMEOUT=5         Max wait time for app instance startup, in minutes"))
			Eventually(session).Should(Say("SEE ALSO:"))
			Eventually(session).Should(Say("apps, create-app-manifest, logs, ssh, start"))
			Eventually(session).Should(Exit(0))
		})
	})
})
