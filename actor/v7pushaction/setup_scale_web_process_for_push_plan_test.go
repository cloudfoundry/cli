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

var _ = Describe("SetupScaleWebProcessForPushPlan", func() {
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
		expectedPushPlan, executeErr = SetupScaleWebProcessForPushPlan(pushPlan, manifestApp)
	})

	When("disk, instances, and memory are not set", func() {
		It("skips the ScaleWebProcess on the push plan", func() {
			Expect(executeErr).ToNot(HaveOccurred())

			Expect(expectedPushPlan.ScaleWebProcess).To(Equal(v7action.Process{}))
			Expect(expectedPushPlan.ScaleWebProcessNeedsUpdate).To(BeFalse())
		})
	})

	When("when the disk is set on flag overrides", func() {
		BeforeEach(func() {
			pushPlan.Overrides.Disk = types.NullUint64{IsSet: true, Value: 555}
		})

		It("sets the disk on the push plan", func() {
			Expect(executeErr).ToNot(HaveOccurred())

			Expect(expectedPushPlan.ScaleWebProcess).To(Equal(v7action.Process{
				Type:     constant.ProcessTypeWeb,
				DiskInMB: types.NullUint64{IsSet: true, Value: 555},
			}))
			Expect(expectedPushPlan.ScaleWebProcessNeedsUpdate).To(BeTrue())
		})
	})

	When("when the instances is set on flag overrides", func() {
		BeforeEach(func() {
			pushPlan.Overrides.Instances = types.NullInt{IsSet: true, Value: 555}
		})

		It("sets the instances on the push plan", func() {
			Expect(executeErr).ToNot(HaveOccurred())

			Expect(expectedPushPlan.ScaleWebProcess).To(Equal(v7action.Process{
				Type:      constant.ProcessTypeWeb,
				Instances: types.NullInt{IsSet: true, Value: 555},
			}))
			Expect(expectedPushPlan.ScaleWebProcessNeedsUpdate).To(BeTrue())
		})
	})

	When("when the memory is set on flag overrides", func() {
		BeforeEach(func() {
			pushPlan.Overrides.Memory = types.NullUint64{IsSet: true, Value: 555}
		})

		It("sets the memory on the push plan", func() {
			Expect(executeErr).ToNot(HaveOccurred())

			Expect(expectedPushPlan.ScaleWebProcess).To(Equal(v7action.Process{
				Type:       constant.ProcessTypeWeb,
				MemoryInMB: types.NullUint64{IsSet: true, Value: 555},
			}))
			Expect(expectedPushPlan.ScaleWebProcessNeedsUpdate).To(BeTrue())
		})
	})
})
