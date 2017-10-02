package pushaction_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/pushaction"
	"code.cloudfoundry.org/cli/actor/pushaction/pushactionfakes"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/types"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Applications", func() {
	var (
		actor       *Actor
		fakeV2Actor *pushactionfakes.FakeV2Actor
	)

	BeforeEach(func() {
		fakeV2Actor = new(pushactionfakes.FakeV2Actor)
		actor = NewActor(fakeV2Actor, nil)
	})

	Describe("CreateOrUpdateApp", func() {
		var (
			config ApplicationConfig

			returnedConfig ApplicationConfig
			event          Event
			warnings       Warnings
			executeErr     error
		)

		BeforeEach(func() {
			config = ApplicationConfig{
				DesiredApplication: Application{
					Application: v2action.Application{
						Name:      "some-app-name",
						SpaceGUID: "some-space-guid",
					},
				},
				Path: "some-path",
			}
		})

		JustBeforeEach(func() {
			returnedConfig, event, warnings, executeErr = actor.CreateOrUpdateApp(config)
		})

		Context("when the app exists", func() {
			BeforeEach(func() {
				config.CurrentApplication = Application{
					Application: v2action.Application{
						Name:      "some-app-name",
						GUID:      "some-app-guid",
						SpaceGUID: "some-space-guid",
						Buildpack: types.FilteredString{Value: "java", IsSet: true},
					},
				}
				config.DesiredApplication = Application{
					Application: v2action.Application{
						Name:      "some-app-name",
						GUID:      "some-app-guid",
						SpaceGUID: "some-space-guid",
						Buildpack: types.FilteredString{Value: "ruby", IsSet: true},
					},
				}
			})

			Context("when the update is successful", func() {
				BeforeEach(func() {
					fakeV2Actor.UpdateApplicationReturns(v2action.Application{
						Name:      "some-app-name",
						GUID:      "some-app-guid",
						SpaceGUID: "some-space-guid",
						Buildpack: types.FilteredString{Value: "ruby", IsSet: true},
					}, v2action.Warnings{"update-warning"}, nil)
				})

				It("updates the application", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("update-warning"))
					Expect(event).To(Equal(UpdatedApplication))

					Expect(returnedConfig.DesiredApplication).To(Equal(Application{
						Application: v2action.Application{
							Name:      "some-app-name",
							GUID:      "some-app-guid",
							SpaceGUID: "some-space-guid",
							Buildpack: types.FilteredString{Value: "ruby", IsSet: true},
						}}))
					Expect(returnedConfig.CurrentApplication).To(Equal(returnedConfig.DesiredApplication))

					Expect(fakeV2Actor.UpdateApplicationCallCount()).To(Equal(1))
					Expect(fakeV2Actor.UpdateApplicationArgsForCall(0)).To(Equal(v2action.Application{
						Name:      "some-app-name",
						GUID:      "some-app-guid",
						SpaceGUID: "some-space-guid",
						Buildpack: types.FilteredString{Value: "ruby", IsSet: true},
					}))
				})

				Context("when the stack guid is not being updated", func() {
					BeforeEach(func() {
						config.CurrentApplication.StackGUID = "some-stack-guid"
						config.DesiredApplication.StackGUID = "some-stack-guid"
					})

					It("does not send the stack guid on update", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(fakeV2Actor.UpdateApplicationCallCount()).To(Equal(1))
						Expect(fakeV2Actor.UpdateApplicationArgsForCall(0)).To(Equal(v2action.Application{
							Name:      "some-app-name",
							GUID:      "some-app-guid",
							SpaceGUID: "some-space-guid",
							Buildpack: types.FilteredString{Value: "ruby", IsSet: true},
						}))
					})
				})
			})

			Context("when the update errors", func() {
				var expectedErr error
				BeforeEach(func() {
					expectedErr = errors.New("oh my")
					fakeV2Actor.UpdateApplicationReturns(v2action.Application{}, v2action.Warnings{"update-warning"}, expectedErr)
				})

				It("returns warnings and error and stops", func() {
					Expect(executeErr).To(MatchError(expectedErr))
					Expect(warnings).To(ConsistOf("update-warning"))
				})
			})
		})

		Context("when the app does not exist", func() {
			Context("when the creation is successful", func() {
				BeforeEach(func() {
					fakeV2Actor.CreateApplicationReturns(v2action.Application{
						Name:      "some-app-name",
						GUID:      "some-app-guid",
						SpaceGUID: "some-space-guid",
						Buildpack: types.FilteredString{Value: "ruby", IsSet: true},
					}, v2action.Warnings{"create-warning"}, nil)
				})

				It("creates the application", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("create-warning"))
					Expect(event).To(Equal(CreatedApplication))

					Expect(returnedConfig.DesiredApplication).To(Equal(Application{
						Application: v2action.Application{
							Name:      "some-app-name",
							GUID:      "some-app-guid",
							SpaceGUID: "some-space-guid",
							Buildpack: types.FilteredString{Value: "ruby", IsSet: true},
						}}))
					Expect(returnedConfig.CurrentApplication).To(Equal(returnedConfig.DesiredApplication))

					Expect(fakeV2Actor.CreateApplicationCallCount()).To(Equal(1))
					Expect(fakeV2Actor.CreateApplicationArgsForCall(0)).To(Equal(v2action.Application{
						Name:      "some-app-name",
						SpaceGUID: "some-space-guid",
					}))
				})
			})

			Context("when the creation errors", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("oh my")
					fakeV2Actor.CreateApplicationReturns(v2action.Application{}, v2action.Warnings{"create-warning"}, expectedErr)
				})

				It("sends the warnings and errors and returns true", func() {
					Expect(executeErr).To(MatchError(expectedErr))
					Expect(warnings).To(ConsistOf("create-warning"))
				})
			})
		})
	})

	Describe("FindOrReturnPartialApp", func() {
		var expectedStack v2action.Stack
		var expectedApp v2action.Application

		Context("when the app exists", func() {
			Context("when retrieving the stack is successful", func() {
				BeforeEach(func() {
					expectedStack = v2action.Stack{
						Name: "some-stack",
						GUID: "some-stack-guid",
					}
					fakeV2Actor.GetStackReturns(expectedStack, v2action.Warnings{"stack-warnings"}, nil)

					expectedApp = v2action.Application{
						GUID:      "some-app-guid",
						Name:      "some-app",
						StackGUID: expectedStack.GUID,
					}
					fakeV2Actor.GetApplicationByNameAndSpaceReturns(expectedApp, v2action.Warnings{"app-warnings"}, nil)
				})

				It("fills in the stack", func() {
					found, app, warnings, err := actor.FindOrReturnPartialApp("some-app", "some-space-guid")
					Expect(err).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("app-warnings", "stack-warnings"))
					Expect(found).To(BeTrue())
					Expect(app).To(Equal(Application{
						Application: expectedApp,
						Stack:       expectedStack,
					}))
				})
			})

			Context("when retrieving the stack errors", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("stack stack stack em up")
					fakeV2Actor.GetStackReturns(v2action.Stack{}, v2action.Warnings{"stack-warnings"}, expectedErr)

					expectedApp = v2action.Application{
						GUID:      "some-app-guid",
						Name:      "some-app",
						StackGUID: "some-stack-guid",
					}
					fakeV2Actor.GetApplicationByNameAndSpaceReturns(expectedApp, v2action.Warnings{"app-warnings"}, nil)
				})

				It("returns error and warnings", func() {
					found, _, warnings, err := actor.FindOrReturnPartialApp("some-app", "some-space-guid")
					Expect(err).To(MatchError(expectedErr))
					Expect(warnings).To(ConsistOf("app-warnings", "stack-warnings"))
					Expect(found).To(BeFalse())
				})
			})
		})

		Context("when the app does not exist", func() {
			BeforeEach(func() {
				fakeV2Actor.GetApplicationByNameAndSpaceReturns(v2action.Application{}, v2action.Warnings{"some-app-warning-1", "some-app-warning-2"}, actionerror.ApplicationNotFoundError{})
			})

			It("returns a partial app and warnings", func() {
				found, app, warnings, err := actor.FindOrReturnPartialApp("some-app", "some-space-guid")
				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("some-app-warning-1", "some-app-warning-2"))
				Expect(found).To(BeFalse())
				Expect(app).To(Equal(Application{
					Application: v2action.Application{
						Name:      "some-app",
						SpaceGUID: "some-space-guid",
					},
				}))
			})
		})

		Context("when retrieving the app errors", func() {
			var expectedErr error
			BeforeEach(func() {
				expectedErr = errors.New("dios mio")
				fakeV2Actor.GetApplicationByNameAndSpaceReturns(v2action.Application{}, v2action.Warnings{"some-app-warning-1", "some-app-warning-2"}, expectedErr)
			})

			It("returns a errors and warnings", func() {
				found, _, warnings, err := actor.FindOrReturnPartialApp("some-app", "some-space-guid")
				Expect(err).To(MatchError(expectedErr))
				Expect(warnings).To(ConsistOf("some-app-warning-1", "some-app-warning-2"))
				Expect(found).To(BeFalse())
			})
		})
	})
})
