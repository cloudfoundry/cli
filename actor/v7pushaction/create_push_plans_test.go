package v7pushaction_test

import (
	. "code.cloudfoundry.org/cli/actor/v7pushaction"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/util/manifestparser"
	"code.cloudfoundry.org/cli/util/manifestparser/manifestparserfakes"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CreatePushPlans", func() {
	var (
		pushActor *Actor

		fakeManifestParser *manifestparserfakes.FakeManifestParser

		flagOverrides FlagOverrides

		pushPlans []PushPlan

		executeErr error

		appNameArg string

		spaceGUID string

		orgGUID string
	)

	BeforeEach(func() {
		pushActor, _, _, _ = getTestPushActor()
		fakeManifestParser = new(manifestparserfakes.FakeManifestParser)
		appNameArg = "my-app"
		flagOverrides = FlagOverrides{}
		orgGUID = "org"
		spaceGUID = "space"
	})

	JustBeforeEach(func() {
		pushPlans, executeErr = pushActor.CreatePushPlans(appNameArg, spaceGUID, orgGUID, fakeManifestParser, flagOverrides)
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

	Describe("Manifest", func() {
		When("There are multiple apps", func() {
			BeforeEach(func() {
				fakeManifestParser.AppsReturns([]manifestparser.Application{
					{
						ApplicationModel: manifestparser.ApplicationModel{
							Name: "my-app",
						},
						FullUnmarshalledApplication: nil,
					},
					{
						ApplicationModel: manifestparser.ApplicationModel{
							Name: "spencers-app",
							Path: "spencers/path",
						},
						FullUnmarshalledApplication: nil,
					},
				}, nil)

				fakeManifestParser.FullRawManifestReturns([]byte("not-nil"))

				appNameArg = ""
			})

			AssertNoExecuteErr()
			AssertPushPlanLength(2)

			It("it creates pushPlans based on the apps in the manifest", func() {
				Expect(pushPlans[0].Application.Name).To(Equal("my-app"))
				Expect(pushPlans[1].Application.Name).To(Equal("spencers-app"))
				Expect(pushPlans[0].BitsPath).To(Equal(getCurrentDir()))
				Expect(pushPlans[1].BitsPath).To(Equal("spencers/path"))
			})
		})

		When("There is an appName specified", func() {
			When("And that appName is NOT present in the manifest", func() {
				BeforeEach(func() {
					fakeManifestParser.AppsReturns(nil, manifestparser.AppNotInManifestError{Name: appNameArg})

					fakeManifestParser.FullRawManifestReturns([]byte("not-nil"))

					appNameArg = "my-app"
				})

				It("it returns an AppNotInManifestError", func() {
					Expect(executeErr).To(MatchError(manifestparser.AppNotInManifestError{Name: appNameArg}))
				})
			})
			When("And that appName is present in the manifest", func() {
				BeforeEach(func() {
					fakeManifestParser.AppsReturns([]manifestparser.Application{
						{
							ApplicationModel: manifestparser.ApplicationModel{
								Name: "my-app",
							},
							FullUnmarshalledApplication: nil,
						},
					}, nil)

					fakeManifestParser.FullRawManifestReturns([]byte("not-nil"))

					appNameArg = "my-app"
					flagOverrides.DockerImage = "image"
				})

				AssertNoExecuteErr()
				AssertPushPlanLength(1)

				It("it creates pushPlans based on the named app in the manifest", func() {
					Expect(pushPlans[0].Application.Name).To(Equal("my-app"))
					Expect(pushPlans[0].BitsPath).To(Equal(getCurrentDir()))
					Expect(pushPlans[0].Application.LifecycleType).To(Equal(constant.AppLifecycleTypeDocker))
				})
			})
		})
	})

	FDescribe("Org and Space guid", func() {
		It("creates pushPlans with org and space guids", func() {
			Expect(pushPlans[0].SpaceGUID).To(Equal(spaceGUID))
			Expect(pushPlans[0].OrgGUID).To(Equal(orgGUID))
		})
	})
})
