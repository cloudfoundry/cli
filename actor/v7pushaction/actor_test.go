package v7pushaction_test

import (
	. "code.cloudfoundry.org/cli/actor/v7pushaction"
	"code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Actor", func() {
	var (
		actor *Actor
	)

	BeforeEach(func() {
		actor, _, _ = getTestPushActor()
	})

	Describe("PreparePushPlanSequence", func() {
		It("is a list of functions for preparing the push plan", func() {
			Expect(actor.PreparePushPlanSequence).To(matchers.MatchFuncsByName(
				SetDefaultBitsPathForPushPlan,
				SetupDropletPathForPushPlan,
				actor.SetupAllResourcesForPushPlan,
				SetupDeploymentStrategyForPushPlan,
				SetupNoStartForPushPlan,
				SetupNoWaitForPushPlan,
				SetupTaskAppForPushPlan,
			))
		})
	})
})
