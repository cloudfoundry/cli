package push

import (
	"fmt"
	"io/ioutil"
	"os"

	"code.cloudfoundry.org/cli/integration/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("when a droplet is provided", func() {
	var (
		appName     string
		dropletPath string
	)

	BeforeEach(func() {
		appName = helpers.NewAppName()

		helpers.WithHelloWorldApp(func(appDir string) {
			tmpfile, err := ioutil.TempFile("", "dropletFile")
			Expect(err).ToNot(HaveOccurred())
			dropletPath = tmpfile.Name()
			Expect(tmpfile.Close()).ToNot(HaveOccurred())

			tempApp := helpers.NewAppName()
			session := helpers.CF(PushCommandName, tempApp, "-b", "staticfile_buildpack")
			Eventually(session).Should(Exit(0))

			appGUID := helpers.AppGUID(tempApp)
			Eventually(helpers.CF("curl", fmt.Sprintf("/v2/apps/%s/droplet/download", appGUID), "--output", dropletPath)).Should(Exit(0))
			_, err = os.Stat(dropletPath)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	AfterEach(func() {
		Expect(os.RemoveAll(dropletPath)).ToNot(HaveOccurred())
	})

	Context("when the v2 api version is lower than the minimum version", func() {
		var server *Server

		BeforeEach(func() {
			server = helpers.StartAndTargetServerWithAPIVersions("2.50.0", helpers.DefaultV3Version)
		})

		AfterEach(func() {
			server.Close()
		})

		It("fails with error message that the minimum version is not met", func() {
			session := helpers.CF(PushCommandName, appName, "--droplet", dropletPath)
			Eventually(session).Should(Say("FAILED"))
			Eventually(session.Err).Should(Say("Option '--droplet' requires CF API version 2\\.63\\.0 or higher\\. Your target is 2\\.50\\.0"))
			Eventually(session).Should(Exit(1))
		})
	})

	Context("when the app does not exist", func() {

		It("creates the app", func() {
			session := helpers.CF(PushCommandName, appName, "--droplet", dropletPath)
			Eventually(session).Should(Say("Getting app info\\.\\.\\."))
			Eventually(session).Should(Say("Creating app with these attributes\\.\\.\\."))
			Eventually(session).Should(Say("\\+\\s+name:\\s+%s", appName))
			Eventually(session).Should(Say("Uploading droplet\\.\\.\\."))
			Eventually(session).Should(Say("Waiting for app to start\\.\\.\\."))
			Eventually(session).Should(Say("requested state:\\s+started"))
			Eventually(session).Should(Exit(0))

			session = helpers.CF("app", appName)
			Eventually(session).Should(Say("name:\\s+%s", appName))
			Eventually(session).Should(Exit(0))
		})
	})

	Context("when the app exists", func() {
		BeforeEach(func() {
			helpers.WithHelloWorldApp(func(appDir string) {
				Eventually(helpers.CF(PushCommandName, appName, "-p", appDir, "-b", "staticfile_buildpack")).Should(Exit(0))
			})
		})

		It("updates the app", func() {
			session := helpers.CF(PushCommandName, appName, "--droplet", dropletPath)
			Eventually(session).Should(Say("Updating app with these attributes\\.\\.\\."))
			Eventually(session).Should(Say("Uploading droplet\\.\\.\\."))
			Eventually(session).Should(Exit(0))
		})
	})
})
