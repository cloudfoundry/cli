package push

import (
	"fmt"
	"regexp"
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/integration/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("push with various flags and no manifest", func() {
	var (
		appName string
	)

	BeforeEach(func() {
		appName = helpers.NewAppName()
	})

	When("the API version is below 3.27.0 and health check type is not http", func() {
		BeforeEach(func() {
			helpers.SkipIfVersionAtLeast(ccversion.MinVersionV3)
		})

		It("creates the app with the specified settings, with the health check type", func() {
			helpers.WithHelloWorldApp(func(dir string) {
				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir},
					PushCommandName, appName,
					"-b", "staticfile_buildpack",
					"-c", fmt.Sprintf("echo 'hi' && %s", helpers.StaticfileBuildpackStartCommand),
					"-u", "port", //works if this stuff is commentted out
					"-k", "300M",
					"-i", "2",
					"-m", "70M",
					"-s", "cflinuxfs2",
					"-t", "180",
				)
				Eventually(session).Should(Say("Getting app info\\.\\.\\."))
				Eventually(session).Should(Say("Creating app with these attributes\\.\\.\\."))
				Eventually(session).Should(Say("\\+\\s+name:\\s+%s", appName))
				Eventually(session).Should(Say("(?m)\\s+buildpacks:\\s+\\+\\s+staticfile_buildpack"))
				Eventually(session).Should(Say("\\s+command:\\s+echo 'hi' && %s", regexp.QuoteMeta(helpers.StaticfileBuildpackStartCommand)))
				Eventually(session).Should(Say("\\s+disk quota:\\s+300M"))
				Eventually(session).Should(Say("\\s+health check timeout:\\s+180"))
				Eventually(session).Should(Say("\\s+health check type:\\s+port"))
				Eventually(session).Should(Say("\\s+instances:\\s+2"))
				Eventually(session).Should(Say("\\s+memory:\\s+70M"))
				Eventually(session).Should(Say("\\s+stack:\\s+cflinuxfs2"))
				Eventually(session).Should(Say("\\s+routes:"))
				Eventually(session).Should(Say("(?i)\\+\\s+%s.%s", appName, helpers.DefaultSharedDomain()))
				Eventually(session).Should(Say("Mapping routes\\.\\.\\."))
				Eventually(session).Should(Say("Uploading files\\.\\.\\."))
				Eventually(session).Should(Say("100.00%"))
				Eventually(session).Should(Say("Waiting for API to complete processing files\\.\\.\\."))
				helpers.ConfirmStagingLogs(session)
				Eventually(session).Should(Say("Waiting for app to start\\.\\.\\."))
				Eventually(session).Should(Say("requested state:\\s+started"))
				Eventually(session).Should(Say("start command:\\s+echo 'hi' && %s", regexp.QuoteMeta(helpers.StaticfileBuildpackStartCommand)))
				Eventually(session).Should(Exit(0))
			})

			// output is different from when API version is above 3.27.0
			session := helpers.CF("app", appName)
			Eventually(session).Should(Say("name:\\s+%s", appName))
			Eventually(session).Should(Say("instances:\\s+\\d/2"))
			Eventually(session).Should(Say("usage:\\s+70M x 2"))
			Eventually(session).Should(Say("stack:\\s+cflinuxfs2"))
			Eventually(session).Should(Say("buildpack:\\s+staticfile_buildpack"))
			Eventually(session).Should(Say("#0.* of 70M"))
			Eventually(session).Should(Exit(0))
		})
	})

	When("the API version is above 3.27.0 and health check type is not http", func() {
		BeforeEach(func() {
			helpers.SkipIfVersionLessThan(ccversion.MinVersionV3)
		})

		It("creates the app with the specified settings, with the health check type", func() {
			helpers.WithHelloWorldApp(func(dir string) {
				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir},
					PushCommandName, appName,
					"-b", "staticfile_buildpack",
					"-c", fmt.Sprintf("echo 'hi' && %s", helpers.StaticfileBuildpackStartCommand),
					"-u", "port", //works if this stuff is commentted out
					"-k", "300M",
					"-i", "2",
					"-m", "70M",
					"-s", "cflinuxfs2",
					"-t", "180",
				)
				Eventually(session).Should(Say("Getting app info\\.\\.\\."))
				Eventually(session).Should(Say("Creating app with these attributes\\.\\.\\."))
				Eventually(session).Should(Say("\\+\\s+name:\\s+%s", appName))
				Eventually(session).Should(Say("(?m)\\s+buildpacks:\\s+\\+\\s+staticfile_buildpack"))
				Eventually(session).Should(Say("\\s+command:\\s+echo 'hi' && %s", regexp.QuoteMeta(helpers.StaticfileBuildpackStartCommand)))
				Eventually(session).Should(Say("\\s+disk quota:\\s+300M"))
				Eventually(session).Should(Say("\\s+health check timeout:\\s+180"))
				Eventually(session).Should(Say("\\s+health check type:\\s+port"))
				Eventually(session).Should(Say("\\s+instances:\\s+2"))
				Eventually(session).Should(Say("\\s+memory:\\s+70M"))
				Eventually(session).Should(Say("\\s+stack:\\s+cflinuxfs2"))
				Eventually(session).Should(Say("\\s+routes:"))
				Eventually(session).Should(Say("(?i)\\+\\s+%s.%s", appName, helpers.DefaultSharedDomain()))
				Eventually(session).Should(Say("Mapping routes\\.\\.\\."))
				Eventually(session).Should(Say("Uploading files\\.\\.\\."))
				Eventually(session).Should(Say("100.00%"))
				Eventually(session).Should(Say("Waiting for API to complete processing files\\.\\.\\."))
				helpers.ConfirmStagingLogs(session)
				Eventually(session).Should(Say("Waiting for app to start\\.\\.\\."))
				Eventually(session).Should(Say("requested state:\\s+started"))
				Eventually(session).Should(Say("start command:\\s+echo 'hi' && %s", regexp.QuoteMeta(helpers.StaticfileBuildpackStartCommand)))
				Eventually(session).Should(Exit(0))
			})

			// output is different from when API version is below 3.27.0
			time.Sleep(5 * time.Second)
			session := helpers.CF("app", appName)
			Eventually(session).Should(Say("name:\\s+%s", appName))
			Eventually(session).Should(Say("last uploaded:\\s+\\w{3} \\d{1,2} \\w{3} \\d{2}:\\d{2}:\\d{2} \\w{3} \\d{4}"))
			Eventually(session).Should(Say("stack:\\s+cflinuxfs2"))
			Eventually(session).Should(Say("buildpacks:\\s+staticfile"))
			Eventually(session).Should(Say("type:\\s+web"))
			Eventually(session).Should(Say("instances:\\s+2/2"))
			Eventually(session).Should(Say("memory usage:\\s+70M"))
			Eventually(session).Should(Say("\\s+state\\s+since\\s+cpu\\s+memory\\s+disk"))
			Eventually(session).Should(Say("#0\\s+running\\s+\\d{4}-[01]\\d-[0-3]\\dT[0-2][0-9]:[0-5]\\d:[0-5]\\dZ"))
			Eventually(session).Should(Exit(0))
		})
	})

	When("the API version is below 3.27.0 and health check type is http", func() {
		BeforeEach(func() {
			helpers.SkipIfVersionAtLeast(ccversion.MinVersionV3)
		})

		It("creates the app with the specified settings, with the health check type being 'http'", func() {
			helpers.WithHelloWorldApp(func(dir string) {
				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir},
					PushCommandName, appName,
					"-b", "staticfile_buildpack",
					"-c", fmt.Sprintf("echo 'hi' && %s", helpers.StaticfileBuildpackStartCommand),
					"-u", "http",
					"-k", "300M",
					"-i", "2",
					"-m", "70M",
					"-s", "cflinuxfs2",
					"-t", "180",
				)

				Eventually(session).Should(Say("Getting app info\\.\\.\\."))
				Eventually(session).Should(Say("Creating app with these attributes\\.\\.\\."))
				Eventually(session).Should(Say("\\+\\s+name:\\s+%s", appName))
				Eventually(session).Should(Say("(?m)\\s+buildpacks:\\s+\\+\\s+staticfile_buildpack"))
				Eventually(session).Should(Say("\\s+command:\\s+echo 'hi' && %s", regexp.QuoteMeta(helpers.StaticfileBuildpackStartCommand)))
				Eventually(session).Should(Say("\\s+disk quota:\\s+300M"))
				Eventually(session).Should(Say("\\s+health check http endpoint:\\s+/"))
				Eventually(session).Should(Say("\\s+health check timeout:\\s+180"))
				Eventually(session).Should(Say("\\s+health check type:\\s+http"))
				Eventually(session).Should(Say("\\s+instances:\\s+2"))
				Eventually(session).Should(Say("\\s+memory:\\s+70M"))
				Eventually(session).Should(Say("\\s+stack:\\s+cflinuxfs2"))
				Eventually(session).Should(Say("\\s+routes:"))
				Eventually(session).Should(Say("(?i)\\+\\s+%s.%s", appName, helpers.DefaultSharedDomain()))
				Eventually(session).Should(Say("Mapping routes\\.\\.\\."))
				Eventually(session).Should(Say("Uploading files\\.\\.\\."))
				Eventually(session).Should(Say("100.00%"))
				Eventually(session).Should(Say("Waiting for API to complete processing files\\.\\.\\."))
				helpers.ConfirmStagingLogs(session)
				Eventually(session).Should(Say("Waiting for app to start\\.\\.\\."))
				Eventually(session).Should(Say("requested state:\\s+started"))
				Eventually(session).Should(Say("start command:\\s+echo 'hi' && %s", regexp.QuoteMeta(helpers.StaticfileBuildpackStartCommand)))
				Eventually(session).Should(Exit(0))
			})

			// output is different from when API version is above 3.27.0
			session := helpers.CF("app", appName)
			Eventually(session).Should(Say("name:\\s+%s", appName))
			Eventually(session).Should(Say("instances:\\s+\\d/2"))
			Eventually(session).Should(Say("usage:\\s+70M x 2"))
			Eventually(session).Should(Say("stack:\\s+cflinuxfs2"))
			Eventually(session).Should(Say("buildpack:\\s+staticfile_buildpack"))
			Eventually(session).Should(Say("#0.* of 70M"))
			Eventually(session).Should(Exit(0))
		})
	})

	When("the API version is above 3.27.0 and health check type is http", func() {
		BeforeEach(func() {
			helpers.SkipIfVersionLessThan(ccversion.MinVersionV3)
		})

		It("creates the app with the specified settings, with the health check type being 'http'", func() {
			helpers.WithHelloWorldApp(func(dir string) {
				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir},
					PushCommandName, appName,
					"-b", "staticfile_buildpack",
					"-c", fmt.Sprintf("echo 'hi' && %s", helpers.StaticfileBuildpackStartCommand),
					"-u", "http",
					"-k", "300M",
					"-i", "2",
					"-m", "70M",
					"-s", "cflinuxfs2",
					"-t", "180",
				)

				Eventually(session).Should(Say("Getting app info\\.\\.\\."))
				Eventually(session).Should(Say("Creating app with these attributes\\.\\.\\."))
				Eventually(session).Should(Say("\\+\\s+name:\\s+%s", appName))
				Eventually(session).Should(Say("(?m)\\s+buildpacks:\\s+\\+\\s+staticfile_buildpack"))
				Eventually(session).Should(Say("\\s+command:\\s+echo 'hi' && %s", regexp.QuoteMeta(helpers.StaticfileBuildpackStartCommand)))
				Eventually(session).Should(Say("\\s+disk quota:\\s+300M"))
				Eventually(session).Should(Say("\\s+health check http endpoint:\\s+/"))
				Eventually(session).Should(Say("\\s+health check timeout:\\s+180"))
				Eventually(session).Should(Say("\\s+health check type:\\s+http"))
				Eventually(session).Should(Say("\\s+instances:\\s+2"))
				Eventually(session).Should(Say("\\s+memory:\\s+70M"))
				Eventually(session).Should(Say("\\s+stack:\\s+cflinuxfs2"))
				Eventually(session).Should(Say("\\s+routes:"))
				Eventually(session).Should(Say("(?i)\\+\\s+%s.%s", appName, helpers.DefaultSharedDomain()))
				Eventually(session).Should(Say("Mapping routes\\.\\.\\."))
				Eventually(session).Should(Say("Uploading files\\.\\.\\."))
				Eventually(session).Should(Say("100.00%"))
				Eventually(session).Should(Say("Waiting for API to complete processing files\\.\\.\\."))
				helpers.ConfirmStagingLogs(session)
				Eventually(session).Should(Say("Waiting for app to start\\.\\.\\."))
				Eventually(session).Should(Say("requested state:\\s+started"))
				Eventually(session).Should(Say("start command:\\s+echo 'hi' && %s", regexp.QuoteMeta(helpers.StaticfileBuildpackStartCommand)))
				Eventually(session).Should(Exit(0))
			})

			// output is different from when API version is below 3.27.0

			time.Sleep(5 * time.Second)
			session := helpers.CF("app", appName)
			Eventually(session).Should(Say("name:\\s+%s", appName))
			Eventually(session).Should(Say("last uploaded:\\s+\\w{3} \\d{1,2} \\w{3} \\d{2}:\\d{2}:\\d{2} \\w{3} \\d{4}"))
			Eventually(session).Should(Say("stack:\\s+cflinuxfs2"))
			Eventually(session).Should(Say("buildpacks:\\s+staticfile"))
			Eventually(session).Should(Say("type:\\s+web"))
			Eventually(session).Should(Say("instances:\\s+2/2"))
			Eventually(session).Should(Say("memory usage:\\s+70M"))
			Eventually(session).Should(Say("\\s+state\\s+since\\s+cpu\\s+memory\\s+disk"))
			Eventually(session).Should(Say("#0\\s+running\\s+\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}Z"))
			Eventually(session).Should(Exit(0))
		})
	})
})
