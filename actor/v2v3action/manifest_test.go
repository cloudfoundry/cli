package v2v3action_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	"code.cloudfoundry.org/cli/actor/v2action"
	. "code.cloudfoundry.org/cli/actor/v2v3action"
	"code.cloudfoundry.org/cli/actor/v2v3action/v2v3actionfakes"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/manifest"
)

var _ = Describe("Manifest", func() {
	var (
		actor       *Actor
		fakeV2Actor *v2v3actionfakes.FakeV2Actor
		fakeV3Actor *v2v3actionfakes.FakeV3Actor

		appName  string
		appSpace string

		manifestApp manifest.Application
		warnings    Warnings
		executeErr  error
	)

	BeforeEach(func() {
		fakeV2Actor = new(v2v3actionfakes.FakeV2Actor)
		fakeV3Actor = new(v2v3actionfakes.FakeV3Actor)

		actor = NewActor(fakeV2Actor, fakeV3Actor)

		appName = "some-app-name"
		appSpace = "some-space-GUID"
	})

	JustBeforeEach(func() {
		manifestApp, warnings, executeErr = actor.CreateApplicationManifestByNameAndSpace(appName, appSpace)
	})

	It("calls v2Actor.CreateManifestApplication with the appName and appSpace", func() {
		Expect(fakeV2Actor.CreateApplicationManifestByNameAndSpaceCallCount()).To(Equal(1))
		appNameArg, appSpaceArg := fakeV2Actor.CreateApplicationManifestByNameAndSpaceArgsForCall(0)
		Expect(appNameArg).To(Equal(appName))
		Expect(appSpaceArg).To(Equal(appSpace))
	})

	When("v2Actor.CreateManifestApplication succeeds", func() {
		BeforeEach(func() {
			v2Application := manifest.Application{
				Buildpacks: []string{"some-buildpack"},
			}

			fakeV2Actor.CreateApplicationManifestByNameAndSpaceReturns(v2Application, v2action.Warnings{"v2-action-warnings"}, nil)
		})

		When("the cc returns an invalid semver", func() {
			BeforeEach(func() {
				fakeV3Actor.CloudControllerAPIVersionReturns("i am invalid")
			})

			It("returns a semver error", func() {
				Expect(executeErr).To(MatchError("No Major.Minor.Patch elements found"))
				Expect(warnings).To(ConsistOf("v2-action-warnings"))
			})

		})

		When("the cc has a v3 buildpacks endpoint ( >= v3.25)", func() {
			BeforeEach(func() {
				fakeV3Actor.CloudControllerAPIVersionReturns("3.25.0")
			})

			It("Calls the v3actor.GetApplicationByNameAndSpace with the appName and appSpace", func() {
				Expect(fakeV3Actor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
				appNameArg, appSpaceArg := fakeV3Actor.GetApplicationByNameAndSpaceArgsForCall(0)
				Expect(appNameArg).To(Equal(appName))
				Expect(appSpaceArg).To(Equal(appSpace))
			})

			When("the v3Actor.GetApplicationByNameAndSpace succeeds", func() {
				BeforeEach(func() {
					v3Application := v3action.Application{LifecycleBuildpacks: []string{"some-buildpack"}}
					fakeV3Actor.GetApplicationByNameAndSpaceReturns(v3Application, v3action.Warnings{"v3-action-warnings"}, nil)
				})

				It("should return an application with v3 specific attributes", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("v2-action-warnings", "v3-action-warnings"))
					Expect(manifestApp).To(MatchFields(IgnoreExtras, Fields{
						"Buildpack":  Equal(types.FilteredString{}),
						"Buildpacks": ConsistOf("some-buildpack"),
					}))
				})
			})

			When("the v3Actor.GetApplicationByNameAndSpace fails", func() {
				BeforeEach(func() {
					fakeV3Actor.GetApplicationByNameAndSpaceReturns(v3action.Application{}, v3action.Warnings{"v3-action-warnings"}, errors.New("i'm a v3 error"))
				})

				It("it raises the error", func() {
					Expect(executeErr).To(MatchError("i'm a v3 error"))
					Expect(warnings).To(ConsistOf("v2-action-warnings", "v3-action-warnings"))
				})
			})
		})

		When("the cc does not have a v3 buildpacks endpoint ( < v3.25)", func() {
			BeforeEach(func() {
				fakeV3Actor.CloudControllerAPIVersionReturns("3.24.0")
			})

			It("does not call the v3actor.GetApplicationByNameAndSpace", func() {
				Expect(fakeV3Actor.GetApplicationByNameAndSpaceCallCount()).To(Equal(0))
			})

			It("returns the v2 only version of a manifest application", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("v2-action-warnings"))
				Expect(manifestApp).To(MatchFields(IgnoreExtras, Fields{
					"Buildpack":  Equal(types.FilteredString{}),
					"Buildpacks": ConsistOf("some-buildpack"),
				}))
			})
		})
	})

	When("v2Actor.CreateManifestApplication fails", func() {
		BeforeEach(func() {
			fakeV2Actor.CreateApplicationManifestByNameAndSpaceReturns(manifest.Application{}, v2action.Warnings{"v2-action-warnings"}, errors.New("spaghetti"))
		})

		It("returns warnings and the error", func() {
			Expect(executeErr).To(MatchError(errors.New("spaghetti")))
			Expect(warnings).To(ConsistOf("v2-action-warnings"))
		})
	})
})
