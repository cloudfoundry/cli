package v7pushaction_test

import (
	. "code.cloudfoundry.org/cli/actor/v7pushaction"
	"code.cloudfoundry.org/cli/util/pushmanifestparser"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SetupBitsPathForPushPlan", func() {
	var (
		pushPlan    PushPlan
		overrides   FlagOverrides
		manifestApp pushmanifestparser.Application

		expectedPushPlan PushPlan
		executeError     error
	)

	BeforeEach(func() {
		pushPlan = PushPlan{}
		overrides = FlagOverrides{}
		manifestApp = pushmanifestparser.Application{}
	})

	JustBeforeEach(func() {
		expectedPushPlan, executeError = SetupBitsPathForPushPlan(pushPlan, overrides, manifestApp)
	})

	Describe("Path", func() {
		When("overrides contain a path", func() {
			BeforeEach(func() {
				overrides.ProvidedAppPath = "some/path"
			})

			It("creates a pushPlan with an app with BitsPath set", func() {
				Expect(executeError).ToNot(HaveOccurred())
				Expect(expectedPushPlan.BitsPath).To(Equal("some/path"))
			})
		})

		When("manifest contains a path", func() {
			BeforeEach(func() {
				manifestApp.Path = "some/path"
			})

			It("creates a pushPlan with an app with BitsPath set", func() {
				Expect(executeError).ToNot(HaveOccurred())
				Expect(expectedPushPlan.BitsPath).To(Equal("some/path"))
			})
		})

		When("neither overrides nor manifest contain a path", func() {
			It("creates a pushPlan with an app with BitsPath set to the currentDir", func() {
				Expect(executeError).ToNot(HaveOccurred())
				Expect(expectedPushPlan.BitsPath).To(Equal(getCurrentDir()))
			})
		})
	})
})
