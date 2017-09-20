package isolated

import (
	"fmt"
	"io/ioutil"
	"os"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Push with manifest", func() {
	var (
		appName           string
		orgName           string
		tempFile          string
		oldDockerPassword string
	)

	BeforeEach(func() {
		orgName = helpers.NewOrgName()
		spaceName := helpers.NewSpaceName()
		setupCF(orgName, spaceName)

		appName = helpers.PrefixedRandomName("app")
		f, err := ioutil.TempFile("", "combination-manifest-with-p")
		Expect(err).ToNot(HaveOccurred())
		Expect(f.Close()).To(Succeed())
		tempFile = f.Name()

		oldDockerPassword = os.Getenv("CF_DOCKER_PASSWORD")
		Expect(os.Setenv("CF_DOCKER_PASSWORD", "my-docker-password")).To(Succeed())
	})

	AfterEach(func() {
		Expect(os.Setenv("CF_DOCKER_PASSWORD", oldDockerPassword)).To(Succeed())
		Expect(os.Remove(tempFile)).ToNot(HaveOccurred())

		helpers.QuickDeleteOrg(orgName)
	})

	Context("when the specified manifest file does not exist", func() {
		It("displays a path does not exist error, help, and exits 1", func() {
			session := helpers.CF("push", "-f", "./non-existent-file")
			Eventually(session.Err).Should(Say("Incorrect Usage: The specified path './non-existent-file' does not exist."))
			Eventually(session.Out).Should(Say("NAME:"))
			Eventually(session.Out).Should(Say("USAGE:"))
			Eventually(session).Should(Exit(1))
		})
	})

	Context("when the same docker property is provided via both manifest and command line", func() {
		Context("when manifest contains 'docker.image' and the '--docker-image' flag is provided", func() {
			BeforeEach(func() {
				manifestContents := []byte(fmt.Sprintf(`
---
applications:
- name: %s
  docker:
    image: some-image
`, appName))
				Expect(ioutil.WriteFile(tempFile, manifestContents, 0666)).To(Succeed())
			})

			It("overrides 'docker.image' in the manifest with the '-o' flag value", func() {
				Eventually(helpers.CF("push", "-o", DockerImage, "-f", tempFile)).Should(Exit(0))

				appGUID := helpers.AppGUID(appName)
				session := helpers.CF("curl", fmt.Sprintf("/v2/apps/%s", appGUID))
				Eventually(session.Out).Should(Say(DockerImage))
				Eventually(session).Should(Exit(0))
			})
		})

		Context("when manifest contains 'docker.username' and the '--docker-username' flag is provided", func() {
			var buffer *Buffer

			BeforeEach(func() {
				buffer = NewBuffer()
				_, err := buffer.Write([]byte("n\n"))
				Expect(err).NotTo(HaveOccurred())

				manifestContents := []byte(fmt.Sprintf(`
---
applications:
- name: %s
  docker:
    image: some-image
    username: some-user
`, appName))
				Expect(ioutil.WriteFile(tempFile, manifestContents, 0666)).To(Succeed())
			})

			It("overrides 'docker.username' in the manifest with the '--docker-username' flag value", func() {
				Eventually(helpers.CFWithStdin(buffer, "push", "--docker-username", "some-other-user", "-f", tempFile)).Should(Exit())

				appGUID := helpers.AppGUID(appName)
				session := helpers.CF("curl", fmt.Sprintf("/v2/apps/%s", appGUID))
				Eventually(session.Out).Should(Say("some-other-user"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	Context("when the docker password is set in the environment", func() {
		Context("when the docker image is provided via the command line and docker username is provided via the manifest", func() {
			BeforeEach(func() {
				manifestContents := []byte(fmt.Sprintf(`
---
applications:
- name: %s
  docker:
    username: some-other-user
`, appName))
				Expect(ioutil.WriteFile(tempFile, manifestContents, 0666)).To(Succeed())
			})

			It("pushes the app using the docker image from command line and username from manifest", func() {
				Eventually(helpers.CF("push", "-o", DockerImage, "-f", tempFile)).Should(Exit())

				appGUID := helpers.AppGUID(appName)
				session := helpers.CF("curl", fmt.Sprintf("/v2/apps/%s", appGUID))
				Eventually(session.Out).Should(Say(DockerImage))
				Eventually(session.Out).Should(Say("some-other-user"))
				Eventually(session).Should(Exit(0))
			})
		})

		Context("when the docker image is provided via the manifest and docker username is provided via the command line", func() {
			var buffer *Buffer

			BeforeEach(func() {
				buffer = NewBuffer()
				_, err := buffer.Write([]byte("my-voice-is-my-passport\n"))
				Expect(err).NotTo(HaveOccurred())

				manifestContents := []byte(fmt.Sprintf(`
---
applications:
- name: %s
  docker:
    image: %s
`, appName, DockerImage))
				Expect(ioutil.WriteFile(tempFile, manifestContents, 0666)).To(Succeed())
			})

			It("pushes the app using the docker image from manifest and username from command line", func() {
				Eventually(helpers.CFWithStdin(buffer, "push", "--docker-username", "some-user", "-f", tempFile)).Should(Exit())

				appGUID := helpers.AppGUID(appName)
				session := helpers.CF("curl", fmt.Sprintf("/v2/apps/%s", appGUID))
				Eventually(session.Out).Should(Say(DockerImage))
				Eventually(session.Out).Should(Say("some-user"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	Context("when the docker password is not set in the environment", func() {
		BeforeEach(func() {
			Expect(os.Unsetenv("CF_DOCKER_PASSWORD")).To(Succeed())
		})

		Context("when the docker username is provided via the command line", func() {
			var buffer *Buffer

			BeforeEach(func() {
				buffer = NewBuffer()
				_, err := buffer.Write([]byte("my-voice-is-my-passport\n"))
				Expect(err).NotTo(HaveOccurred())
			})

			It("prompts the user for the docker password", func() {
				session := helpers.CFWithStdin(buffer, "push", appName, "--docker-image", DockerImage, "--docker-username", "some-user")
				Eventually(session).Should(Say("Environment variable CF_DOCKER_PASSWORD not set\\."))
				Eventually(session).Should(Exit())
			})
		})

		Context("when the docker username is provided via the manifest", func() {
			BeforeEach(func() {
				manifestContents := []byte(fmt.Sprintf(`
---
applications:
- name: %s
  docker:
    image: some-image
    username: some-user
`, appName))
				Expect(ioutil.WriteFile(tempFile, manifestContents, 0666)).To(Succeed())
			})

			It("displays an error and exits 1", func() {
				session := helpers.CF("push", "-f", tempFile)
				Eventually(session).Should(Say("No Docker password was provided\\. Please provide the password by setting the CF_DOCKER_PASSWORD environment variable\\."))
				Eventually(session).Should(Exit(1))
			})
		})
	})

	Context("when invalid manifest properties are provided together", func() {
		Context("manifest contains both 'buildpack' and 'docker.image'", func() {
			BeforeEach(func() {
				manifestContents := []byte(fmt.Sprintf(`
---
applications:
- name: %s
  buildpack: staticfile_buildpack
  docker:
    image: %s
`, appName, DockerImage))
				Expect(ioutil.WriteFile(tempFile, manifestContents, 0666)).To(Succeed())
			})

			It("displays an error and exits 1", func() {
				session := helpers.CF("push", appName, "-f", tempFile)
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Say("Invalid application configuration:"))
				Eventually(session).Should(Say("Application %s must not be configured with both 'buildpack' and 'docker'", appName))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("manifest contains both 'docker.image' and 'path'", func() {
			BeforeEach(func() {
				manifestContents := []byte(fmt.Sprintf(`
---
applications:
- name: %s
  path: .
  docker:
    image: %s
`, appName, DockerImage))
				Expect(ioutil.WriteFile(tempFile, manifestContents, 0666)).To(Succeed())
			})

			It("displays an error and exits 1", func() {
				session := helpers.CF("push", appName, "-f", tempFile)
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Say("Invalid application configuration:"))
				Eventually(session).Should(Say("Application %s must not be configured with both 'docker' and 'path'", appName))
				Eventually(session).Should(Exit(1))
			})
		})
	})
})
