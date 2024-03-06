package experimental

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"time"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("v3-zdt-push command", func() {
	var (
		orgName           string
		spaceName         string
		appName           string
		userName          string
		PublicDockerImage = "cloudfoundry/diego-docker-app-custom"
	)

	BeforeEach(func() {
		orgName = helpers.NewOrgName()
		spaceName = helpers.NewSpaceName()
		appName = helpers.PrefixedRandomName("app")
		userName, _ = helpers.GetCredentials()
		helpers.TurnOffExperimental()
	})

	AfterEach(func() {
		helpers.TurnOnExperimental()
	})

	Describe("help", func() {
		Context("when --help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("v3-zdt-push", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("v3-zdt-push - Update an app with zero down time"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say(`cf v3-zdt-push APP_NAME \[-b BUILDPACK\]\.\.\. \[-p APP_PATH\] \[--no-route\]`))
				Eventually(session).Should(Say(`cf v3-zdt-push APP_NAME --docker-image \[REGISTRY_HOST:PORT/\]IMAGE\[:TAG\] \[--docker-username USERNAME\] \[--no-route\]`))
				Eventually(session).Should(Say("OPTIONS:"))
				Eventually(session).Should(Say(`-b\s+Custom buildpack by name \(e\.g\. my-buildpack\) or Git URL \(e\.g\. 'https://github.com/cloudfoundry/java-buildpack.git'\) or Git URL with a branch or tag \(e\.g\. 'https://github.com/cloudfoundry/java-buildpack\.git#v3.3.0' for 'v3.3.0' tag\)\. To use built-in buildpacks only, specify 'default' or 'null'`))
				Eventually(session).Should(Say(`-s\s+Stack to use \(a stack is a pre-built file system, including an operating system, that can run apps\)`))
				Eventually(session).Should(Say(`--docker-image, -o\s+Docker image to use \(e\.g\. user/docker-image-name\)`))
				Eventually(session).Should(Say(`--docker-username\s+Repository username; used with password from environment variable CF_DOCKER_PASSWORD`))
				Eventually(session).Should(Say(`--no-route\s+Do not map a route to this app`))
				Eventually(session).Should(Say(`-p\s+Path to app directory or to a zip file of the contents of the app directory`))
				Eventually(session).Should(Say("ENVIRONMENT:"))
				Eventually(session).Should(Say(`CF_DOCKER_PASSWORD=\s+Password used for private docker repository`))
				Eventually(session).Should(Say(`CF_STAGING_TIMEOUT=15\s+Max wait time for buildpack staging, in minutes`))
				Eventually(session).Should(Say(`CF_STARTUP_TIMEOUT=5\s+Max wait time for app instance startup, in minutes`))

				Eventually(session).Should(Exit(0))
			})
		})
	})

	Context("when the app name is not provided", func() {
		It("tells the user that the app name is required, prints help text, and exits 1", func() {
			session := helpers.CF("v3-zdt-push")

			Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `APP_NAME` was not provided"))
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Exit(1))
		})
	})

	It("displays the experimental warning", func() {
		session := helpers.CF("v3-zdt-push", appName)
		Eventually(session.Err).Should(Say("This command is in EXPERIMENTAL stage and may change without notice"))
		Eventually(session).Should(Exit())
	})

	Context("when the -b flag is not given an arg", func() {
		It("tells the user that the flag requires an arg, prints help text, and exits 1", func() {
			session := helpers.CF("v3-zdt-push", appName, "-b")

			Eventually(session.Err).Should(Say("Incorrect Usage: expected argument for flag `-b'"))
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Exit(1))
		})
	})

	Context("when the -p flag is not given an arg", func() {
		It("tells the user that the flag requires an arg, prints help text, and exits 1", func() {
			session := helpers.CF("v3-zdt-push", appName, "-p")

			Eventually(session.Err).Should(Say("Incorrect Usage: expected argument for flag `-p'"))
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Exit(1))
		})
	})

	Context("when the -p flag path does not exist", func() {
		It("tells the user that the flag requires an arg, prints help text, and exits 1", func() {
			session := helpers.CF("v3-zdt-push", appName, "-p", "path/that/does/not/exist")

			Eventually(session.Err).Should(Say("Incorrect Usage: The specified path 'path/that/does/not/exist' does not exist."))
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Exit(1))
		})
	})

	Context("when the environment is not setup correctly", func() {
		Context("when the v3 api version is lower than the minimum version", func() {
			var server *Server

			BeforeEach(func() {
				server = helpers.StartAndTargetMockServerWithAPIVersions(helpers.DefaultV2Version, "3.0.0")
			})

			AfterEach(func() {
				server.Close()
			})

			It("fails with error message that the minimum version is not met", func() {
				session := helpers.CF("v3-zdt-push", appName)
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say(`This command requires CF API version 3\.63\.0 or higher\.`))
				Eventually(session).Should(Exit(1))
			})
		})

		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "v3-zdt-push", appName)
		})
	})

	Context("when the environment is set up correctly", func() {
		var domainName string

		BeforeEach(func() {
			helpers.SetupCF(orgName, spaceName)
			domainName = helpers.DefaultSharedDomain()
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		Context("when the app exists", func() {
			var session *Session
			BeforeEach(func() {
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "v3-zdt-push", appName)).Should(Exit(0))
				})

				helpers.WithHelloWorldApp(func(appDir string) {
					session = helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "v3-zdt-push", appName, "-b", "https://github.com/cloudfoundry/staticfile-buildpack")
					Eventually(session).Should(Exit(0))
				})
			})

			It("pushes the app", func() {
				Eventually(session).Should(Say(`Updating app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Say(""))
				Eventually(session).Should(Say(`Uploading and creating bits package for app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Say(""))
				Eventually(session).Should(Say(`Staging package for app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Say(""))
				Eventually(session).Should(Say(`Mapping routes\.\.\.`))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Say(""))
				Eventually(session).Should(Say(`Starting deployment for app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Say(""))
				Eventually(session).Should(Say(`Waiting for app to start\.\.\.`))
				Eventually(session).Should(Say(`Showing health and status for app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
				Eventually(session).Should(Say(""))
				Eventually(session).Should(Say(`name:\s+%s`, appName))
				Eventually(session).Should(Say(`requested state:\s+started`))
				Eventually(session).Should(Say(`routes:\s+%s\.%s`, appName, domainName))
				Eventually(session).Should(Say(`stack:\s+cflinuxfs`))

				// TODO: Uncomment when capi sorts out droplet buildpack name/detectoutput
				// Eventually(session).Should(Say(`buildpacks:\s+https://github.com/cloudfoundry/staticfile-buildpack`))
				Eventually(session).Should(Say(""))
				Eventually(session).Should(Say(`type:\s+web`))
				Eventually(session).Should(Say(`instances:\s+1/1`))
				Eventually(session).Should(Say(`memory usage:\s+\d+(M|G)`))
				Eventually(session).Should(Say(`state\s+since\s+cpu\s+memory\s+disk`))
				Eventually(session).Should(Say(`#0\s+running\s+\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2} [AP]M`))
			})
		})

		Context("when the app does not already exist", func() {
			var session *Session

			BeforeEach(func() {
				helpers.WithHelloWorldApp(func(appDir string) {
					session = helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "v3-zdt-push", appName)
					Eventually(session).Should(Exit(0))
				})
			})

			It("pushes the app", func() {
				Eventually(session).Should(Say(`Creating app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Say(""))
				Eventually(session).Should(Say(`Uploading and creating bits package for app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Say(""))
				Consistently(session).ShouldNot(Say("Stopping app %s", appName))
				Eventually(session).Should(Say(`Staging package for app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Say(""))
				Eventually(session).Should(Say(`Mapping routes\.\.\.`))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Say(""))
				Eventually(session).Should(Say(`Setting app %s to droplet .+ in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Say(""))
				Eventually(session).Should(Say(`Starting app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Say(""))
				Eventually(session).Should(Say(`Waiting for app to start\.\.\.`))
				Eventually(session).Should(Say(`Showing health and status for app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
				Eventually(session).Should(Say(""))
				Eventually(session).Should(Say(`name:\s+%s`, appName))
				Eventually(session).Should(Say(`requested state:\s+started`))
				Eventually(session).Should(Say(`routes:\s+%s\.%s`, appName, domainName))
				Eventually(session).Should(Say(`stack:\s+cflinuxfs`))
				Eventually(session).Should(Say(`buildpacks:\s+staticfile`))
				Eventually(session).Should(Say(""))
				Eventually(session).Should(Say(`type:\s+web`))
				Eventually(session).Should(Say(`instances:\s+1/1`))
				Eventually(session).Should(Say(`memory usage:\s+\d+(M|G)`))
				Eventually(session).Should(Say(`state\s+since\s+cpu\s+memory\s+disk`))
				Eventually(session).Should(Say(`#0\s+running\s+\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2} [AP]M`))
			})
		})

		Context("when the app crashes", func() {
			var session *Session

			BeforeEach(func() {
				helpers.WithCrashingApp(func(appDir string) {
					session = helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "v3-zdt-push", appName)
					Eventually(session).Should(Exit(0))
				})
			})

			It("pushes the app", func() {
				Eventually(session).Should(Say(`Creating app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Say(""))
				Eventually(session).Should(Say(`Uploading and creating bits package for app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Say(""))
				Consistently(session).ShouldNot(Say("Stopping app %s", appName))
				Eventually(session).Should(Say(`Staging package for app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Say(""))
				Eventually(session).Should(Say(`Mapping routes\.\.\.`))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Say(""))
				Eventually(session).Should(Say(`Setting app %s to droplet .+ in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Say(""))
				Eventually(session).Should(Say(`Starting app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Say(""))
				Eventually(session).Should(Say(`Waiting for app to start\.\.\.`))
				Eventually(session).Should(Say(`Showing health and status for app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
				Eventually(session).Should(Say(""))
				Eventually(session).Should(Say(`name:\s+%s`, appName))
				Eventually(session).Should(Say(`requested state:\s+started`))
				Eventually(session).Should(Say(`routes:\s+%s\.%s`, appName, domainName))
				Eventually(session).Should(Say(`stack:\s+cflinuxfs`))
				Eventually(session).Should(Say(`buildpacks:\s+ruby`))
				Eventually(session).Should(Say(""))
				Eventually(session).Should(Say(`type:\s+web`))
				Eventually(session).Should(Say(`instances:\s+0/1`))
				Eventually(session).Should(Say(`memory usage:\s+\d+(M|G)`))
				Eventually(session).Should(Say(`state\s+since\s+cpu\s+memory\s+disk`))
				Eventually(session).Should(Say(`#0\s+crashed\s+\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2} [AP]M`))
			})
		})

		Context("when the -p flag is provided", func() {
			Context("when the path is a directory", func() {
				Context("when the directory contains files", func() {
					It("pushes the app from the directory", func() {
						helpers.WithHelloWorldApp(func(appDir string) {
							session := helpers.CF("v3-zdt-push", appName, "-p", appDir)

							Eventually(session).Should(Say(`Starting app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
							Eventually(session).Should(Say(`Waiting for app to start\.\.\.`))
							Eventually(session).Should(Say(`Showing health and status for app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
							Eventually(session).Should(Say(""))
							Eventually(session).Should(Say(`name:\s+%s`, appName))
							Eventually(session).Should(Say(`requested state:\s+started`))
							Eventually(session).Should(Say(`routes:\s+%s\.%s`, appName, domainName))
							Eventually(session).Should(Say(`stack:\s+cflinuxfs`))
							Eventually(session).Should(Say(`buildpacks:\s+staticfile`))
							Eventually(session).Should(Say(""))
							Eventually(session).Should(Say(`type:\s+web`))
							Eventually(session).Should(Say(`instances:\s+1/1`))
							Eventually(session).Should(Say(`memory usage:\s+\d+(M|G)`))
							Eventually(session).Should(Say(`state\s+since\s+cpu\s+memory\s+disk`))
							Eventually(session).Should(Say(`#0\s+running\s+\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2} [AP]M`))

							Eventually(session).Should(Exit(0))
						})
					})
				})

				Context("when the directory is empty", func() {
					var emptyDir string

					BeforeEach(func() {
						var err error
						emptyDir, err = ioutil.TempDir("", "integration-push-path-empty")
						Expect(err).ToNot(HaveOccurred())
					})

					AfterEach(func() {
						Expect(os.RemoveAll(emptyDir)).ToNot(HaveOccurred())
					})

					It("returns an error", func() {
						session := helpers.CF("v3-zdt-push", appName, "-p", emptyDir)
						Eventually(session.Err).Should(Say("No app files found in '%s'", regexp.QuoteMeta(emptyDir)))
						Eventually(session).Should(Exit(1))
					})
				})
			})

			Context("when the path is a zip file", func() {
				Context("pushing a zip file", func() {
					var archive string

					BeforeEach(func() {
						helpers.WithHelloWorldApp(func(appDir string) {
							tmpfile, err := ioutil.TempFile("", "push-archive-integration")
							Expect(err).ToNot(HaveOccurred())
							archive = tmpfile.Name()
							Expect(tmpfile.Close())

							err = helpers.Zipit(appDir, archive, "")
							Expect(err).ToNot(HaveOccurred())
						})
					})

					AfterEach(func() {
						Expect(os.RemoveAll(archive)).ToNot(HaveOccurred())
					})

					It("pushes the app from the zip file", func() {
						session := helpers.CF("v3-zdt-push", appName, "-p", archive)

						Eventually(session).Should(Say(`Starting app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
						Eventually(session).Should(Say(`Waiting for app to start\.\.\.`))
						Eventually(session).Should(Say(`Showing health and status for app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
						Eventually(session).Should(Say(""))
						Eventually(session).Should(Say(`name:\s+%s`, appName))
						Eventually(session).Should(Say(`requested state:\s+started`))
						Eventually(session).Should(Say(`routes:\s+%s\.%s`, appName, domainName))
						Eventually(session).Should(Say(`stack:\s+cflinuxfs`))
						Eventually(session).Should(Say(`buildpacks:\s+staticfile`))
						Eventually(session).Should(Say(""))
						Eventually(session).Should(Say(`type:\s+web`))
						Eventually(session).Should(Say(`instances:\s+1/1`))
						Eventually(session).Should(Say(`memory usage:\s+\d+(M|G)`))
						Eventually(session).Should(Say(`state\s+since\s+cpu\s+memory\s+disk`))
						Eventually(session).Should(Say(`#0\s+running\s+\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2} [AP]M`))

						Eventually(session).Should(Exit(0))
					})
				})
			})

			Context("when the path is a symlink to a directory", func() {
				var symlinkPath string

				BeforeEach(func() {
					tempFile, err := ioutil.TempFile("", "symlink-")
					Expect(err).ToNot(HaveOccurred())
					Expect(tempFile.Close()).To(Succeed())

					symlinkPath = tempFile.Name()
					Expect(os.Remove(symlinkPath)).To(Succeed())
				})

				AfterEach(func() {
					Expect(os.Remove(symlinkPath)).To(Succeed())
				})

				It("creates and uploads the package from the directory", func() {
					helpers.WithHelloWorldApp(func(appDir string) {
						Expect(os.Symlink(appDir, symlinkPath)).To(Succeed())

						session := helpers.CF("v3-zdt-push", appName, "-p", symlinkPath)

						Eventually(session).Should(Say(`Starting app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
						Eventually(session).Should(Say(`Waiting for app to start\.\.\.`))
						Eventually(session).Should(Say(`Showing health and status for app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
						Eventually(session).Should(Say(""))
						Eventually(session).Should(Say(`name:\s+%s`, appName))
						Eventually(session).Should(Say(`requested state:\s+started`))
						Eventually(session).Should(Say(`routes:\s+%s\.%s`, appName, domainName))
						Eventually(session).Should(Say(`stack:\s+cflinuxfs`))
						Eventually(session).Should(Say(`buildpacks:\s+staticfile`))
						Eventually(session).Should(Say(""))
						Eventually(session).Should(Say(`type:\s+web`))
						Eventually(session).Should(Say(`instances:\s+1/1`))
						Eventually(session).Should(Say(`memory usage:\s+\d+(M|G)`))
						Eventually(session).Should(Say(`state\s+since\s+cpu\s+memory\s+disk`))
						Eventually(session).Should(Say(`#0\s+running\s+\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2} [AP]M`))

						Eventually(session).Should(Exit(0))
					})
				})
			})

			Context("when the path is a symlink to a zip file", func() {
				var (
					archive     string
					symlinkPath string
				)

				BeforeEach(func() {
					helpers.WithHelloWorldApp(func(appDir string) {
						tmpfile, err := ioutil.TempFile("", "package-archive-integration")
						Expect(err).ToNot(HaveOccurred())
						archive = tmpfile.Name()
						Expect(tmpfile.Close())

						err = helpers.Zipit(appDir, archive, "")
						Expect(err).ToNot(HaveOccurred())
					})

					tempFile, err := ioutil.TempFile("", "symlink-to-archive-")
					Expect(err).ToNot(HaveOccurred())
					Expect(tempFile.Close()).To(Succeed())

					symlinkPath = tempFile.Name()
					Expect(os.Remove(symlinkPath)).To(Succeed())
					Expect(os.Symlink(archive, symlinkPath)).To(Succeed())
				})

				AfterEach(func() {
					Expect(os.Remove(archive)).To(Succeed())
					Expect(os.Remove(symlinkPath)).To(Succeed())
				})

				It("creates and uploads the package from the zip file", func() {
					session := helpers.CF("v3-zdt-push", appName, "-p", symlinkPath)

					Eventually(session).Should(Say(`Starting app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
					Eventually(session).Should(Say(`Waiting for app to start\.\.\.`))
					Eventually(session).Should(Say(`Showing health and status for app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
					Eventually(session).Should(Say(""))
					Eventually(session).Should(Say(`name:\s+%s`, appName))
					Eventually(session).Should(Say(`requested state:\s+started`))
					Eventually(session).Should(Say(`routes:\s+%s\.%s`, appName, domainName))
					Eventually(session).Should(Say(`stack:\s+cflinuxfs`))
					Eventually(session).Should(Say(`buildpacks:\s+staticfile`))
					Eventually(session).Should(Say(""))
					Eventually(session).Should(Say(`type:\s+web`))
					Eventually(session).Should(Say(`instances:\s+1/1`))
					Eventually(session).Should(Say(`memory usage:\s+\d+(M|G)`))
					Eventually(session).Should(Say(`state\s+since\s+cpu\s+memory\s+disk`))
					Eventually(session).Should(Say(`#0\s+running\s+\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2} [AP]M`))

					Eventually(session).Should(Exit(0))
				})
			})
		})

		Context("when the --no-route flag is set", func() {
			var session *Session

			BeforeEach(func() {
				helpers.WithHelloWorldApp(func(appDir string) {
					session = helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "v3-zdt-push", appName, "--no-route")
					Eventually(session).Should(Exit(0))
				})
			})

			It("does not map any routes to the app", func() {
				Consistently(session).ShouldNot(Say(`Mapping routes\.\.\.`))
				Eventually(session).Should(Say(`name:\s+%s`, appName))
				Eventually(session).Should(Say(`requested state:\s+started`))
				Eventually(session).Should(Say(`routes:\s+\n`))
				Eventually(session).Should(Say(`stack:\s+cflinuxfs`))
				Eventually(session).Should(Say(`buildpacks:\s+staticfile`))
				Eventually(session).Should(Say(""))
				Eventually(session).Should(Say(`type:\s+web`))
				Eventually(session).Should(Say(`instances:\s+1/1`))
				Eventually(session).Should(Say(`memory usage:\s+\d+(M|G)`))
				Eventually(session).Should(Say(`state\s+since\s+cpu\s+memory\s+disk`))
				Eventually(session).Should(Say(`#0\s+running\s+\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2} [AP]M`))
			})
		})

		Context("when the -s flag is set", func() {
			When("the default stack is specified", func() {
				It("uses the specified stack", func() {
					var session *Session
					helpers.WithHelloWorldApp(func(appDir string) {
						session = helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "v3-zdt-push", appName, "-s", "cflinuxfs3")
						Eventually(session).Should(Exit(0))
					})
					Eventually(session).Should(Say(`name:\s+%s`, appName))
					Eventually(session).Should(Say(`requested state:\s+started`))
					Eventually(session).Should(Say(`routes:\s+%s\.%s`, appName, domainName))
					Eventually(session).Should(Say(`stack:\s+cflinuxfs3`))
				})
			})

			When("a non-default stack is specified", func() {
				It("uses the specified stack", func() {

					Skip("we have only one stack")

				})
			})

			When("the both -s and -b are specified", func() {
				When("the buildpack and stack both exist", func() {
					When("the buildpack specified exists for the stack specified", func() {
						It("creates the app with the specified stack and buildpack", func() {
							var session *Session
							helpers.WithHelloWorldApp(func(appDir string) {
								session = helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "v3-zdt-push", appName, "-b", "https://github.com/cloudfoundry/staticfile-buildpack", "-s", "cflinuxfs3")
								Eventually(session).Should(Exit(0))
							})
							// TODO: assert specific error text when it is written
						})
					})

					When("the buildpack specified does not exist for the stack specified", func() {
						It("prints the appropriate error", func() {
							var session *Session
							helpers.WithHelloWorldApp(func(appDir string) {
								session = helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "v3-zdt-push", appName, "-b", "ruby_buildpack", "-s", "windows2012R2")
								Eventually(session).Should(Exit(1))
							})
							// TODO: assert specific error text when it is written
						})
					})
				})
			})

			When("the specified stack does not exist", func() {
				It("errors", func() {
					var session *Session
					helpers.WithHelloWorldApp(func(appDir string) {
						session = helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "v3-zdt-push", appName, "-s", "invalid_stack")
						Eventually(session).Should(Exit(1))
					})
					Eventually(session).Should(Say("FAILED"))
					// TODO: confirm the text of the error message
					Eventually(session.Err).Should(Say(`Stack must be an existing stack`))
				})
			})
		})

		Context("when the -b flag is set", func() {
			var session *Session

			Context("when pushing a multi-buildpack app", func() {
				BeforeEach(func() {
					helpers.WithMultiBuildpackApp(func(appDir string) {
						session = helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "v3-zdt-push", appName, "-b", "ruby_buildpack", "-b", "go_buildpack")

						// TODO: uncomment this expectation once capi-release displays all buildpacks on droplet
						// Story: https://www.pivotaltracker.com/story/show/150425459
						// Eventually(session).Should(Say("buildpacks:.*ruby_buildpack, go_buildpack"))

						Eventually(session, 10*time.Minute).Should(Exit(0))
					})
				})

				It("successfully compiles and runs the app", func() {
					resp, err := http.Get(fmt.Sprintf("http://%s.%s", appName, helpers.DefaultSharedDomain()))
					Expect(err).ToNot(HaveOccurred())
					Expect(resp.StatusCode).To(Equal(http.StatusOK))
				})
			})

			Context("when resetting the buildpack to default", func() {
				BeforeEach(func() {
					helpers.WithHelloWorldApp(func(appDir string) {
						Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "v3-zdt-push", appName, "-b", "java_buildpack")).Should(Exit(1))
						session = helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "v3-zdt-push", appName, "-b", "default")
						Eventually(session).Should(Exit(0))
					})
				})

				It("successfully pushes the app", func() {
					Eventually(session).Should(Say(`name:\s+%s`, appName))
					Eventually(session).Should(Say(`state\s+since\s+cpu\s+memory\s+disk`))
					Eventually(session).Should(Say(`#0\s+running\s+\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2} [AP]M`))
				})
			})

			Context("when omitting the buildpack", func() {
				BeforeEach(func() {
					helpers.WithHelloWorldApp(func(appDir string) {
						Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "v3-zdt-push", appName, "-b", "java_buildpack")).Should(Exit(1))
						session = helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "v3-zdt-push", appName)
						Eventually(session).Should(Exit(1))
					})
				})

				It("continues using previously set buildpack", func() {
					Eventually(session).Should(Say("FAILED"))
				})
			})

			Context("when the buildpack is invalid", func() {
				BeforeEach(func() {
					helpers.WithHelloWorldApp(func(appDir string) {
						session = helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "v3-zdt-push", appName, "-b", "wut")
						Eventually(session).Should(Exit(1))
					})
				})

				It("errors and does not push the app", func() {
					Consistently(session).ShouldNot(Say("Creating app"))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say(`Buildpack "wut" must be an existing admin buildpack or a valid git URI`))
				})
			})

			Context("when the buildpack is valid", func() {
				BeforeEach(func() {
					helpers.WithHelloWorldApp(func(appDir string) {
						session = helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "v3-zdt-push", appName, "-b", "https://github.com/cloudfoundry/staticfile-buildpack")
						Eventually(session).Should(Exit(0))
					})
				})

				It("uses the specified buildpack", func() {
					Eventually(session).Should(Say(`name:\s+%s`, appName))
					Eventually(session).Should(Say(`requested state:\s+started`))
					Eventually(session).Should(Say(`routes:\s+%s\.%s`, appName, domainName))
					Eventually(session).Should(Say(`stack:\s+cflinuxfs`))

					// TODO: Uncomment when capi sorts out droplet buildpack name/detectoutput
					// Eventually(session).Should(Say(`buildpacks:\s+https://github.com/cloudfoundry/staticfile-buildpack`))
					Eventually(session).Should(Say(""))
					Eventually(session).Should(Say(`type:\s+web`))
					Eventually(session).Should(Say(`instances:\s+1/1`))
					Eventually(session).Should(Say(`memory usage:\s+\d+(M|G)`))
					Eventually(session).Should(Say(`state\s+since\s+cpu\s+memory\s+disk`))
					Eventually(session).Should(Say(`#0\s+running\s+\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2} [AP]M`))
				})
			})
		})

		Context("when the -o flag is set", func() {
			Context("when the docker image is valid", func() {
				It("uses the specified docker image", func() {
					session := helpers.CF("v3-zdt-push", appName, "-o", PublicDockerImage)

					Eventually(session).Should(Say(`name:\s+%s`, appName))
					Eventually(session).Should(Say(`requested state:\s+started`))
					Eventually(session).Should(Say(`routes:\s+%s\.%s`, appName, domainName))
					Eventually(session).Should(Say("stack:"))
					Eventually(session).ShouldNot(Say("buildpacks:"))
					Eventually(session).Should(Say(`docker image:\s+%s`, PublicDockerImage))
					Eventually(session).Should(Say(""))
					Eventually(session).Should(Say(`type:\s+web`))
					Eventually(session).Should(Say(`instances:\s+1/1`))
					Eventually(session).Should(Say(`memory usage:\s+\d+(M|G)`))
					Eventually(session).Should(Say(`state\s+since\s+cpu\s+memory\s+disk`))
					Eventually(session).Should(Say(`#0\s+running\s+\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2} [AP]M`))
					Eventually(session).Should(Exit(0))
				})
			})

			Context("when the docker image is invalid", func() {
				It("displays an error and exits 1", func() {
					session := helpers.CF("v3-zdt-push", appName, "-o", "some-invalid-docker-image")
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("StagingError - Staging error: staging failed"))
					Eventually(session).Should(Exit(1))
				})
			})

			Context("when a docker username and password are provided with a private image", func() {
				var (
					privateDockerImage    string
					privateDockerUsername string
					privateDockerPassword string
				)

				BeforeEach(func() {
					privateDockerImage = os.Getenv("CF_INT_DOCKER_IMAGE")
					privateDockerUsername = os.Getenv("CF_INT_DOCKER_USERNAME")
					privateDockerPassword = os.Getenv("CF_INT_DOCKER_PASSWORD")

					if privateDockerImage == "" || privateDockerUsername == "" || privateDockerPassword == "" {
						Skip("CF_INT_DOCKER_IMAGE, CF_INT_DOCKER_USERNAME, or CF_INT_DOCKER_PASSWORD is not set")
					}
				})

				It("uses the specified private docker image", func() {
					session := helpers.CustomCF(
						helpers.CFEnv{
							EnvVars: map[string]string{"CF_DOCKER_PASSWORD": privateDockerPassword},
						},
						"v3-zdt-push", "--docker-username", privateDockerUsername, "--docker-image", privateDockerImage, appName,
					)

					Eventually(session).Should(Say(`name:\s+%s`, appName))
					Eventually(session).Should(Say(`requested state:\s+started`))
					Eventually(session).Should(Say(`routes:\s+%s\.%s`, appName, domainName))
					Eventually(session).Should(Say("stack:"))
					Eventually(session).ShouldNot(Say("buildpacks:"))
					Eventually(session).Should(Say(`docker image:\s+%s`, privateDockerImage))
					Eventually(session).Should(Say(""))
					Eventually(session).Should(Say(`type:\s+web`))
					Eventually(session).Should(Say(`instances:\s+1/1`))
					Eventually(session).Should(Say(`memory usage:\s+\d+(M|G)`))
					Eventually(session).Should(Say(`state\s+since\s+cpu\s+memory\s+disk`))
					Eventually(session).Should(Say(`#0\s+running\s+\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2} [AP]M`))
					Eventually(session).Should(Exit(0))
				})
			})
		})

		Describe("argument combination errors", func() {
			Context("when the --docker-username is provided without the -o flag", func() {
				It("displays an error and exits 1", func() {
					helpers.WithHelloWorldApp(func(appDir string) {
						session := helpers.CF("v3-zdt-push", appName, "--docker-username", "some-username")
						Eventually(session).Should(Say("FAILED"))
						Eventually(session.Err).Should(Say("Incorrect Usage: '--docker-image, -o' and '--docker-username' must be used together."))
						Eventually(session).Should(Say("NAME:"))
						Eventually(session).Should(Exit(1))
					})
				})
			})

			Context("when the --docker-username and -p flags are provided together", func() {
				It("displays an error and exits 1", func() {
					helpers.WithHelloWorldApp(func(appDir string) {
						session := helpers.CF("v3-zdt-push", appName, "--docker-username", "some-username", "-p", appDir)
						Eventually(session).Should(Say("FAILED"))
						Eventually(session.Err).Should(Say("Incorrect Usage: '--docker-image, -o' and '--docker-username' must be used together."))
						Eventually(session).Should(Say("NAME:"))
						Eventually(session).Should(Exit(1))
					})
				})
			})

			Context("when the --docker-username is provided without a password", func() {
				var oldPassword string

				BeforeEach(func() {
					oldPassword = os.Getenv("CF_DOCKER_PASSWORD")
					err := os.Unsetenv("CF_DOCKER_PASSWORD")
					Expect(err).ToNot(HaveOccurred())
				})

				AfterEach(func() {
					err := os.Setenv("CF_DOCKER_PASSWORD", oldPassword)
					Expect(err).ToNot(HaveOccurred())
				})

				It("displays an error and exits 1", func() {
					helpers.WithHelloWorldApp(func(appDir string) {
						session := helpers.CF("v3-zdt-push", appName, "--docker-username", "some-username", "--docker-image", "some-image")
						Eventually(session).Should(Say("FAILED"))
						Eventually(session.Err).Should(Say(`Environment variable CF_DOCKER_PASSWORD not set\.`))
						Eventually(session).Should(Exit(1))
					})
				})
			})

			Context("when the -o and -p flags are provided together", func() {
				It("displays an error and exits 1", func() {
					helpers.WithHelloWorldApp(func(appDir string) {
						session := helpers.CF("v3-zdt-push", appName, "-o", PublicDockerImage, "-p", appDir)
						Eventually(session).Should(Say("FAILED"))
						Eventually(session.Err).Should(Say("Incorrect Usage: The following arguments cannot be used together: --docker-image, -o, -p"))
						Eventually(session).Should(Say("NAME:"))
						Eventually(session).Should(Exit(1))
					})
				})
			})

			Context("when the -o and -b flags are provided together", func() {
				It("displays an error and exits 1", func() {
					helpers.WithHelloWorldApp(func(appDir string) {
						session := helpers.CF("v3-zdt-push", appName, "-o", PublicDockerImage, "-b", "some-buildpack")
						Eventually(session).Should(Say("FAILED"))
						Eventually(session.Err).Should(Say("Incorrect Usage: The following arguments cannot be used together: -b, --docker-image, -o"))
						Eventually(session).Should(Say("NAME:"))
						Eventually(session).Should(Exit(1))
					})
				})
			})
		})
	})
})
