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
		manifestApp manifestparser.Application

		expectedPushPlan PushPlan
		executeErr       error
	)

	BeforeEach(func() {
		pushPlan = PushPlan{}
		manifestApp = manifestparser.Application{}
	})

	JustBeforeEach(func() {
		expectedPushPlan, executeErr = SetupDockerImageCredentialsForPushPlan(pushPlan, manifestApp)
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
				pushPlan.Overrides.DockerImage = "some-image"
				pushPlan.Overrides.DockerUsername = "some-username"
				pushPlan.Overrides.DockerPassword = "some-password"
			})

			It("sets the docker credentials on the push plan", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(expectedPushPlan.DockerImageCredentials).To(Equal(v7action.DockerImageCredentials{
					Path:     "some-image",
					Username: "some-username",
					Password: "some-password",
				}))
				Expect(expectedPushPlan.DockerImageCredentialsNeedsUpdate).To(BeTrue())
			})
		})
	})
})
