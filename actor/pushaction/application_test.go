package pushaction_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/pushaction"
	"code.cloudfoundry.org/cli/actor/pushaction/pushactionfakes"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/types"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("Applications", func() {
	var (
		actor       *Actor
		fakeV2Actor *pushactionfakes.FakeV2Actor
		fakeV3Actor *pushactionfakes.FakeV3Actor
	)

	BeforeEach(func() {
		fakeV2Actor = new(pushactionfakes.FakeV2Actor)
		fakeV3Actor = new(pushactionfakes.FakeV3Actor)
		actor = NewActor(fakeV2Actor, fakeV3Actor, nil)
	})

	Describe("UpdateApplication", func() {
		var (
			config ApplicationConfig

			returnedConfig ApplicationConfig
			event          Event
			warnings       Warnings
			executeErr     error

			updatedApplication v2action.Application
		)

		BeforeEach(func() {
			config = ApplicationConfig{
				DesiredApplication: Application{
					Application: v2action.Application{
						Name:      "some-app-name",
						GUID:      "some-app-guid",
						SpaceGUID: "some-space-guid",
					},
				},
				CurrentApplication: Application{
					Application: v2action.Application{
						Name:      "some-app-name",
						GUID:      "some-app-guid",
						SpaceGUID: "some-space-guid",
					},
				},
				Path: "some-path",
			}
		})

		JustBeforeEach(func() {
			returnedConfig, event, warnings, executeErr = actor.UpdateApplication(config)
		})

		Context("when the update is successful", func() {
			BeforeEach(func() {
				updatedApplication = v2action.Application{
					Name:      "some-app-name",
					GUID:      "some-app-guid",
					SpaceGUID: "some-space-guid",
					Buildpack: types.FilteredString{Value: "ruby", IsSet: true},
				}
				fakeV2Actor.UpdateApplicationReturns(updatedApplication, v2action.Warnings{"update-warning"}, nil)
			})

			It("updates the application", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("update-warning"))
				Expect(event).To(Equal(UpdatedApplication))

				Expect(returnedConfig.DesiredApplication.Application).To(Equal(updatedApplication))
				Expect(returnedConfig.CurrentApplication).To(Equal(returnedConfig.DesiredApplication))

				Expect(fakeV2Actor.UpdateApplicationCallCount()).To(Equal(1))
				submitApp := fakeV2Actor.UpdateApplicationArgsForCall(0)
				Expect(submitApp).To(MatchFields(IgnoreExtras, Fields{
					"Name":      Equal("some-app-name"),
					"GUID":      Equal("some-app-guid"),
					"SpaceGUID": Equal("some-space-guid"),
				}))
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

		Context("State", func() {
			Context("when the state is not being updated", func() {
				BeforeEach(func() {
					config.CurrentApplication.State = "some-state"
					config.DesiredApplication.State = "some-state"
				})

				It("does not send the state on update", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(fakeV2Actor.UpdateApplicationCallCount()).To(Equal(1))
					Expect(fakeV2Actor.UpdateApplicationArgsForCall(0)).To(MatchFields(IgnoreExtras, Fields{
						"Name":      Equal("some-app-name"),
						"GUID":      Equal("some-app-guid"),
						"SpaceGUID": Equal("some-space-guid"),
					}))
				})
			})
		})

		Context("StackGUID", func() {
			Context("when the stack guid is not being updated", func() {
				BeforeEach(func() {
					config.CurrentApplication.StackGUID = "some-stack-guid"
					config.DesiredApplication.StackGUID = "some-stack-guid"
				})

				It("does not send the stack guid on update", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(fakeV2Actor.UpdateApplicationCallCount()).To(Equal(1))
					Expect(fakeV2Actor.UpdateApplicationArgsForCall(0)).To(MatchFields(IgnoreExtras, Fields{
						"Name":      Equal("some-app-name"),
						"GUID":      Equal("some-app-guid"),
						"SpaceGUID": Equal("some-space-guid"),
					}))
				})
			})
		})

		Context("Buildpack(s)", func() {
			var (
				buildpack  types.FilteredString
				buildpacks []types.FilteredString
			)

			BeforeEach(func() {
				buildpack = types.FilteredString{}
				buildpacks = nil
			})

			Context("when buildpack is set", func() {
				BeforeEach(func() {
					buildpack = types.FilteredString{Value: "ruby", IsSet: true}
					config.DesiredApplication.Buildpack = buildpack

					updatedApplication = v2action.Application{Buildpack: buildpack}
					fakeV2Actor.UpdateApplicationReturns(updatedApplication, v2action.Warnings{"update-warning"}, nil)
				})

				It("keeps buildpack in the desired application", func() {
					Expect(fakeV2Actor.UpdateApplicationCallCount()).To(Equal(1))
					submitApp := fakeV2Actor.UpdateApplicationArgsForCall(0)
					Expect(submitApp).To(MatchFields(IgnoreExtras, Fields{
						"Buildpack": Equal(buildpack),
					}))

					Expect(fakeV3Actor.UpdateApplicationCallCount()).To(Equal(0))
					Expect(returnedConfig.DesiredApplication.Application).To(Equal(updatedApplication))
				})
			})

			Context("when buildpacks is set with one buildpack", func() {
				BeforeEach(func() {
					buildpacks = []types.FilteredString{
						{Value: "ruby", IsSet: true},
					}
					config.DesiredApplication.Buildpacks = buildpacks

					updatedApplication = v2action.Application{Buildpack: buildpacks[0]}
					fakeV2Actor.UpdateApplicationReturns(updatedApplication, v2action.Warnings{"update-warning"}, nil)
				})

				It("sets buildpack to the set buildpack in buildpacks", func() {
					Expect(fakeV2Actor.UpdateApplicationCallCount()).To(Equal(1))
					submitApp := fakeV2Actor.UpdateApplicationArgsForCall(0)
					Expect(submitApp).To(MatchFields(IgnoreExtras, Fields{
						"Buildpack": Equal(buildpacks[0]),
					}))

					Expect(fakeV3Actor.UpdateApplicationCallCount()).To(Equal(0))
					Expect(returnedConfig.DesiredApplication.Application).To(Equal(updatedApplication))
				})
			})

			Context("when buildpacks is set with more than one buildpack", func() {
				BeforeEach(func() {
					buildpacks = []types.FilteredString{
						{Value: "ruby", IsSet: true},
						{Value: "java", IsSet: true},
					}
					config.DesiredApplication.Buildpacks = buildpacks

					updatedApplication = v2action.Application{}
					fakeV2Actor.UpdateApplicationReturns(updatedApplication, v2action.Warnings{"update-warning"}, nil)
				})

				It("does not set buildpack", func() {
					Expect(fakeV2Actor.UpdateApplicationCallCount()).To(Equal(1))
					submitApp := fakeV2Actor.UpdateApplicationArgsForCall(0)
					Expect(submitApp).To(MatchFields(IgnoreExtras, Fields{
						"Buildpack": Equal(types.FilteredString{}),
					}))

					Expect(returnedConfig.DesiredApplication.Application).To(Equal(updatedApplication))
				})

				Context("when the v3 update is successful", func() {
					var submitApp v3action.Application

					BeforeEach(func() {
						updatedApplication = config.DesiredApplication.Application
						updatedApplication.GUID = "yay-im-a-guid"
						submitApp = v3action.Application{
							Name:                updatedApplication.Name,
							GUID:                updatedApplication.GUID,
							LifecycleBuildpacks: []string{"ruby", "java"},
						}

						fakeV2Actor.UpdateApplicationReturns(updatedApplication, v2action.Warnings{"v2-create-application-warnings"}, nil)
						fakeV3Actor.UpdateApplicationReturns(v3action.Application{}, v3action.Warnings{"v3-update-application-warnings"}, nil)
					})

					It("updates only the buildpacks in ccv3", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(warnings).To(ConsistOf("v2-create-application-warnings", "v3-update-application-warnings"))

						Expect(fakeV3Actor.UpdateApplicationCallCount()).To(Equal(1))
						Expect(fakeV3Actor.UpdateApplicationArgsForCall(0)).To(Equal(submitApp))

						Expect(returnedConfig.DesiredApplication.Application).To(Equal(updatedApplication))
					})
				})

				Context("when the v3 update fails", func() {
					BeforeEach(func() {
						fakeV2Actor.UpdateApplicationReturns(v2action.Application{}, v2action.Warnings{"v2-create-application-warnings"}, nil)
						fakeV3Actor.UpdateApplicationReturns(v3action.Application{}, v3action.Warnings{"v3-update-application-warnings"}, errors.New("boom"))
					})

					It("raises an error", func() {
						Expect(executeErr).To(MatchError("boom"))
						Expect(warnings).To(ConsistOf("v2-create-application-warnings", "v3-update-application-warnings"))

						Expect(fakeV3Actor.UpdateApplicationCallCount()).To(Equal(1))
					})
				})
			})
		})
	})

	Describe("CreateApplication", func() {
		var (
			config ApplicationConfig

			returnedConfig ApplicationConfig
			event          Event
			warnings       Warnings
			executeErr     error

			createdApplication v2action.Application
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
			returnedConfig, event, warnings, executeErr = actor.CreateApplication(config)
		})

		Context("when the creation is successful", func() {
			BeforeEach(func() {
				createdApplication = v2action.Application{
					Name:      "some-app-name",
					GUID:      "some-app-guid",
					SpaceGUID: "some-space-guid",
				}

				fakeV2Actor.CreateApplicationReturns(createdApplication, v2action.Warnings{"create-warning"}, nil)
			})

			It("creates the application", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("create-warning"))
				Expect(event).To(Equal(CreatedApplication))

				Expect(returnedConfig.DesiredApplication.Application).To(Equal(createdApplication))
				Expect(returnedConfig.CurrentApplication).To(Equal(returnedConfig.DesiredApplication))

				Expect(fakeV2Actor.CreateApplicationCallCount()).To(Equal(1))
				submitApp := fakeV2Actor.CreateApplicationArgsForCall(0)
				Expect(submitApp).To(MatchFields(IgnoreExtras, Fields{
					"Name":      Equal("some-app-name"),
					"SpaceGUID": Equal("some-space-guid"),
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

		Context("Buildpack(s)", func() {
			var (
				buildpack  types.FilteredString
				buildpacks []types.FilteredString
			)

			Context("when buildpack is set", func() {
				BeforeEach(func() {
					buildpack = types.FilteredString{Value: "ruby", IsSet: true}
					config.DesiredApplication.Buildpack = buildpack

					createdApplication = v2action.Application{Buildpack: buildpack}
					fakeV2Actor.CreateApplicationReturns(createdApplication, v2action.Warnings{"create-warning"}, nil)
				})

				It("keeps buildpack in the desired application", func() {
					Expect(fakeV2Actor.CreateApplicationCallCount()).To(Equal(1))
					submitApp := fakeV2Actor.CreateApplicationArgsForCall(0)
					Expect(submitApp).To(MatchFields(IgnoreExtras, Fields{
						"Buildpack": Equal(buildpack),
					}))

					Expect(fakeV3Actor.UpdateApplicationCallCount()).To(Equal(0))
					Expect(returnedConfig.DesiredApplication.Application).To(Equal(createdApplication))
				})
			})

			Context("when buildpacks is set with one buildpack", func() {
				BeforeEach(func() {
					buildpacks = []types.FilteredString{
						{Value: "ruby", IsSet: true},
					}
					config.DesiredApplication.Buildpacks = buildpacks

					createdApplication = v2action.Application{Buildpack: buildpacks[0]}
					fakeV2Actor.CreateApplicationReturns(createdApplication, v2action.Warnings{"create-warning"}, nil)
				})

				It("sets buildpack to the set buildpack in buildpacks", func() {
					Expect(fakeV2Actor.CreateApplicationCallCount()).To(Equal(1))
					submitApp := fakeV2Actor.CreateApplicationArgsForCall(0)
					Expect(submitApp).To(MatchFields(IgnoreExtras, Fields{
						"Buildpack": Equal(buildpacks[0]),
					}))

					Expect(fakeV3Actor.UpdateApplicationCallCount()).To(Equal(0))
					Expect(returnedConfig.DesiredApplication.Application).To(Equal(createdApplication))
				})
			})

			Context("when buildpacks is set with more than one buildpack", func() {
				BeforeEach(func() {
					buildpacks = []types.FilteredString{
						{Value: "ruby", IsSet: true},
						{Value: "java", IsSet: true},
					}
					config.DesiredApplication.Buildpacks = buildpacks

					createdApplication = v2action.Application{}
					fakeV2Actor.CreateApplicationReturns(createdApplication, v2action.Warnings{"create-warning"}, nil)
				})

				It("does not set buildpack", func() {
					Expect(fakeV2Actor.CreateApplicationCallCount()).To(Equal(1))
					submitApp := fakeV2Actor.CreateApplicationArgsForCall(0)
					Expect(submitApp).To(MatchFields(IgnoreExtras, Fields{
						"Buildpack": Equal(types.FilteredString{}),
					}))

					Expect(returnedConfig.DesiredApplication.Application).To(Equal(createdApplication))
				})

				Context("when the v3 update is successful", func() {
					var submitApp v3action.Application

					BeforeEach(func() {
						createdApplication = config.DesiredApplication.Application
						createdApplication.GUID = "yay-im-a-guid"
						submitApp = v3action.Application{
							Name:                createdApplication.Name,
							GUID:                createdApplication.GUID,
							LifecycleBuildpacks: []string{"ruby", "java"},
						}

						fakeV2Actor.CreateApplicationReturns(createdApplication, v2action.Warnings{"v2-create-application-warnings"}, nil)
						fakeV3Actor.UpdateApplicationReturns(v3action.Application{}, v3action.Warnings{"v3-update-application-warnings"}, nil)
					})

					It("updates only the buildpacks in ccv3", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(warnings).To(ConsistOf("v2-create-application-warnings", "v3-update-application-warnings"))

						Expect(fakeV3Actor.UpdateApplicationCallCount()).To(Equal(1))
						Expect(fakeV3Actor.UpdateApplicationArgsForCall(0)).To(Equal(submitApp))

						Expect(returnedConfig.DesiredApplication.Application).To(Equal(createdApplication))
					})
				})

				Context("when the v3 update fails", func() {
					BeforeEach(func() {
						fakeV2Actor.CreateApplicationReturns(v2action.Application{}, v2action.Warnings{"v2-create-application-warnings"}, nil)
						fakeV3Actor.UpdateApplicationReturns(v3action.Application{}, v3action.Warnings{"v3-update-application-warnings"}, errors.New("boom"))
					})

					It("raises an error", func() {
						Expect(executeErr).To(MatchError("boom"))
						Expect(warnings).To(ConsistOf("v2-create-application-warnings", "v3-update-application-warnings"))

						Expect(fakeV3Actor.UpdateApplicationCallCount()).To(Equal(1))
					})
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
