package models_test

import (
	"os"

	. "code.cloudfoundry.org/cli/cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Merge", func() {
	var (
		appParams   AppParams
		flagContext AppParams
	)

	BeforeEach(func() {
		appParams = AppParams{}
		flagContext = AppParams{}
	})

	JustBeforeEach(func() {
		appParams.Merge(&flagContext)
	})

	Context("when the Docker username is set in neither flag context nor manifest", func() {
		It("leaves the merged Docker password as nil", func() {
			Expect(appParams.DockerPassword).To(BeNil())
		})
	})

	Context("when the Docker username is set in the flag context", func() {
		BeforeEach(func() {
			user := "my-user"
			flagContext.DockerUsername = &user

			password := "my-pass"
			flagContext.DockerPassword = &password
		})

		It("copies the Docker password from the flag context", func() {
			Expect(appParams.DockerPassword).NotTo(BeNil())
			Expect(*appParams.DockerPassword).To(Equal("my-pass"))
		})
	})

	Context("when the Docker username is not set in the flag context but is set in the manifest", func() {
		var oldDockerPassword string

		BeforeEach(func() {
			oldDockerPassword = os.Getenv("CF_DOCKER_PASSWORD")
			Expect(os.Setenv("CF_DOCKER_PASSWORD", "some-docker-pass")).ToNot(HaveOccurred())

			password := "should-not-be-me"
			flagContext.DockerPassword = &password

			user := "my-manifest-user"
			appParams.DockerUsername = &user
		})

		AfterEach(func() {
			Expect(os.Setenv("CF_DOCKER_PASSWORD", oldDockerPassword)).ToNot(HaveOccurred())
		})

		It("grabs the Docker password from the env var CF_DOCKER_PASSWORD", func() {
			Expect(appParams.DockerPassword).NotTo(BeNil())
			Expect(*appParams.DockerPassword).To(Equal("some-docker-pass"))
		})
	})
})
