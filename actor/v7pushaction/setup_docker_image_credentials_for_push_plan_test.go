package v7pushaction_test

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/util/manifestparser"

	"code.cloudfoundry.org/cli/actor/v7action"
	. "code.cloudfoundry.org/cli/actor/v7pushaction"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SetupDockerImageCredentialsForPushPlan", func() {
	var (
		pushPlan    PushPlan
		overrides   FlagOverrides
		manifestApp manifestparser.Application

		expectedPushPlan PushPlan
		executeErr       error
	)

	BeforeEach(func() {
		pushPlan = PushPlan{}
		overrides = FlagOverrides{}
		manifestApp = manifestparser.Application{}
	})

	JustBeforeEach(func() {
		expectedPushPlan, executeErr = SetupDockerImageCredentialsForPushPlan(pushPlan, overrides, manifestApp)
	})

	When("the LifecycleType is not Docker", func() {
		It("skips the docker credentials on the push plan", func() {
			Expect(executeErr).ToNot(HaveOccurred())

			Expect(expectedPushPlan.DockerImageCredentials).To(Equal(v7action.DockerImageCredentials{}))
			Expect(expectedPushPlan.DockerImageCredentialsNeedsUpdate).To(BeFalse())
		})
	})

	When("when the LifecycleType is Docker", func() {
		BeforeEach(func() {
			pushPlan.Application.LifecycleType = constant.AppLifecycleTypeDocker
		})

		When("when the flag overrides contain docker settings", func() {
			BeforeEach(func() {
				overrides.DockerImage = "some-image"
				overrides.DockerUsername = "some-username"
				overrides.DockerPassword = "some-password"

				manifestApp.Docker = &manifestparser.Docker{
					Image:    "ignored-docker-image",
					Username: "ignored-docker-username",
				}
			})

			It("sets the flag overrides docker credentials on the push plan", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(expectedPushPlan.DockerImageCredentials).To(Equal(v7action.DockerImageCredentials{
					Path:     "some-image",
					Username: "some-username",
					Password: "some-password",
				}))
				Expect(expectedPushPlan.DockerImageCredentialsNeedsUpdate).To(BeTrue())
			})
		})

		When("the manifest contains flag overrides", func() {
			BeforeEach(func() {
				manifestApp.Docker = &manifestparser.Docker{
					Image:    "manifest-docker-image",
					Username: "manifest-docker-username",
				}
				overrides.DockerPassword = "some-password"
			})

			It("sets the manifest docker credentials on the push plan", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(expectedPushPlan.DockerImageCredentials).To(Equal(v7action.DockerImageCredentials{
					Path:     "manifest-docker-image",
					Username: "manifest-docker-username",
					Password: "some-password",
				}))
				Expect(expectedPushPlan.DockerImageCredentialsNeedsUpdate).To(BeTrue())
			})
		})
	})
})
