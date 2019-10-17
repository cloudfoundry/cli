package v7pushaction_test

import (
	"code.cloudfoundry.org/cli/actor/v7action"
	. "code.cloudfoundry.org/cli/actor/v7pushaction"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Actor", func() {
	var (
		actor    *Actor
		plan     PushPlan
		sequence []ChangeApplicationFunc
	)

	BeforeEach(func() {
		actor, _, _ = getTestPushActor()
	})

	Describe("GetPrepareApplicationSourceSequence", func() {
		JustBeforeEach(func() {
			sequence = actor.GetPrepareApplicationSourceSequence(plan)
		})

		When("the plan requires creating a bits package", func() {
			BeforeEach(func() {
				plan = PushPlan{}
			})

			It("returns a sequence including creating a bits package", func() {
				Expect(sequence).To(matchers.MatchFuncsByName(actor.CreateBitsPackageForApplication))
			})
		})

		When("the plan requires creating a docker package", func() {
			BeforeEach(func() {
				plan = PushPlan{
					DockerImageCredentials: v7action.DockerImageCredentials{Path: "not empty"},
				}
			})

			It("returns a sequence including creating a docker package", func() {
				Expect(sequence).To(matchers.MatchFuncsByName(actor.CreateDockerPackageForApplication))
			})
		})

		When("the plan requires uploading a droplet", func() {
			BeforeEach(func() {
				plan = PushPlan{
					DropletPath: "path",
				}
			})

			It("returns a sequence including creating a droplet", func() {
				Expect(sequence).To(matchers.MatchFuncsByName(actor.CreateDropletForApplication))
			})
		})
	})

	Describe("GetRuntimeSequence", func() {
		JustBeforeEach(func() {
			sequence = actor.GetRuntimeSequence(plan)
		})

		When("the plan requires staging a package", func() {
			BeforeEach(func() {
				plan = PushPlan{}
			})

			It("returns a sequence including staging, setting the droplet, and restarting", func() {
				Expect(sequence).To(matchers.MatchFuncsByName(actor.StagePackageForApplication, actor.SetDropletForApplication, actor.RestartApplication))
			})
		})

		When("the plan requires stopping an app", func() {
			BeforeEach(func() {
				plan = PushPlan{
					NoStart:     true,
					Application: v7action.Application{State: constant.ApplicationStarted},
				}
			})

			It("returns a sequence including stopping the application", func() {
				Expect(sequence).To(matchers.MatchFuncsByName(actor.StopApplication))
			})
		})

		When("the plan requires setting a droplet", func() {
			BeforeEach(func() {
				plan = PushPlan{
					DropletPath: "path",
				}
			})

			It("returns a sequence including setting the droplet and restarting", func() {
				Expect(sequence).To(matchers.MatchFuncsByName(actor.SetDropletForApplication, actor.RestartApplication))
			})
		})

		When("the plan has strategy 'rolling'", func() {
			BeforeEach(func() {
				plan = PushPlan{
					Strategy: constant.DeploymentStrategyRolling,
				}
			})

			It("returns a sequence that creates a deployment without stopping/restarting the app", func() {
				Expect(sequence).To(matchers.MatchFuncsByName(actor.StagePackageForApplication, actor.CreateDeploymentForApplication))
			})
		})

		When("the plan has task application type", func() {
			BeforeEach(func() {
				plan = PushPlan{
					TaskTypeApplication: true,
				}
			})

			It("returns a sequence that stages, sets droplet, and stops the app", func() {
				Expect(sequence).To(matchers.MatchFuncsByName(actor.StagePackageForApplication, actor.StopApplication, actor.SetDropletForApplication))
			})
		})
	})
})
