package v7pushaction_test

import (
	"code.cloudfoundry.org/cli/actor/v7action"
	. "code.cloudfoundry.org/cli/actor/v7pushaction"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SetupBitsPathForPushPlan", func() {
	var (
		pushPlan  PushPlan
		overrides FlagOverrides

		expectedPushPlan PushPlan
		executeError     error
	)

	BeforeEach(func() {
		pushPlan = PushPlan{}
		overrides = FlagOverrides{}
	})

	JustBeforeEach(func() {
		expectedPushPlan, executeError = SetDefaultBitsPathForPushPlan(pushPlan, overrides)
	})

	Describe("Path", func() {
		When("PushPlan contains a BitPath", func() {
			BeforeEach(func() {
				pushPlan.BitsPath = "myPath"
			})
			It("creates a pushPlan with an app with BitsPath set to the currentDir", func() {
				Expect(executeError).ToNot(HaveOccurred())
				Expect(expectedPushPlan.BitsPath).To(Equal("myPath"))
			})
		})

		When("PushPlan does not contain a BitPath", func() {
			When("Pushplan does not have droplet or docker", func() {
				It("creates a pushPlan with an app with BitsPath set to the currentDir", func() {
					Expect(executeError).ToNot(HaveOccurred())
					Expect(expectedPushPlan.BitsPath).To(Equal(getCurrentDir()))
				})
			})

			When("Pushplan has droplet", func() {
				BeforeEach(func() {
					pushPlan.DropletPath = "some-path"
				})

				It("does not set the BitsPath on the push plan", func() {
					Expect(executeError).ToNot(HaveOccurred())
					Expect(expectedPushPlan.BitsPath).To(Equal(""))
				})
			})

			When("Pushplan has docker", func() {
				BeforeEach(func() {
					pushPlan.DockerImageCredentials = v7action.DockerImageCredentials{
						Path:     "some-path",
						Username: "",
						Password: "",
					}
				})

				It("does not set the BitsPath on the push plan", func() {
					Expect(executeError).ToNot(HaveOccurred())
					Expect(expectedPushPlan.BitsPath).To(Equal(""))
				})
			})

		})
	})
})
