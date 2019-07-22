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

	Describe("GetUpdateSequence", func() {
		JustBeforeEach(func() {
			sequence = actor.GetUpdateSequence(plan)
		})

		When("the plan requires updating application", func() {
			BeforeEach(func() {
				plan = PushPlan{
					ApplicationNeedsUpdate: true,
					SkipRouteCreation:      true,
				}
			})

			It("returns a sequence including UpdateApplication", func() {
				Expect(sequence).To(matchers.MatchFuncsByName(actor.UpdateApplication))
			})
		})

		When("the plan has existing routes", func() {
			BeforeEach(func() {
				plan = PushPlan{
					ApplicationRoutes: []v7action.Route{
						{GUID: "some-route-guid"},
					},
				}
			})

			When("--random-route is given", func() {
				BeforeEach(func() {
					plan.RandomRoute = true
				})

				It("does not add the random route function", func() {
					Expect(sequence).To(BeEmpty())
				})
			})

			When("no-route is not given", func() {
				BeforeEach(func() {
					plan.NoRouteFlag = true
				})

				It("unmaps the existing route from the application", func() {
					Expect(sequence).To(matchers.MatchFuncsByName(actor.UnmapRoutesFromApplication))
				})
			})
		})

		When("the plan requires **no** route", func() {
			BeforeEach(func() {
				plan = PushPlan{}
			})

			It("returns a sequence including unmapping he routes from the app", func() {
			})
		})

		When("the plan requires updating routes", func() {
			BeforeEach(func() {
				plan = PushPlan{}
			})

			It("returns a sequence including updating the routes for the app", func() {
				Expect(sequence).To(matchers.MatchFuncsByName(actor.UpdateRoutesForApplicationWithDefaultRoute))
			})
		})

		When("the plan specifies random route", func() {
			BeforeEach(func() {
				plan = PushPlan{RandomRoute: true}
			})

			It("returns a sequence including updating the routes for the app", func() {
				Expect(sequence).To(matchers.MatchFuncsByName(actor.UpdateRoutesForApplicationWithRandomRoute))
			})
		})

		When("the plan requires scaling the web process", func() {
			BeforeEach(func() {
				plan = PushPlan{
					ScaleWebProcessNeedsUpdate: true,
					SkipRouteCreation:          true,
				}
			})

			It("returns a sequence including scaling web process", func() {
				Expect(sequence).To(matchers.MatchFuncsByName(actor.ScaleWebProcessForApplication))
			})
		})

		When("the plan requires updating the web process", func() {
			BeforeEach(func() {
				plan = PushPlan{
					UpdateWebProcessNeedsUpdate: true,
					SkipRouteCreation:           true,
				}
			})

			It("returns a sequence including updating a web process", func() {
				Expect(sequence).To(matchers.MatchFuncsByName(actor.UpdateWebProcessForApplication))
			})
		})
	})

	Describe("GetPrepareApplicationSourceSequence", func() {
		JustBeforeEach(func() {
			sequence = actor.GetPrepareApplicationSourceSequence(plan)
		})

		When("the plan requires creating a bits package", func() {
			BeforeEach(func() {
				plan = PushPlan{
					DockerImageCredentialsNeedsUpdate: false,
				}
			})

			It("returns a sequence including creating a bits package", func() {
				Expect(sequence).To(matchers.MatchFuncsByName(actor.CreateBitsPackageForApplication))
			})
		})

		When("the plan requires creating a docker package", func() {
			BeforeEach(func() {
				plan = PushPlan{
					DockerImageCredentialsNeedsUpdate: true,
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
	})
})
