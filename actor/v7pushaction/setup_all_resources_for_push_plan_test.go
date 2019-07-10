package v7pushaction_test

import (
	"errors"
	"io/ioutil"
	"os"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/util/manifestparser"

	. "code.cloudfoundry.org/cli/actor/v7pushaction"
	"code.cloudfoundry.org/cli/actor/v7pushaction/v7pushactionfakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SetupAllResourcesForPushPlan", func() {
	var (
		actor           *Actor
		fakeSharedActor *v7pushactionfakes.FakeSharedActor

		pushPlan    PushPlan
		overrides   FlagOverrides
		manifestApp manifestparser.Application

		expectedPushPlan PushPlan
		executeErr       error
	)

	BeforeEach(func() {
		actor, _, fakeSharedActor = getTestPushActor()

		pushPlan = PushPlan{}
		overrides = FlagOverrides{}
		manifestApp = manifestparser.Application{}
	})

	JustBeforeEach(func() {
		expectedPushPlan, executeErr = actor.SetupAllResourcesForPushPlan(pushPlan, overrides, manifestApp)
	})

	When("the plan has a droplet path", func() {
		BeforeEach(func() {
			pushPlan.DropletPath = "some-droplet.tgz"
		})

		It("skips settings the resources", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(pushPlan.AllResources).To(BeEmpty())

			Expect(fakeSharedActor.GatherArchiveResourcesCallCount()).To(Equal(0))
			Expect(fakeSharedActor.GatherDirectoryResourcesCallCount()).To(Equal(0))
		})
	})

	When("the application is a docker app", func() {
		BeforeEach(func() {
			pushPlan.Application.LifecycleType = constant.AppLifecycleTypeDocker
		})

		It("skips settings the resources", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(pushPlan.AllResources).To(BeEmpty())

			Expect(fakeSharedActor.GatherArchiveResourcesCallCount()).To(Equal(0))
			Expect(fakeSharedActor.GatherDirectoryResourcesCallCount()).To(Equal(0))
		})
	})

	When("the application is a buildpack app", func() {
		When("push plan's bits path is not set", func() {
			It("returns an error", func() {
				Expect(executeErr).To(MatchError("developer error: Bits Path needs to be set prior to generating app resources"))

				Expect(fakeSharedActor.GatherArchiveResourcesCallCount()).To(Equal(0))
				Expect(fakeSharedActor.GatherDirectoryResourcesCallCount()).To(Equal(0))
			})
		})

		When("the app resources are given as a directory", func() {
			var pwd string

			BeforeEach(func() {
				var err error
				pwd, err = os.Getwd()
				Expect(err).To(Not(HaveOccurred()))
				pushPlan.BitsPath = pwd
			})

			When("gathering the resources is successful", func() {
				var resources []sharedaction.Resource

				BeforeEach(func() {
					resources = []sharedaction.Resource{
						{
							Filename: "fake-app-file",
						},
					}
					fakeSharedActor.GatherDirectoryResourcesReturns(resources, nil)
				})

				It("adds the gathered resources to the push plan", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(fakeSharedActor.GatherDirectoryResourcesCallCount()).To(Equal(1))
					Expect(fakeSharedActor.GatherDirectoryResourcesArgsForCall(0)).To(Equal(pwd))
					Expect(expectedPushPlan.AllResources[0]).To(Equal(resources[0].ToV3Resource()))
				})

				It("sets Archive to false", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(expectedPushPlan.Archive).To(BeFalse())
				})
			})

			When("gathering the resources errors", func() {
				BeforeEach(func() {
					fakeSharedActor.GatherDirectoryResourcesReturns(nil, errors.New("kaboom"))
				})

				It("returns the error", func() {
					Expect(executeErr).To(MatchError("kaboom"))
				})
			})
		})

		When("the app resources are given as an archive", func() {
			var archivePath string

			BeforeEach(func() {
				archive, err := ioutil.TempFile("", "push-plan-archive")
				Expect(err).ToNot(HaveOccurred())
				defer archive.Close()

				archivePath = archive.Name()
				pushPlan.BitsPath = archivePath
			})

			AfterEach(func() {
				Expect(os.RemoveAll(archivePath)).ToNot(HaveOccurred())
			})

			When("gathering the resources is successful", func() {
				var resources []sharedaction.Resource

				BeforeEach(func() {
					resources = []sharedaction.Resource{
						{
							Filename: "fake-app-file",
						},
					}
					fakeSharedActor.GatherArchiveResourcesReturns(resources, nil)
				})

				It("adds the gathered resources to the push plan", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(fakeSharedActor.GatherArchiveResourcesCallCount()).To(Equal(1))
					Expect(fakeSharedActor.GatherArchiveResourcesArgsForCall(0)).To(Equal(archivePath))
					Expect(expectedPushPlan.AllResources[0]).To(Equal(resources[0].ToV3Resource()))
				})

				It("sets Archive to true", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(expectedPushPlan.Archive).To(BeTrue())
				})
			})

			When("gathering the resources errors", func() {
				BeforeEach(func() {
					fakeSharedActor.GatherArchiveResourcesReturns(nil, errors.New("kaboom"))
				})

				It("returns the error", func() {
					Expect(executeErr).To(MatchError("kaboom"))
				})
			})
		})
	})
})
