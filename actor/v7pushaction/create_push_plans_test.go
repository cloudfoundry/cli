package v7pushaction_test

import (
	. "code.cloudfoundry.org/cli/actor/v7pushaction"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/util/manifestparser"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CreatePushPlans", func() {
	var (
		pushActor *Actor

		manifestParser *manifestparser.Parser

		flagOverrides FlagOverrides

		pushPlans []PushPlan

		executeErr error

		appNameArg string
	)

	BeforeEach(func() {
		pushActor, _, _, _ = getTestPushActor()
		manifestParser = manifestparser.NewParser()
		appNameArg = "my-app"
		flagOverrides = FlagOverrides{}
	})

	JustBeforeEach(func() {
		pushPlans, executeErr = pushActor.CreatePushPlans(appNameArg, *manifestParser, flagOverrides)
	})

	AssertNameIsSet := func() {
		It("sets the name", func() {
			Expect(pushPlans[0].Application.Name).To(Equal(appNameArg))
		})
	}

	AssertNoExecuteErr := func() {
		It("returns nil", func() {
			Expect(executeErr).ToNot(HaveOccurred())
		})
	}

	AssertPushPlanLength := func(length int) {
		It(fmt.Sprintf("creates a []pushPlan with length %d", length), func() {
			Expect(pushPlans).To(HaveLen(length))
		})
	}

	Describe("LifecycleType", func() {
		When("our overrides contain a DockerImage", func() {
			BeforeEach(func() {
				flagOverrides.DockerImage = "docker://yes/yes"
			})

			It("creates a pushPlan with an app with LifecycleType docker", func() {
				Expect(pushPlans[0].Application.LifecycleType).
					To(Equal(constant.AppLifecycleTypeDocker))
			})

			AssertNoExecuteErr()
			AssertNameIsSet()
			AssertPushPlanLength(1)
		})

		When("our overrides do not contain a DockerImage", func() {
			BeforeEach(func() {
				flagOverrides.DockerImage = ""
			})

			It("Creates a pushPlan with an app without LifecycleType Buildpack", func() {
				Expect(pushPlans[0].Application.LifecycleType).
					ToNot(Equal(constant.AppLifecycleTypeDocker))
			})

			AssertNoExecuteErr()
			AssertNameIsSet()
			AssertPushPlanLength(1)
		})
	})

	Describe("Buildpacks", func() {
		When("our overrides contain one or more buildpacks", func() {
			BeforeEach(func() {
				flagOverrides.Buildpacks = []string{"buildpack-1", "buildpack-2"}
			})

			It("creates a pushPlan with an app with Buildpacks set", func() {
				Expect(pushPlans[0].Application.LifecycleBuildpacks).To(Equal(
					[]string{"buildpack-1", "buildpack-2"},
				))
			})

			It("creates a pushPlan with an app with LifecycleType set to Buildpacks", func() {
				Expect(pushPlans[0].Application.LifecycleType).
					To(Equal(constant.AppLifecycleTypeBuildpack))
			})

			It("creates a pushPlan with an app with applicationNeedsUpdate set", func() {
				Expect(pushPlans[0].ApplicationNeedsUpdate).To(BeTrue())
			})

			AssertNoExecuteErr()
			AssertNameIsSet()
			AssertPushPlanLength(1)
		})

		When("our overrides do not contain buildpacks", func() {
			BeforeEach(func() {
				flagOverrides.Buildpacks = []string{}
			})

			It("creates a pushPlan with an app without Buildpacks set", func() {
				Expect(pushPlans[0].Application.LifecycleBuildpacks).
					To(HaveLen(0))
			})

			It("creates a pushPlan with an app without applicationNeedsUpdate", func() {
				Expect(pushPlans[0].ApplicationNeedsUpdate).To(BeFalse())
			})

			AssertNoExecuteErr()
			AssertNameIsSet()
			AssertPushPlanLength(1)
		})
	})

	Describe("Stacks", func() {
		When("our overrides contain a stack", func() {
			BeforeEach(func() {
				flagOverrides.Stack = "stack"
			})

			It("creates a pushPlan with an application with StackName matching the override", func() {
				Expect(pushPlans[0].Application.StackName).To(Equal("stack"))
			})

			It("creates a pushPlan with an app with LifecycleType set to Buildpacks", func() {
				Expect(pushPlans[0].Application.LifecycleType).
					To(Equal(constant.AppLifecycleTypeBuildpack))
			})

			It("creates a pushPlan with an app with ApplicationNeedsUpdate set", func() {
				Expect(pushPlans[0].ApplicationNeedsUpdate).To(BeTrue())
			})

			AssertNoExecuteErr()
			AssertNameIsSet()
			AssertPushPlanLength(1)
		})

		When("our overrides do not contain a stack", func() {
			BeforeEach(func() {
				flagOverrides.Stack = ""
			})

			It("creates a pushPlan with an app without StackName set", func() {
				Expect(pushPlans[0].Application.StackName).To(BeEmpty())
			})

			It("creates a pushPlan with an app without Buildpacks set", func() {
				Expect(pushPlans[0].Application.LifecycleBuildpacks).
					To(HaveLen(0))
			})

			It("creates a pushPlan with an app without applicationNeedsUpdate", func() {
				Expect(pushPlans[0].ApplicationNeedsUpdate).To(BeFalse())
			})

			AssertNoExecuteErr()
			AssertNameIsSet()
			AssertPushPlanLength(1)
		})
	})

	Describe("Path", func() {
		When("our overrides contain a path", func() {
			BeforeEach(func() {
				flagOverrides.ProvidedAppPath = "some/path"
			})

			It("creates a pushPlan with an app with BitsPath set", func() {
				Expect(pushPlans[0].BitsPath).To(Equal("some/path"))
			})

			It("creates a pushPlan with an app without applicationNeedsUpdate", func() {
				Expect(pushPlans[0].ApplicationNeedsUpdate).To(BeFalse())
			})

			AssertNoExecuteErr()
			AssertNameIsSet()
			AssertPushPlanLength(1)
		})

		When("our overrides do not contain a path", func() {
			BeforeEach(func() {
				flagOverrides.ProvidedAppPath = ""
			})

			It("creates a pushPlan with an app with BitsPath set to the currentDir", func() {
				Expect(pushPlans[0].BitsPath).To(Equal(getCurrentDir()))
			})

			It("creates a pushPlan with an app without applicationNeedsUpdate", func() {
				Expect(pushPlans[0].ApplicationNeedsUpdate).To(BeFalse())
			})

			AssertNoExecuteErr()
			AssertNameIsSet()
			AssertPushPlanLength(1)
		})
	})
})
