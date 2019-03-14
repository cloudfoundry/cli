package v7pushaction_test

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/manifestparser"

	"code.cloudfoundry.org/cli/actor/v7action"
	. "code.cloudfoundry.org/cli/actor/v7pushaction"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SetupUpdateWebProcessForPushPlan", func() {
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
		expectedPushPlan, executeErr = SetupUpdateWebProcessForPushPlan(pushPlan, manifestApp)
	})

	When("start command, health check type, and health check timeout are not set", func() {
		It("skips the UpdateWebProcess on the push plan", func() {
			Expect(executeErr).ToNot(HaveOccurred())

			Expect(expectedPushPlan.UpdateWebProcess).To(Equal(v7action.Process{}))
			Expect(expectedPushPlan.UpdateWebProcessNeedsUpdate).To(BeFalse())
		})
	})

	When("the start command is set on flag overrides", func() {
		BeforeEach(func() {
			pushPlan.Overrides.StartCommand = types.FilteredString{IsSet: true, Value: "some-command"}
		})

		It("sets the start command on the push plan", func() {
			Expect(executeErr).ToNot(HaveOccurred())

			Expect(expectedPushPlan.UpdateWebProcess).To(Equal(v7action.Process{
				Command: types.FilteredString{IsSet: true, Value: "some-command"},
			}))
			Expect(expectedPushPlan.UpdateWebProcessNeedsUpdate).To(BeTrue())
		})
	})

	When("the health check type is set on flag overrides", func() {
		BeforeEach(func() {
			pushPlan.Overrides.HealthCheckType = constant.HTTP
			pushPlan.Overrides.HealthCheckEndpoint = "/potato"
		})

		It("sets the health check type and endpoint on the push plan", func() {
			Expect(executeErr).ToNot(HaveOccurred())

			Expect(expectedPushPlan.UpdateWebProcess).To(Equal(v7action.Process{
				HealthCheckType:     constant.HTTP,
				HealthCheckEndpoint: "/potato",
			}))
			Expect(expectedPushPlan.UpdateWebProcessNeedsUpdate).To(BeTrue())
		})
	})

	When("the health check timeout is set on flag overrides", func() {
		BeforeEach(func() {
			pushPlan.Overrides.HealthCheckTimeout = 100
		})

		It("sets the health check timeout on the push plan", func() {
			Expect(executeErr).ToNot(HaveOccurred())

			Expect(expectedPushPlan.UpdateWebProcess).To(Equal(v7action.Process{
				HealthCheckTimeout: 100,
			}))
			Expect(expectedPushPlan.UpdateWebProcessNeedsUpdate).To(BeTrue())
		})
	})
})
