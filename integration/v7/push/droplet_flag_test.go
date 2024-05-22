package push

import (
	"fmt"
	"io/ioutil"
	"os"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("--droplet flag", func() {
	When("the --droplet flag is provided", func() {
		var (
			appName     string
			dropletPath string
			originalApp string
		)

		BeforeEach(func() {
			appName = helpers.NewAppName()

			helpers.WithHelloWorldApp(func(appDir string) {
				tmpfile, err := ioutil.TempFile("", "dropletFile*.tgz")
				Expect(err).ToNot(HaveOccurred())
				dropletPath = tmpfile.Name()
				Expect(tmpfile.Close()).ToNot(HaveOccurred())

				originalApp = helpers.NewAppName()
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, PushCommandName, originalApp, "-b", "staticfile_buildpack", "-v")).Should(Exit(0))
				})

				appGUID := helpers.AppGUID(originalApp)
				Eventually(helpers.CF("curl", fmt.Sprintf("/v2/apps/%s/droplet/download", appGUID), "--output", dropletPath)).Should(Exit(0))
				_, err = os.Stat(dropletPath)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		AfterEach(func() {
			Expect(os.RemoveAll(dropletPath)).ToNot(HaveOccurred())
		})

		When("the app does not exist", func() {
			It("creates the app with the given droplet", func() {
				session := helpers.CF(PushCommandName, appName, "--droplet", dropletPath)
				Eventually(session).Should(Say(`Pushing app %s`, appName))
				Eventually(session).Should(Say(`Uploading droplet bits\.\.\.`))
				Eventually(session).Should(Say(`Waiting for app %s to start\.\.\.`, appName))
				Eventually(session).Should(Say(`requested state:\s+started`))
				Eventually(session).Should(Exit(0))

				session = helpers.CF("app", appName)
				Eventually(session).Should(Say(`name:\s+%s`, appName))
				Eventually(session).Should(Exit(0))
			})
		})

		When("the app already exists", func() {
			It("updates the app with the given droplet", func() {
				session := helpers.CF(PushCommandName, originalApp, "--droplet", dropletPath)
				Eventually(session).Should(Say(`Pushing app %s`, originalApp))
				Eventually(session).Should(Say(`Uploading droplet bits\.\.\.`))
				Eventually(session).Should(Say(`Waiting for app %s to start\.\.\.`, originalApp))
				Eventually(session).Should(Say(`requested state:\s+started`))
				Eventually(session).Should(Exit(0))

				session = helpers.CF("app", originalApp)
				Eventually(session).Should(Say(`name:\s+%s`, originalApp))
				Eventually(session).Should(Exit(0))
			})
		})

		When("the droplet bits path is not a gzipped tarball", func() {
			It("fails with a helpful error message", func() {
				nonTgzFile, err := ioutil.TempFile("", "dropletFile*.txt")
				Expect(err).ToNot(HaveOccurred())
				session := helpers.CF(PushCommandName, appName, "--droplet", nonTgzFile.Name())
				Eventually(session).Should(Say(`FAILED`))
				Eventually(session.Err).Should(Say(`Uploaded droplet file is invalid: .+ not a tgz`))
				Eventually(session).Should(Exit(1))
			})
		})

		When("along with the --no-start flag", func() {
			It("updates the app with the given droplet", func() {
				originalAppGUID := helpers.AppGUID(originalApp)

				var routeResponse struct {
					GUID string `json:"guid"`
				}

				currentDropletEndpoint := fmt.Sprintf("v3/apps/%s/droplets/current", originalAppGUID)

				helpers.Curl(&routeResponse, currentDropletEndpoint)
				preUploadDropletGUID := routeResponse.GUID

				session := helpers.CF(PushCommandName, originalApp, "--droplet", dropletPath, "--no-start")
				Eventually(session).Should(Say(`Pushing app %s`, originalApp))
				Eventually(session).Should(Say(`Uploading droplet bits\.\.\.`))
				Eventually(session).Should(Say(`requested state:\s+stopped`))
				Eventually(session).Should(Exit(0))

				helpers.Curl(&routeResponse, currentDropletEndpoint)
				postUploadDropletGUID := routeResponse.GUID

				Expect(preUploadDropletGUID).To(Not(Equal(postUploadDropletGUID)))
			})
		})
	})
})
