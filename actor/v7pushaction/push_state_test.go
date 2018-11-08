package v7pushaction_test

import (
	"errors"
	"os"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	. "code.cloudfoundry.org/cli/actor/v7pushaction"
	"code.cloudfoundry.org/cli/actor/v7pushaction/v7pushactionfakes"
	"code.cloudfoundry.org/cli/types"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("Push State", func() {
	var (
		actor           *Actor
		fakeV7Actor     *v7pushactionfakes.FakeV7Actor
		fakeSharedActor *v7pushactionfakes.FakeSharedActor

		pwd string
	)

	BeforeEach(func() {
		actor, _, fakeV7Actor, fakeSharedActor = getTestPushActor()
	})

	Describe("Conceptualize", func() {
		var (
			appName       string
			spaceGUID     string
			orgGUID       string
			currentDir    string
			flagOverrides FlagOverrides

			states     []PushState
			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			var err error
			pwd, err = os.Getwd()
			Expect(err).ToNot(HaveOccurred())
			appName = "some-app-name"
			currentDir = pwd
			flagOverrides = FlagOverrides{}

			spaceGUID = "some-space-guid"
			orgGUID = "some-org-guid"
		})

		JustBeforeEach(func() {
			states, warnings, executeErr = actor.Conceptualize(appName, spaceGUID, orgGUID, currentDir, flagOverrides)
		})

		Describe("application", func() {
			When("the application exists", func() {
				var app v7action.Application

				BeforeEach(func() {
					app = v7action.Application{
						GUID: "some-app-guid",
						Name: "some-app-name",
					}

					fakeV7Actor.GetApplicationByNameAndSpaceReturns(app, v7action.Warnings{"some-app-warning"}, nil)
				})

				It("uses the found app in the application state", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("some-app-warning"))
					Expect(states).To(HaveLen(1))

					Expect(states[0]).To(MatchFields(IgnoreExtras,
						Fields{
							"Application": Equal(app),
							"SpaceGUID":   Equal(spaceGUID),
							"OrgGUID":     Equal(orgGUID),
						}))

					Expect(fakeV7Actor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
					passedName, passedSpaceGUID := fakeV7Actor.GetApplicationByNameAndSpaceArgsForCall(0)
					Expect(passedName).To(Equal("some-app-name"))
					Expect(passedSpaceGUID).To(Equal(spaceGUID))
				})
			})

			When("the application does not exist", func() {
				BeforeEach(func() {
					fakeV7Actor.GetApplicationByNameAndSpaceReturns(v7action.Application{}, v7action.Warnings{"some-app-warning"}, actionerror.ApplicationNotFoundError{})
				})

				It("creates a new app in the application state", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("some-app-warning"))
					Expect(states).To(HaveLen(1))

					Expect(states[0]).To(MatchFields(IgnoreExtras,
						Fields{
							"Application": Equal(v7action.Application{
								Name: "some-app-name",
							}),
							"SpaceGUID": Equal(spaceGUID),
							"OrgGUID":   Equal(orgGUID),
						}))
				})
			})

			When("the application lookup errors", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("some-error")
					fakeV7Actor.GetApplicationByNameAndSpaceReturns(v7action.Application{}, v7action.Warnings{"some-app-warning"}, expectedErr)
				})

				It("translates command line settings into a single push state", func() {
					Expect(executeErr).To(MatchError(expectedErr))
					Expect(warnings).To(ConsistOf("some-app-warning"))
				})
			})
		})

		Describe("bits path", func() {
			When("no app path is provided in the command line settings", func() {
				It("sets the bits path to the current directory in the settings", func() {
					Expect(states[0].BitsPath).To(Equal(pwd))
				})
			})

			When("an app path is provided in the command line settings", func() {
				BeforeEach(func() {
					flagOverrides.ProvidedAppPath = "my-app-path"
				})

				It("sets the bits path to the provided app path", func() {
					Expect(states[0].BitsPath).To(Equal("my-app-path"))
				})
			})
		})

		Describe("all resources", func() {
			When("the app resources are given as a directory", func() {
				var resources []sharedaction.Resource

				When("gathering the resources is successful", func() {

					BeforeEach(func() {
						resources = []sharedaction.Resource{
							{
								Filename: "fake-app-file",
							},
						}
						fakeSharedActor.GatherDirectoryResourcesReturns(resources, nil)
					})

					It("adds the gathered resources to the push state", func() {
						Expect(fakeSharedActor.GatherDirectoryResourcesCallCount()).To(Equal(1))
						Expect(fakeSharedActor.GatherDirectoryResourcesArgsForCall(0)).To(Equal(pwd))
						Expect(states[0].AllResources).To(Equal(resources))
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
		})

		When("flag overrides are passed", func() {
			BeforeEach(func() {
				flagOverrides.Memory = types.NullUint64{IsSet: true, Value: 123456}
			})

			It("sets the all the flag overrides on the state", func() {
				Expect(states[0].Overrides).To(Equal(flagOverrides))
			})
		})
	})
})
