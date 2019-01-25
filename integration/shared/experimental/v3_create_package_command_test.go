package experimental

import (
	"io/ioutil"
	"os"
	"regexp"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("v3-create-package command", func() {
	var (
		orgName   string
		spaceName string
		appName   string
	)

	BeforeEach(func() {
		orgName = helpers.NewOrgName()
		spaceName = helpers.NewSpaceName()
		appName = helpers.PrefixedRandomName("app")
	})

	Describe("help", func() {
		When("--help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("v3-create-package", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("v3-create-package - Uploads a V3 Package"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say(`cf v3-create-package APP_NAME \[-p APP_PATH \| --docker-image \[REGISTRY_HOST:PORT/\]IMAGE\[:TAG\]\]`))
				Eventually(session).Should(Say("OPTIONS:"))
				Eventually(session).Should(Say(`--docker-image, -o\s+Docker image to use \(e\.g\. user/docker-image-name\)`))
				Eventually(session).Should(Say(`-p\s+Path to app directory or to a zip file of the contents of the app directory`))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the app name is not provided", func() {
		It("tells the user that the app name is required, prints help text, and exits 1", func() {
			session := helpers.CF("v3-create-package")

			Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `APP_NAME` was not provided"))
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Exit(1))
		})
	})

	It("displays the experimental warning", func() {
		session := helpers.CF("v3-create-package", appName)
		Eventually(session.Err).Should(Say("This command is in EXPERIMENTAL stage and may change without notice"))
		Eventually(session).Should(Exit())
	})

	When("the -p flag is not given an arg", func() {
		It("tells the user that the flag requires an arg, prints help text, and exits 1", func() {
			session := helpers.CF("v3-create-package", appName, "-p")

			Eventually(session.Err).Should(Say("Incorrect Usage: expected argument for flag `-p'"))
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Exit(1))
		})
	})

	When("the -p flag path does not exist", func() {
		It("tells the user that the flag requires an arg, prints help text, and exits 1", func() {
			session := helpers.CF("v3-create-package", appName, "-p", "path/that/does/not/exist")

			Eventually(session.Err).Should(Say("Incorrect Usage: The specified path 'path/that/does/not/exist' does not exist."))
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Exit(1))
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "v3-create-package", appName)
		})
	})

	When("the environment is set up correctly", func() {
		BeforeEach(func() {
			helpers.SetupCF(orgName, spaceName)
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		When("the app does not exist", func() {
			It("returns a not found error", func() {
				session := helpers.CF("v3-create-package", appName)
				userName, _ := helpers.GetCredentials()
				Eventually(session).Should(Say("Uploading and creating bits package for app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName))
				Eventually(session.Err).Should(Say("App %s not found", appName))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the app exists", func() {
			BeforeEach(func() {
				Eventually(helpers.CF("v3-create-app", appName)).Should(Exit(0))
			})

			It("creates the package", func() {
				session := helpers.CF("v3-create-package", appName)
				userName, _ := helpers.GetCredentials()
				Eventually(session).Should(Say("Uploading and creating bits package for app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName))
				Eventually(session).Should(Say("package guid: %s", helpers.GUIDRegex))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))
			})

			When("the --docker-image flag is provided", func() {
				When("the docker-image exists", func() {
					It("creates the package", func() {
						session := helpers.CF("v3-create-package", appName, "--docker-image", PublicDockerImage)
						userName, _ := helpers.GetCredentials()
						Eventually(session).Should(Say("Creating docker package for app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName))
						Eventually(session).Should(Say("package guid: %s", helpers.GUIDRegex))
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Exit(0))
					})
				})
			})

			When("the -p flag is provided", func() {
				When("the path is a directory", func() {
					When("the directory contains files", func() {
						It("creates and uploads the package from the directory", func() {
							helpers.WithHelloWorldApp(func(appDir string) {
								session := helpers.CF("v3-create-package", appName, "-p", appDir)
								userName, _ := helpers.GetCredentials()

								Eventually(session).Should(Say("Uploading and creating bits package for app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName))
								Eventually(session).Should(Say("package guid: %s", helpers.GUIDRegex))
								Eventually(session).Should(Say("OK"))
								Eventually(session).Should(Exit(0))
							})
						})
					})

					When("the directory is empty", func() {
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
							session := helpers.CF("v3-create-package", appName, "-p", emptyDir)
							// TODO: Modify this after changing code if necessary
							Eventually(session.Err).Should(Say("No app files found in '%s'", regexp.QuoteMeta(emptyDir)))
							Eventually(session).Should(Exit(1))
						})
					})
				})

				When("the path is a zip file", func() {
					Context("pushing a zip file", func() {
						var archive string

						BeforeEach(func() {
							helpers.WithHelloWorldApp(func(appDir string) {
								tmpfile, err := ioutil.TempFile("", "package-archive-integration")
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

						It("creates and uploads the package from the zip file", func() {
							session := helpers.CF("v3-create-package", appName, "-p", archive)

							userName, _ := helpers.GetCredentials()

							Eventually(session).Should(Say("Uploading and creating bits package for app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName))
							Eventually(session).Should(Say("package guid: %s", helpers.GUIDRegex))
							Eventually(session).Should(Say("OK"))
							Eventually(session).Should(Exit(0))
						})
					})
				})

				When("the path is a symlink to a directory", func() {
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

							session := helpers.CF("v3-create-package", appName, "-p", symlinkPath)
							userName, _ := helpers.GetCredentials()

							Eventually(session).Should(Say("Uploading and creating bits package for app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName))
							Eventually(session).Should(Say("package guid: %s", helpers.GUIDRegex))
							Eventually(session).Should(Say("OK"))
							Eventually(session).Should(Exit(0))
						})
					})
				})
			})

			When("the path is a symlink to a zip file", func() {
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
					session := helpers.CF("v3-create-package", appName, "-p", symlinkPath)

					userName, _ := helpers.GetCredentials()

					Eventually(session).Should(Say("Uploading and creating bits package for app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName))
					Eventually(session).Should(Say("package guid: %s", helpers.GUIDRegex))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))
				})
			})

			When("the -o and -p flags are provided together", func() {
				It("displays an error and exits 1", func() {
					helpers.WithHelloWorldApp(func(appDir string) {
						session := helpers.CF("v3-create-package", appName, "-o", PublicDockerImage, "-p", appDir)
						Eventually(session).Should(Say("FAILED"))
						Eventually(session.Err).Should(Say("Incorrect Usage: The following arguments cannot be used together: --docker-image, -o, -p"))
						Eventually(session).Should(Say("NAME:"))
						Eventually(session).Should(Exit(1))
					})
				})
			})
		})
	})
})
