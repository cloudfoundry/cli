package v7_test

import (
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo/v2"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("GetDockerPassword", func() {
	var (
		cmd        PushCommand
		fakeConfig *commandfakes.FakeConfig
		testUI     *ui.UI

		dockerUsername        string
		containsPrivateDocker bool

		executeErr     error
		dockerPassword string

		input *Buffer
	)

	BeforeEach(func() {
		input = NewBuffer()
		testUI = ui.NewTestUI(input, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)

		cmd = PushCommand{
			BaseCommand: BaseCommand{
				Config: fakeConfig,
				UI:     testUI,
			},
		}
	})

	Describe("Get", func() {
		JustBeforeEach(func() {
			dockerPassword, executeErr = cmd.GetDockerPassword(dockerUsername, containsPrivateDocker)
		})

		When("docker image is private", func() {
			When("there is a manifest", func() {
				BeforeEach(func() {
					dockerUsername = ""
					containsPrivateDocker = true
				})

				When("a password is provided via environment variable", func() {
					BeforeEach(func() {
						fakeConfig.DockerPasswordReturns("some-docker-password")
					})

					It("takes the password from the environment", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Out).ToNot(Say("Environment variable CF_DOCKER_PASSWORD not set."))
						Expect(testUI.Out).ToNot(Say("Docker password"))

						Expect(testUI.Out).To(Say("Using docker repository password from environment variable CF_DOCKER_PASSWORD."))

						Expect(dockerPassword).To(Equal("some-docker-password"))
					})
				})

				When("no password is provided", func() {
					BeforeEach(func() {
						_, err := input.Write([]byte("some-docker-password\n"))
						Expect(err).ToNot(HaveOccurred())
					})

					It("prompts for a password", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Out).To(Say("Environment variable CF_DOCKER_PASSWORD not set."))
						Expect(testUI.Out).To(Say("Docker password"))

						Expect(dockerPassword).To(Equal("some-docker-password"))
					})
				})
			})

			When("there is no manifest", func() {
				BeforeEach(func() {
					dockerUsername = "some-docker-username"
					containsPrivateDocker = false
				})

				When("a password is provided via environment variable", func() {
					BeforeEach(func() {
						fakeConfig.DockerPasswordReturns("some-docker-password")
					})

					It("takes the password from the environment", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Out).ToNot(Say("Environment variable CF_DOCKER_PASSWORD not set."))
						Expect(testUI.Out).ToNot(Say("Docker password"))

						Expect(testUI.Out).To(Say("Using docker repository password from environment variable CF_DOCKER_PASSWORD."))

						Expect(dockerPassword).To(Equal("some-docker-password"))
					})
				})

				When("no password is provided", func() {
					BeforeEach(func() {
						_, err := input.Write([]byte("some-docker-password\n"))
						Expect(err).ToNot(HaveOccurred())
					})

					It("prompts for a password", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Out).To(Say("Environment variable CF_DOCKER_PASSWORD not set."))
						Expect(testUI.Out).To(Say("Docker password"))

						Expect(dockerPassword).To(Equal("some-docker-password"))
					})
				})
			})
		})
		When("docker image is public", func() {
			BeforeEach(func() {
				dockerUsername = ""
				containsPrivateDocker = false
			})

			It("does not prompt for a password", func() {
				Expect(testUI.Out).ToNot(Say("Environment variable CF_DOCKER_PASSWORD not set."))
				Expect(testUI.Out).ToNot(Say("Docker password"))
				Expect(testUI.Out).ToNot(Say("Using docker repository password from environment variable CF_DOCKER_PASSWORD."))
			})

			It("returns an empty password", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(dockerPassword).To(Equal(""))
			})
		})
	})
})
