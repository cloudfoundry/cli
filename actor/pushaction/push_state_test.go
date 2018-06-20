package pushaction_test

import (
	"errors"
	"os"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/pushaction"
	"code.cloudfoundry.org/cli/actor/pushaction/pushactionfakes"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v3action"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("Push State", func() {
	var (
		actor           *Actor
		fakeV2Actor     *pushactionfakes.FakeV2Actor
		fakeV3Actor     *pushactionfakes.FakeV3Actor
		fakeSharedActor *pushactionfakes.FakeSharedActor

		pwd string
	)

	BeforeEach(func() {
		fakeV2Actor = new(pushactionfakes.FakeV2Actor)
		fakeV3Actor = new(pushactionfakes.FakeV3Actor)
		fakeSharedActor = new(pushactionfakes.FakeSharedActor)
		actor = NewActor(fakeV2Actor, fakeV3Actor, fakeSharedActor)
	})

	Describe("Conceptualize", func() {
		var (
			settings  CommandLineSettings
			spaceGUID string

			states     []PushState
			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			var err error
			pwd, err = os.Getwd()
			Expect(err).ToNot(HaveOccurred())
			settings = CommandLineSettings{
				Name:             "some-app-name",
				CurrentDirectory: pwd,
			}
			spaceGUID = "some-space-guid"
		})

		JustBeforeEach(func() {
			states, warnings, executeErr = actor.Conceptualize(settings, spaceGUID)
		})

		Describe("application", func() {
			Context("when the application exists", func() {
				var app v3action.Application

				BeforeEach(func() {
					app = v3action.Application{
						GUID: "some-app-guid",
						Name: "some-app-name",
					}

					fakeV3Actor.GetApplicationByNameAndSpaceReturns(app, v3action.Warnings{"some-app-warning"}, nil)
				})

				It("uses the found app in the application state", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("some-app-warning"))
					Expect(states).To(HaveLen(1))

					Expect(states[0]).To(MatchFields(IgnoreExtras,
						Fields{
							"Application": Equal(app),
							"SpaceGUID":   Equal(spaceGUID),
						}))

					Expect(fakeV3Actor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
					passedName, passedSpaceGUID := fakeV3Actor.GetApplicationByNameAndSpaceArgsForCall(0)
					Expect(passedName).To(Equal("some-app-name"))
					Expect(passedSpaceGUID).To(Equal(spaceGUID))
				})
			})

			Context("when the application does not exist", func() {
				BeforeEach(func() {
					fakeV3Actor.GetApplicationByNameAndSpaceReturns(v3action.Application{}, v3action.Warnings{"some-app-warning"}, actionerror.ApplicationNotFoundError{})
				})

				It("creates a new app in the application state", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("some-app-warning"))
					Expect(states).To(HaveLen(1))

					Expect(states[0]).To(MatchFields(IgnoreExtras,
						Fields{
							"Application": Equal(v3action.Application{
								Name: "some-app-name",
							}),
							"SpaceGUID": Equal(spaceGUID),
						}))
				})
			})

			Context("when the application lookup errors", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("some-error")
					fakeV3Actor.GetApplicationByNameAndSpaceReturns(v3action.Application{}, v3action.Warnings{"some-app-warning"}, expectedErr)
				})

				It("translates command line settings into a single push state", func() {
					Expect(executeErr).To(MatchError(expectedErr))
					Expect(warnings).To(ConsistOf("some-app-warning"))
				})
			})
		})

		Describe("bits path", func() {
			Context("when no app path is provided in the command line settings", func() {
				It("sets the bits path to the current directory in the settings", func() {
					Expect(states[0].BitsPath).To(Equal(pwd))
				})
			})

			Context("when an app path is provided in the command line settings", func() {
				BeforeEach(func() {
					settings.ProvidedAppPath = "my-app-path"
				})

				It("sets the bits path to the provided app path", func() {
					Expect(states[0].BitsPath).To(Equal("my-app-path"))
				})
			})
		})

		Describe("all resources", func() {
			Context("when the app resources are given as a directory", func() {
				var resources []sharedaction.Resource

				Context("when gathering the resources is successful", func() {

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

				Context("when gathering the resources errors", func() {
					BeforeEach(func() {
						fakeSharedActor.GatherDirectoryResourcesReturns(nil, errors.New("kaboom"))
					})

					It("returns the error", func() {
						Expect(executeErr).To(MatchError("kaboom"))
					})
				})
			})
		})
	})
})
