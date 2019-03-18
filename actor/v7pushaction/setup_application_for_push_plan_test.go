package v7pushaction_test

import (
	. "code.cloudfoundry.org/cli/actor/v7pushaction"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/util/manifestparser"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SetupApplicationForPushPlan", func() {
	var (
		pushPlan    PushPlan
		overrides   FlagOverrides
		manifestApp manifestparser.Application

		expectedPushPlan PushPlan
		executeErr       error

		appName string
	)

	BeforeEach(func() {
		appName = "some-app-name"

		pushPlan = PushPlan{}
		overrides = FlagOverrides{}
		manifestApp = manifestparser.Application{}
		manifestApp.Name = appName
	})

	JustBeforeEach(func() {
		expectedPushPlan, executeErr = SetupApplicationForPushPlan(pushPlan, overrides, manifestApp)
	})

	AssertNameIsSet := func() {
		It("sets the name", func() {
			Expect(expectedPushPlan.Application.Name).To(Equal(appName))
		})
	}

	AssertNoExecuteErr := func() {
		It("returns nil", func() {
			Expect(executeErr).ToNot(HaveOccurred())
		})
	}

	Describe("LifecycleType", func() {
		When("our overrides contain a DockerImage", func() {
			BeforeEach(func() {
				overrides.DockerImage = "docker://yes/yes"
			})

			It("creates a pushPlan with an app with LifecycleType docker", func() {
				Expect(expectedPushPlan.Application.LifecycleType).
					To(Equal(constant.AppLifecycleTypeDocker))
			})

			AssertNoExecuteErr()
			AssertNameIsSet()
		})

		When("our overrides do not contain a DockerImage", func() {
			When("The app manifest contains a Docker image", func() {
				BeforeEach(func() {
					manifestApp.Docker = new(manifestparser.Docker)
					manifestApp.Docker.Image = "docker-image"
				})
				It("creates a pushPlan with an app with LifecycleType docker", func() {
					Expect(expectedPushPlan.Application.LifecycleType).
						To(Equal(constant.AppLifecycleTypeDocker))
				})

				AssertNoExecuteErr()
				AssertNameIsSet()
			})

			When("The app manifest does not contain a Docker image", func() {
				It("Creates a pushPlan with an app without LifecycleType Docker", func() {
					Expect(expectedPushPlan.Application.LifecycleType).
						ToNot(Equal(constant.AppLifecycleTypeDocker))
				})

				AssertNoExecuteErr()
				AssertNameIsSet()
			})
		})
	})

	Describe("Buildpacks", func() {
		When("our overrides contain one or more buildpacks", func() {
			BeforeEach(func() {
				overrides.Buildpacks = []string{"buildpack-1", "buildpack-2"}
			})

			It("creates a pushPlan with an app with Buildpacks set", func() {
				Expect(expectedPushPlan.Application.LifecycleBuildpacks).To(Equal(
					[]string{"buildpack-1", "buildpack-2"},
				))
			})

			It("creates a pushPlan with an app with LifecycleType set to Buildpacks", func() {
				Expect(expectedPushPlan.Application.LifecycleType).
					To(Equal(constant.AppLifecycleTypeBuildpack))
			})

			It("creates a pushPlan with an app with applicationNeedsUpdate set", func() {
				Expect(expectedPushPlan.ApplicationNeedsUpdate).To(BeTrue())
			})

			AssertNoExecuteErr()
			AssertNameIsSet()
		})

		When("our overrides do not contain buildpacks", func() {
			It("creates a pushPlan with an app without Buildpacks set", func() {
				Expect(expectedPushPlan.Application.LifecycleBuildpacks).
					To(HaveLen(0))
			})

			It("creates a pushPlan with an app without applicationNeedsUpdate", func() {
				Expect(expectedPushPlan.ApplicationNeedsUpdate).To(BeFalse())
			})

			AssertNoExecuteErr()
			AssertNameIsSet()
		})
	})

	Describe("Stacks", func() {
		When("our overrides contain a stack", func() {
			BeforeEach(func() {
				overrides.Stack = "stack"
			})

			It("creates a pushPlan with an application with StackName matching the override", func() {
				Expect(expectedPushPlan.Application.StackName).To(Equal("stack"))
			})

			It("creates a pushPlan with an app with LifecycleType set to Buildpacks", func() {
				Expect(expectedPushPlan.Application.LifecycleType).
					To(Equal(constant.AppLifecycleTypeBuildpack))
			})

			It("creates a pushPlan with an app with ApplicationNeedsUpdate set", func() {
				Expect(expectedPushPlan.ApplicationNeedsUpdate).To(BeTrue())
			})

			AssertNoExecuteErr()
			AssertNameIsSet()
		})

		When("our overrides do not contain a stack", func() {
			It("creates a pushPlan with an app without StackName set", func() {
				Expect(expectedPushPlan.Application.StackName).To(BeEmpty())
			})

			It("creates a pushPlan with an app without Buildpacks set", func() {
				Expect(expectedPushPlan.Application.LifecycleBuildpacks).
					To(HaveLen(0))
			})

			It("creates a pushPlan with an app without applicationNeedsUpdate", func() {
				Expect(expectedPushPlan.ApplicationNeedsUpdate).To(BeFalse())
			})
		})
	})
})
