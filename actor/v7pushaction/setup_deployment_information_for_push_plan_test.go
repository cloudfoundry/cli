package v7pushaction_test

import (
	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccv3/constant"
    "code.cloudfoundry.org/cli/v8/cf/errors"
    "code.cloudfoundry.org/cli/v8/types"

	. "code.cloudfoundry.org/cli/v8/actor/v7pushaction"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("SetupDeploymentInformationForPushPlan", func() {
	var (
		pushPlan  PushPlan
		overrides FlagOverrides

		expectedPushPlan PushPlan
		executeErr       error
	)

	// A helper function to encapsulate the common flag override tests.
	// This function will be called within the context of different strategy tests.
	runCommonFlagOverrideTests := func() {
		When("instance overrides is specified", func() {
			BeforeEach(func() {
				overrides.Instances = types.NullInt{IsSet: true, Value: 3}
			})
			It("should set the instances", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(expectedPushPlan.Instances).To(Equal(types.NullInt{IsSet: true, Value: 3}))
			})
		})

		When("memory overrides is specified with incorrect unit", func() {
			var expectedErr error
			BeforeEach(func() {
				overrides.Memory = "10k"
				expectedErr = errors.New("Byte quantity must be an integer with a unit of measurement like M, MB, G, or GB")
			})
			It("should return error", func() {
				Expect(executeErr).To(HaveOccurred())
				Expect(executeErr).To(MatchError(expectedErr))
			})
		})

		When("memory overrides is specified in GB", func() {
			BeforeEach(func() {
				overrides.Memory = "1GB"
			})
			It("should set the memory in MB", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(expectedPushPlan.MemoryInMB).To(Equal(types.NullUint64{IsSet: true, Value: 1 * 1024}))
			})
		})

		When("memory overrides is specified in MB", func() {
			BeforeEach(func() {
				overrides.Memory = "1M"
			})
			It("should set the memory in MB", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(expectedPushPlan.MemoryInMB).To(Equal(types.NullUint64{IsSet: true, Value: 1}))
			})
		})

		When("disk overrides is specified with incorrect unit", func() {
			var expectedErr error
			BeforeEach(func() {
				overrides.Disk = "10k"
				expectedErr = errors.New("Byte quantity must be an integer with a unit of measurement like M, MB, G, or GB")
			})
			It("should return error", func() {
				Expect(executeErr).To(HaveOccurred())
				Expect(executeErr).To(MatchError(expectedErr))
			})
		})

		When("disk overrides is specified in GB", func() {
			BeforeEach(func() {
				overrides.Disk = "2GB"
			})
			It("should set the disk in MB", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(expectedPushPlan.DiskInMB).To(Equal(types.NullUint64{IsSet: true, Value: 2 * 1024}))
			})
		})

		When("disk overrides is specified in MB", func() {
			BeforeEach(func() {
				overrides.Disk = "1M"
			})
			It("should set the disk in MB", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(expectedPushPlan.DiskInMB).To(Equal(types.NullUint64{IsSet: true, Value: 1}))
			})
		})

		When("log rate limit is specified with incorrect unit", func() {
			var expectedErr error
			BeforeEach(func() {
				overrides.LogRateLimit = "10A"
				expectedErr = errors.New("Byte quantity must be an integer with a unit of measurement like B, K, KB, M, MB, G, or GB")
			})
			It("should return error", func() {
				Expect(executeErr).To(HaveOccurred())
				Expect(executeErr).To(MatchError(expectedErr))
			})
		})

		When("unlimited log rate limit is specified", func() {
			BeforeEach(func() {
				overrides.LogRateLimit = "-1"
			})
			It("should set the log rate limit", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(expectedPushPlan.LogRateLimitInBPS).To(Equal(types.NullInt{IsSet: true, Value: -1}))
			})
		})

		When("log rate limit overrides is specified in Bytes", func() { // Original comment was "disk overrides", corrected
			BeforeEach(func() {
				overrides.LogRateLimit = "10B"
			})
			It("should set the log rate limit in Bytes", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(expectedPushPlan.LogRateLimitInBPS).To(Equal(types.NullInt{IsSet: true, Value: 10}))
			})
		})

		When("log rate limit overrides is specified in KB", func() {
			BeforeEach(func() {
				overrides.LogRateLimit = "2K"
			})
			It("should set the log rate limit in Bytes", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(expectedPushPlan.LogRateLimitInBPS).To(Equal(types.NullInt{IsSet: true, Value: 2 * 1024}))
			})
		})

		When("log rate limit overrides is specified in MB", func() { // Original comment was "disk overrides", corrected
			BeforeEach(func() {
				overrides.LogRateLimit = "1MB"
			})
			It("should set the log rate limit in Bytes", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(expectedPushPlan.LogRateLimitInBPS).To(Equal(types.NullInt{IsSet: true, Value: 1 * 1024 * 1024}))
			})
		})
	}

	BeforeEach(func() {
		pushPlan = PushPlan{}
		overrides = FlagOverrides{}
	})

	JustBeforeEach(func() {
		expectedPushPlan, executeErr = SetupDeploymentInformationForPushPlan(pushPlan, overrides)
	})

	When("flag overrides specifies strategy", func() {
		BeforeEach(func() {
			// These values are common for both rolling and canary when a strategy is specified
			maxInFlight := 5
			overrides.MaxInFlight = &maxInFlight
			overrides.InstanceSteps = []int64{1, 2, 3, 4}
		})

		DescribeTableSubtree("sets strategy and related options correctly",
			func(strategy constant.DeploymentStrategy, expectedDeploymentStrategy constant.DeploymentStrategy, expectedInstanceSteps []int64) {
				BeforeEach(func() {
					overrides.Strategy = strategy
				})

				It("sets the strategy on the push plan", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(expectedPushPlan.Strategy).To(Equal(expectedDeploymentStrategy))
				})

				It("sets the max in flight on the push plan", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(expectedPushPlan.MaxInFlight).To(Equal(5))
				})

				It("sets the instance steps correctly", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					if len(expectedInstanceSteps) > 0 {
						Expect(expectedPushPlan.InstanceSteps).To(ContainElements(expectedInstanceSteps))
					} else {
						Expect(expectedPushPlan.InstanceSteps).To(BeEmpty())
					}
				})

				runCommonFlagOverrideTests()
			},
			Entry("when strategy is rolling", constant.DeploymentStrategyRolling, constant.DeploymentStrategyRolling, []int64{}), // No instance steps for rolling
			Entry("when strategy is canary", constant.DeploymentStrategyCanary, constant.DeploymentStrategyCanary, []int64{1, 2, 3, 4}),
		)
	})

	When("flag overrides does not specify strategy", func() {
		BeforeEach(func() {
			maxInFlight := 10
			overrides.MaxInFlight = &maxInFlight
			overrides.InstanceSteps = []int64{1, 2, 3, 4}
			overrides.Instances = types.NullInt{IsSet: true, Value: 3}
			overrides.Memory = "10k"
			overrides.Disk = "20K"
			overrides.LogRateLimit = "20K"
		})

		It("leaves the strategy as its default value on the push plan", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(expectedPushPlan.Strategy).To(Equal(constant.DeploymentStrategyDefault))
		})

		It("does not set MaxInFlight", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(expectedPushPlan.MaxInFlight).To(Equal(0))
		})

		It("does not set canary steps", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(expectedPushPlan.InstanceSteps).To(BeEmpty())
		})

		It("does not set instances", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(expectedPushPlan.Instances).To(Equal(types.NullInt{IsSet: false, Value: 0}))
		})
		It("does not set Memory", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(expectedPushPlan.MemoryInMB).To(Equal(types.NullUint64{IsSet: false, Value: 0}))
		})
		It("does not set Disk", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(expectedPushPlan.DiskInMB).To(Equal(types.NullUint64{IsSet: false, Value: 0}))
		})
		It("does not set log rate limit", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(expectedPushPlan.LogRateLimitInBPS).To(Equal(types.NullInt{IsSet: false, Value: 0}))
		})
	})

	When("no flag overrides are provided", func() {
		BeforeEach(func() {
			overrides = FlagOverrides{}
		})
		It("does not set MaxInFlight", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(expectedPushPlan.MaxInFlight).To(Equal(0))
		})
		It("does not set the canary steps", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(expectedPushPlan.InstanceSteps).To(BeEmpty())
		})
		It("does not set instances", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(expectedPushPlan.Instances).To(Equal(types.NullInt{IsSet: false, Value: 0}))
		})
		It("does not set Memory", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(expectedPushPlan.MemoryInMB).To(Equal(types.NullUint64{IsSet: false, Value: 0}))
		})
		It("does not set Disk", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(expectedPushPlan.DiskInMB).To(Equal(types.NullUint64{IsSet: false, Value: 0}))
		})
		It("does not set log rate limit", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(expectedPushPlan.LogRateLimitInBPS).To(Equal(types.NullInt{IsSet: false, Value: 0}))
		})
		It("leaves the strategy as its default value on the push plan", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(expectedPushPlan.Strategy).To(Equal(constant.DeploymentStrategyDefault))
		})
	})
})
