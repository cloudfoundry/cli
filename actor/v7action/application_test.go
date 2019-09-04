package v7action_test

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/clock"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Application Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
		fakeConfig                *v7actionfakes.FakeConfig
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v7actionfakes.FakeCloudControllerClient)
		fakeConfig = new(v7actionfakes.FakeConfig)
		actor = NewActor(fakeCloudControllerClient, fakeConfig, nil, nil, clock.NewClock())
	})

	Describe("DeleteApplicationByNameAndSpace", func() {
		var (
			warnings           Warnings
			executeErr         error
			deleteMappedRoutes bool
			appName            string
		)

		JustBeforeEach(func() {
			appName = "some-app"
			warnings, executeErr = actor.DeleteApplicationByNameAndSpace(appName, "some-space-guid", deleteMappedRoutes)
		})

		When("looking up the app guid fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns([]ccv3.Application{}, ccv3.Warnings{"some-get-app-warning"}, errors.New("some-get-app-error"))
			})

			It("returns the warnings and error", func() {
				Expect(warnings).To(ConsistOf("some-get-app-warning"))
				Expect(executeErr).To(MatchError("some-get-app-error"))
			})
		})

		When("looking up the app guid succeeds", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns([]ccv3.Application{{Name: "some-app", GUID: "abc123"}}, ccv3.Warnings{"some-get-app-warning"}, nil)
			})

			When("sending the delete fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.DeleteApplicationReturns("", ccv3.Warnings{"some-delete-app-warning"}, errors.New("some-delete-app-error"))
				})

				It("returns the warnings and error", func() {
					Expect(warnings).To(ConsistOf("some-get-app-warning", "some-delete-app-warning"))
					Expect(executeErr).To(MatchError("some-delete-app-error"))
				})
			})

			When("sending the delete succeeds", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.DeleteApplicationReturns("/some-job-url", ccv3.Warnings{"some-delete-app-warning"}, nil)
				})

				When("polling fails", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.PollJobReturns(ccv3.Warnings{"some-poll-warning"}, errors.New("some-poll-error"))
					})

					It("returns the warnings and poll error", func() {
						Expect(warnings).To(ConsistOf("some-get-app-warning", "some-delete-app-warning", "some-poll-warning"))
						Expect(executeErr).To(MatchError("some-poll-error"))
					})
				})

				When("polling succeeds", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.PollJobReturns(ccv3.Warnings{"some-poll-warning"}, nil)
					})

					It("returns all the warnings and no error", func() {
						Expect(warnings).To(ConsistOf("some-get-app-warning", "some-delete-app-warning", "some-poll-warning"))
						Expect(executeErr).ToNot(HaveOccurred())
					})
				})
			})
		})

		When("attempting to delete mapped routes", func() {
			BeforeEach(func() {
				deleteMappedRoutes = true
				fakeCloudControllerClient.GetApplicationsReturns([]ccv3.Application{{Name: "some-app", GUID: "abc123"}}, nil, nil)
			})

			When("getting the routes fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetApplicationRoutesReturns(nil, ccv3.Warnings{"get-routes-warning"}, errors.New("get-routes-error"))
				})

				It("returns the warnings and an error", func() {
					Expect(warnings).To(ConsistOf("get-routes-warning"))
					Expect(executeErr).To(MatchError("get-routes-error"))
				})
			})

			When("getting the routes succeeds", func() {
				When("there are no routes", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.GetApplicationRoutesReturns([]ccv3.Route{}, nil, nil)
					})

					It("does not delete any routes", func() {
						Expect(fakeCloudControllerClient.DeleteRouteCallCount()).To(Equal(0))
					})
				})

				When("there are routes", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.GetApplicationRoutesReturns([]ccv3.Route{{GUID: "route-1-guid"}, {GUID: "route-2-guid", URL: "route-2.example.com"}}, nil, nil)
					})

					When("getting route destinations fails", func() {
						BeforeEach(func() {
							fakeCloudControllerClient.GetRouteDestinationsReturnsOnCall(0, nil, ccv3.Warnings{"get-route-destinations-warning"}, errors.New("get-route-destinations-error"))
						})

						It("returns all warnings and an error", func() {
							Expect(warnings).To(ConsistOf("get-route-destinations-warning"))
							Expect(executeErr).To(MatchError("get-route-destinations-error"))
						})
					})

					When("getting route destinations succeeds", func() {
						It("deletes the routes", func() {
							Expect(fakeCloudControllerClient.GetApplicationRoutesCallCount()).To(Equal(1))
							Expect(fakeCloudControllerClient.GetApplicationRoutesArgsForCall(0)).To(Equal("abc123"))
							Expect(fakeCloudControllerClient.DeleteRouteCallCount()).To(Equal(2))
							guids := []string{fakeCloudControllerClient.DeleteRouteArgsForCall(0), fakeCloudControllerClient.DeleteRouteArgsForCall(1)}
							Expect(guids).To(ConsistOf("route-1-guid", "route-2-guid"))
						})

						When("the route has already been deleted", func() {
							BeforeEach(func() {
								fakeCloudControllerClient.DeleteRouteReturnsOnCall(0,
									"",
									ccv3.Warnings{"delete-route-1-warning"},
									ccerror.ResourceNotFoundError{},
								)
								fakeCloudControllerClient.DeleteRouteReturnsOnCall(1,
									"poll-job-url",
									ccv3.Warnings{"delete-route-2-warning"},
									nil,
								)
								fakeCloudControllerClient.PollJobReturnsOnCall(1, ccv3.Warnings{"poll-job-warning"}, nil)
							})

							It("does **not** fail", func() {
								Expect(executeErr).ToNot(HaveOccurred())
								Expect(warnings).To(ConsistOf("delete-route-1-warning", "delete-route-2-warning", "poll-job-warning"))
								Expect(fakeCloudControllerClient.DeleteRouteCallCount()).To(Equal(2))
								Expect(fakeCloudControllerClient.PollJobCallCount()).To(Equal(2))
								Expect(fakeCloudControllerClient.PollJobArgsForCall(1)).To(BeEquivalentTo("poll-job-url"))
							})
						})

						When("app to delete has a route bound to another app", func() {
							BeforeEach(func() {
								fakeCloudControllerClient.GetRouteDestinationsReturnsOnCall(1,
									[]ccv3.RouteDestination{
										{App: ccv3.RouteDestinationApp{GUID: "abc123"}},
										{App: ccv3.RouteDestinationApp{GUID: "different-app-guid"}},
									},
									ccv3.Warnings{"get-destination-warning"},
									nil)
							})
							It("refuses the entire operation", func() {
								Expect(executeErr).To(MatchError(actionerror.RouteBoundToMultipleAppsError{AppName: "some-app", RouteURL: "route-2.example.com"}))
								Expect(warnings).To(ConsistOf("get-destination-warning"))
								Expect(fakeCloudControllerClient.GetRouteDestinationsCallCount()).To(Equal(2))
								Expect(fakeCloudControllerClient.DeleteApplicationCallCount()).To(Equal(0))
								Expect(fakeCloudControllerClient.DeleteRouteCallCount()).To(Equal(0))
							})
						})

						When("deleting the route fails", func() {
							BeforeEach(func() {
								fakeCloudControllerClient.DeleteRouteReturnsOnCall(0,
									"poll-job-url",
									ccv3.Warnings{"delete-route-1-warning"},
									nil,
								)
								fakeCloudControllerClient.DeleteRouteReturnsOnCall(1,
									"",
									ccv3.Warnings{"delete-route-2-warning"},
									errors.New("delete-route-2-error"),
								)
								fakeCloudControllerClient.PollJobReturnsOnCall(1, ccv3.Warnings{"poll-job-warning"}, nil)
							})

							It("returns the error", func() {
								Expect(executeErr).To(MatchError("delete-route-2-error"))
								Expect(warnings).To(ConsistOf("delete-route-1-warning", "delete-route-2-warning", "poll-job-warning"))
							})
						})

					})

					When("the polling job fails", func() {
						BeforeEach(func() {
							fakeCloudControllerClient.PollJobReturns(ccv3.Warnings{"poll-job-warning"}, errors.New("poll-job-error"))
						})

						It("returns the error", func() {
							Expect(executeErr).To(MatchError("poll-job-error"))
						})
					})

				})
			})
		})
	})

	Describe("GetApplicationsByGUIDs", func() {
		When("all of the requested apps exist", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv3.Application{
						{
							Name: "some-app-name",
							GUID: "some-app-guid",
						},
						{
							Name: "other-app-name",
							GUID: "other-app-guid",
						},
					},
					ccv3.Warnings{"some-warning"},
					nil,
				)
			})

			It("returns the applications and warnings", func() {
				apps, warnings, err := actor.GetApplicationsByGUIDs([]string{"some-app-guid", "other-app-guid"})
				Expect(err).ToNot(HaveOccurred())
				Expect(apps).To(ConsistOf(
					Application{
						Name: "some-app-name",
						GUID: "some-app-guid",
					},
					Application{
						Name: "other-app-name",
						GUID: "other-app-guid",
					},
				))
				Expect(warnings).To(ConsistOf("some-warning"))

				Expect(fakeCloudControllerClient.GetApplicationsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetApplicationsArgsForCall(0)).To(ConsistOf(
					ccv3.Query{Key: ccv3.GUIDFilter, Values: []string{"some-app-guid", "other-app-guid"}},
				))
			})
		})

		When("at least one of the requested apps does not exist", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv3.Application{
						{
							Name: "some-app-name",
							GUID: "some-app-guid",
						},
					},
					ccv3.Warnings{"some-warning"},
					nil,
				)
			})

			It("returns an ApplicationNotFoundError and the warnings", func() {
				_, warnings, err := actor.GetApplicationsByGUIDs([]string{"some-app-guid", "non-existent-app-guid"})
				Expect(warnings).To(ConsistOf("some-warning"))
				Expect(err).To(MatchError(actionerror.ApplicationsNotFoundError{}))
			})
		})

		When("a single app has two routes", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv3.Application{
						{
							Name: "some-app-name",
							GUID: "some-app-guid",
						},
					},
					ccv3.Warnings{"some-warning"},
					nil,
				)
			})

			It("returns an ApplicationNotFoundError and the warnings", func() {
				_, warnings, err := actor.GetApplicationsByGUIDs([]string{"some-app-guid", "some-app-guid"})
				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("some-warning"))
			})
		})

		When("the cloud controller client returns an error", func() {
			var expectedError error

			BeforeEach(func() {
				expectedError = errors.New("I am a CloudControllerClient Error")
				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv3.Application{},
					ccv3.Warnings{"some-warning"},
					expectedError)
			})

			It("returns the warnings and the error", func() {
				_, warnings, err := actor.GetApplicationsByGUIDs([]string{"some-app-guid"})
				Expect(warnings).To(ConsistOf("some-warning"))
				Expect(err).To(MatchError(expectedError))
			})
		})
	})

	Describe("GetApplicationsByNameAndSpace", func() {
		When("all of the requested apps exist", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv3.Application{
						{
							Name: "some-app-name",
							GUID: "some-app-guid",
						},
						{
							Name: "other-app-name",
							GUID: "other-app-guid",
						},
					},
					ccv3.Warnings{"some-warning"},
					nil,
				)
			})

			It("returns the applications and warnings", func() {
				apps, warnings, err := actor.GetApplicationsByNamesAndSpace([]string{"some-app-name", "other-app-name"}, "some-space-guid")
				Expect(err).ToNot(HaveOccurred())
				Expect(apps).To(ConsistOf(
					Application{
						Name: "some-app-name",
						GUID: "some-app-guid",
					},
					Application{
						Name: "other-app-name",
						GUID: "other-app-guid",
					},
				))
				Expect(warnings).To(ConsistOf("some-warning"))

				Expect(fakeCloudControllerClient.GetApplicationsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetApplicationsArgsForCall(0)).To(ConsistOf(
					ccv3.Query{Key: ccv3.NameFilter, Values: []string{"some-app-name", "other-app-name"}},
					ccv3.Query{Key: ccv3.SpaceGUIDFilter, Values: []string{"some-space-guid"}},
				))
			})
		})

		When("at least one of the requested apps does not exist", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv3.Application{
						{
							Name: "some-app-name",
						},
					},
					ccv3.Warnings{"some-warning"},
					nil,
				)
			})

			It("returns an ApplicationNotFoundError and the warnings", func() {
				_, warnings, err := actor.GetApplicationsByNamesAndSpace([]string{"some-app-name", "other-app-name"}, "some-space-guid")
				Expect(warnings).To(ConsistOf("some-warning"))
				Expect(err).To(MatchError(actionerror.ApplicationsNotFoundError{}))
			})
		})

		When("a given app has two routes", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv3.Application{
						{
							Name: "some-app-name",
						},
					},
					ccv3.Warnings{"some-warning"},
					nil,
				)
			})

			It("returns an ApplicationNotFoundError and the warnings", func() {
				_, warnings, err := actor.GetApplicationsByNamesAndSpace([]string{"some-app-name", "some-app-name"}, "some-space-guid")
				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("some-warning"))
			})
		})

		When("the cloud controller client returns an error", func() {
			var expectedError error

			BeforeEach(func() {
				expectedError = errors.New("I am a CloudControllerClient Error")
				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv3.Application{},
					ccv3.Warnings{"some-warning"},
					expectedError)
			})

			It("returns the warnings and the error", func() {
				_, warnings, err := actor.GetApplicationsByNamesAndSpace([]string{"some-app-name"}, "some-space-guid")
				Expect(warnings).To(ConsistOf("some-warning"))
				Expect(err).To(MatchError(expectedError))
			})
		})
	})

	Describe("GetApplicationByNameAndSpace", func() {
		When("the app exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv3.Application{
						{
							Name: "some-app-name",
							GUID: "some-app-guid",
							Metadata: &ccv3.Metadata{
								Labels: map[string]types.NullString{
									"some-key": types.NewNullString("some-value"),
								},
							},
						},
					},
					ccv3.Warnings{"some-warning"},
					nil,
				)
			})

			It("returns the application and warnings", func() {
				app, warnings, err := actor.GetApplicationByNameAndSpace("some-app-name", "some-space-guid")
				Expect(err).ToNot(HaveOccurred())
				Expect(app).To(Equal(Application{
					Name: "some-app-name",
					GUID: "some-app-guid",
					Metadata: &Metadata{
						Labels: map[string]types.NullString{"some-key": types.NewNullString("some-value")},
					},
				}))
				Expect(warnings).To(ConsistOf("some-warning"))

				Expect(fakeCloudControllerClient.GetApplicationsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetApplicationsArgsForCall(0)).To(ConsistOf(
					ccv3.Query{Key: ccv3.NameFilter, Values: []string{"some-app-name"}},
					ccv3.Query{Key: ccv3.SpaceGUIDFilter, Values: []string{"some-space-guid"}},
				))
			})
		})

		When("the cloud controller client returns an error", func() {
			var expectedError error

			BeforeEach(func() {
				expectedError = errors.New("I am a CloudControllerClient Error")
				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv3.Application{},
					ccv3.Warnings{"some-warning"},
					expectedError)
			})

			It("returns the warnings and the error", func() {
				_, warnings, err := actor.GetApplicationByNameAndSpace("some-app-name", "some-space-guid")
				Expect(warnings).To(ConsistOf("some-warning"))
				Expect(err).To(MatchError(expectedError))
			})
		})

		When("the app does not exist", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv3.Application{},
					ccv3.Warnings{"some-warning"},
					nil,
				)
			})

			It("returns an ApplicationNotFoundError and the warnings", func() {
				_, warnings, err := actor.GetApplicationByNameAndSpace("some-app-name", "some-space-guid")
				Expect(warnings).To(ConsistOf("some-warning"))
				Expect(err).To(MatchError(actionerror.ApplicationNotFoundError{Name: "some-app-name"}))
			})
		})
	})

	Describe("GetApplicationsBySpace", func() {
		When("the there are applications in the space", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv3.Application{
						{
							GUID: "some-app-guid-1",
							Name: "some-app-1",
						},
						{
							GUID: "some-app-guid-2",
							Name: "some-app-2",
						},
					},
					ccv3.Warnings{"warning-1", "warning-2"},
					nil,
				)
			})

			It("returns the application and warnings", func() {
				apps, warnings, err := actor.GetApplicationsBySpace("some-space-guid")
				Expect(err).ToNot(HaveOccurred())
				Expect(apps).To(ConsistOf(
					Application{
						GUID: "some-app-guid-1",
						Name: "some-app-1",
					},
					Application{
						GUID: "some-app-guid-2",
						Name: "some-app-2",
					},
				))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))

				Expect(fakeCloudControllerClient.GetApplicationsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetApplicationsArgsForCall(0)).To(ConsistOf(
					ccv3.Query{Key: ccv3.SpaceGUIDFilter, Values: []string{"some-space-guid"}},
				))
			})
		})

		When("the cloud controller client returns an error", func() {
			var expectedError error

			BeforeEach(func() {
				expectedError = errors.New("I am a CloudControllerClient Error")
				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv3.Application{},
					ccv3.Warnings{"some-warning"},
					expectedError)
			})

			It("returns the error and warnings", func() {
				_, warnings, err := actor.GetApplicationsBySpace("some-space-guid")
				Expect(warnings).To(ConsistOf("some-warning"))
				Expect(err).To(MatchError(expectedError))
			})
		})
	})

	Describe("CreateApplicationInSpace", func() {
		var (
			application Application
			warnings    Warnings
			err         error
		)

		JustBeforeEach(func() {
			application, warnings, err = actor.CreateApplicationInSpace(Application{
				Name:                "some-app-name",
				LifecycleType:       constant.AppLifecycleTypeBuildpack,
				LifecycleBuildpacks: []string{"buildpack-1", "buildpack-2"},
			}, "some-space-guid")
		})

		When("the app successfully gets created", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CreateApplicationReturns(
					ccv3.Application{
						Name:                "some-app-name",
						GUID:                "some-app-guid",
						LifecycleType:       constant.AppLifecycleTypeBuildpack,
						LifecycleBuildpacks: []string{"buildpack-1", "buildpack-2"},
					},
					ccv3.Warnings{"some-warning"},
					nil,
				)
			})

			It("creates and returns the application and warnings", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(application).To(Equal(Application{
					Name:                "some-app-name",
					GUID:                "some-app-guid",
					LifecycleType:       constant.AppLifecycleTypeBuildpack,
					LifecycleBuildpacks: []string{"buildpack-1", "buildpack-2"},
				}))
				Expect(warnings).To(ConsistOf("some-warning"))

				Expect(fakeCloudControllerClient.CreateApplicationCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.CreateApplicationArgsForCall(0)).To(Equal(ccv3.Application{
					Name: "some-app-name",
					Relationships: ccv3.Relationships{
						constant.RelationshipTypeSpace: ccv3.Relationship{GUID: "some-space-guid"},
					},
					LifecycleType:       constant.AppLifecycleTypeBuildpack,
					LifecycleBuildpacks: []string{"buildpack-1", "buildpack-2"},
				}))
			})
		})

		When("the cc client returns an error", func() {
			var expectedError error

			BeforeEach(func() {
				expectedError = errors.New("I am a CloudControllerClient Error")
				fakeCloudControllerClient.CreateApplicationReturns(
					ccv3.Application{},
					ccv3.Warnings{"some-warning"},
					expectedError,
				)
			})

			It("raises the error and warnings", func() {
				Expect(err).To(MatchError(expectedError))
				Expect(warnings).To(ConsistOf("some-warning"))
			})
		})

		When("the cc client returns an NameNotUniqueInSpaceError", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CreateApplicationReturns(
					ccv3.Application{},
					ccv3.Warnings{"some-warning"},
					ccerror.NameNotUniqueInSpaceError{},
				)
			})

			It("returns the ApplicationAlreadyExistsError and warnings", func() {
				Expect(err).To(MatchError(actionerror.ApplicationAlreadyExistsError{Name: "some-app-name"}))
				Expect(warnings).To(ConsistOf("some-warning"))
			})
		})
	})

	Describe("UpdateApplication", func() {
		var (
			submitApp, resultApp Application
			warnings             Warnings
			err                  error
		)

		JustBeforeEach(func() {
			submitApp = Application{
				GUID:                "some-app-guid",
				StackName:           "some-stack-name",
				LifecycleType:       constant.AppLifecycleTypeBuildpack,
				LifecycleBuildpacks: []string{"buildpack-1", "buildpack-2"},
				Metadata: &Metadata{Labels: map[string]types.NullString{
					"some-label":  types.NewNullString("some-value"),
					"other-label": types.NewNullString("other-value"),
				}},
			}

			resultApp, warnings, err = actor.UpdateApplication(submitApp)
		})

		When("the app successfully gets updated", func() {
			var apiResponseApp ccv3.Application

			BeforeEach(func() {
				apiResponseApp = ccv3.Application{
					GUID:                "response-app-guid",
					StackName:           "response-stack-name",
					LifecycleType:       constant.AppLifecycleTypeBuildpack,
					LifecycleBuildpacks: []string{"response-buildpack-1", "response-buildpack-2"},
				}
				fakeCloudControllerClient.UpdateApplicationReturns(
					apiResponseApp,
					ccv3.Warnings{"some-warning"},
					nil,
				)
			})

			It("creates and returns the application and warnings", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(resultApp).To(Equal(Application{
					GUID:                apiResponseApp.GUID,
					StackName:           apiResponseApp.StackName,
					LifecycleType:       apiResponseApp.LifecycleType,
					LifecycleBuildpacks: apiResponseApp.LifecycleBuildpacks,
				}))
				Expect(warnings).To(ConsistOf("some-warning"))

				Expect(fakeCloudControllerClient.UpdateApplicationCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.UpdateApplicationArgsForCall(0)).To(Equal(ccv3.Application{
					GUID:                submitApp.GUID,
					StackName:           submitApp.StackName,
					LifecycleType:       submitApp.LifecycleType,
					LifecycleBuildpacks: submitApp.LifecycleBuildpacks,
					Metadata:            (*ccv3.Metadata)(submitApp.Metadata),
				}))
			})
		})

		When("the cc client returns an error", func() {
			var expectedError error

			BeforeEach(func() {
				expectedError = errors.New("I am a CloudControllerClient Error")
				fakeCloudControllerClient.UpdateApplicationReturns(
					ccv3.Application{},
					ccv3.Warnings{"some-warning"},
					expectedError,
				)
			})

			It("raises the error and warnings", func() {
				Expect(err).To(MatchError(expectedError))
				Expect(warnings).To(ConsistOf("some-warning"))
			})
		})
	})

	Describe("PollStart", func() {
		var (
			appGUID string
			noWait  bool

			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			appGUID = "some-guid"
			noWait = false
		})

		JustBeforeEach(func() {
			warnings, executeErr = actor.PollStart(appGUID, noWait)
		})

		When("getting the application processes fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationProcessesReturns(nil, ccv3.Warnings{"get-app-warning-1", "get-app-warning-2"}, errors.New("some-error"))
			})

			It("returns the error and all warnings", func() {
				Expect(executeErr).To(MatchError(errors.New("some-error")))
				Expect(warnings).To(ConsistOf("get-app-warning-1", "get-app-warning-2"))
			})
		})

		When("getting the application processes succeeds", func() {
			var processes []ccv3.Process

			BeforeEach(func() {
				fakeConfig.StartupTimeoutReturns(time.Second)
				fakeConfig.PollingIntervalReturns(0)
			})

			When("there is a single process", func() {
				BeforeEach(func() {
					processes = []ccv3.Process{{GUID: "abc123"}}
					fakeCloudControllerClient.GetApplicationProcessesReturns(
						processes,
						ccv3.Warnings{"get-app-warning-1"}, nil)
				})

				When("the polling times out", func() {
					BeforeEach(func() {
						fakeConfig.StartupTimeoutReturns(time.Millisecond)
						fakeConfig.PollingIntervalReturns(time.Millisecond * 2)
						fakeCloudControllerClient.GetProcessInstancesReturns(
							[]ccv3.ProcessInstance{{State: constant.ProcessInstanceStarting}},
							ccv3.Warnings{"get-process-warning-1", "get-process-warning-2"},
							nil,
						)
					})

					It("returns the timeout error", func() {
						Expect(executeErr).To(MatchError(actionerror.StartupTimeoutError{}))
						Expect(warnings).To(ConsistOf("get-app-warning-1", "get-process-warning-1", "get-process-warning-2"))
					})

					It("gets polling and timeout values from the config", func() {
						Expect(fakeConfig.StartupTimeoutCallCount()).To(Equal(1))
						Expect(fakeConfig.PollingIntervalCallCount()).To(Equal(1))
					})
				})

				When("getting the process instances errors", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.GetProcessInstancesReturns(
							nil,
							ccv3.Warnings{"get-process-warning-1", "get-process-warning-2"},
							errors.New("some-error"),
						)
					})

					It("returns the error", func() {
						Expect(executeErr).To(MatchError("some-error"))
						Expect(warnings).To(ConsistOf("get-app-warning-1", "get-process-warning-1", "get-process-warning-2"))
					})
				})

				When("getting the process instances succeeds", func() {
					var (
						initialInstanceStates    []ccv3.ProcessInstance
						eventualInstanceStates   []ccv3.ProcessInstance
						processInstanceCallCount int
					)

					BeforeEach(func() {
						processInstanceCallCount = 0

						fakeCloudControllerClient.GetProcessInstancesStub = func(processGuid string) ([]ccv3.ProcessInstance, ccv3.Warnings, error) {
							defer func() { processInstanceCallCount++ }()
							if processInstanceCallCount == 0 {
								return initialInstanceStates,
									ccv3.Warnings{"get-process-warning-1", "get-process-warning-2"},
									nil
							} else {
								return eventualInstanceStates,
									ccv3.Warnings{fmt.Sprintf("get-process-warning-%d", processInstanceCallCount+2)},
									nil
							}
						}
					})

					When("there are no process instances", func() {
						BeforeEach(func() {
							initialInstanceStates = []ccv3.ProcessInstance{}
							eventualInstanceStates = []ccv3.ProcessInstance{}
						})

						It("should not return an error", func() {
							Expect(executeErr).NotTo(HaveOccurred())
						})

						It("should only call GetProcessInstances once before exiting", func() {
							Expect(processInstanceCallCount).To(Equal(1))
						})

						It("should return correct warnings", func() {
							Expect(warnings).To(ConsistOf("get-app-warning-1", "get-process-warning-1", "get-process-warning-2"))
						})
					})

					When("all instances become running by the second call", func() {
						BeforeEach(func() {
							initialInstanceStates = []ccv3.ProcessInstance{{State: constant.ProcessInstanceStarting}, {State: constant.ProcessInstanceStarting}}
							eventualInstanceStates = []ccv3.ProcessInstance{{State: constant.ProcessInstanceRunning}, {State: constant.ProcessInstanceRunning}}
						})

						It("should not return an error", func() {
							Expect(executeErr).NotTo(HaveOccurred())
						})

						It("should call GetProcessInstances twice", func() {
							Expect(processInstanceCallCount).To(Equal(2))
						})

						It("should return correct warnings", func() {
							Expect(warnings).To(ConsistOf("get-app-warning-1", "get-process-warning-1", "get-process-warning-2", "get-process-warning-3"))
						})
					})

					When("at least one instance has become running by the second call", func() {
						BeforeEach(func() {
							initialInstanceStates = []ccv3.ProcessInstance{{State: constant.ProcessInstanceStarting}, {State: constant.ProcessInstanceStarting}, {State: constant.ProcessInstanceStarting}}
							eventualInstanceStates = []ccv3.ProcessInstance{{State: constant.ProcessInstanceStarting}, {State: constant.ProcessInstanceStarting}, {State: constant.ProcessInstanceRunning}}
						})

						It("should not return an error", func() {
							Expect(executeErr).NotTo(HaveOccurred())
						})

						It("should call GetProcessInstances twice", func() {
							Expect(processInstanceCallCount).To(Equal(2))
						})

						It("should return correct warnings", func() {
							Expect(warnings).To(ConsistOf("get-app-warning-1", "get-process-warning-1", "get-process-warning-2", "get-process-warning-3"))
						})
					})

					When("all of the instances have crashed by the second call", func() {
						BeforeEach(func() {
							initialInstanceStates = []ccv3.ProcessInstance{{State: constant.ProcessInstanceStarting}, {State: constant.ProcessInstanceStarting}, {State: constant.ProcessInstanceStarting}}
							eventualInstanceStates = []ccv3.ProcessInstance{{State: constant.ProcessInstanceCrashed}, {State: constant.ProcessInstanceCrashed}, {State: constant.ProcessInstanceCrashed}}
						})

						It("should not return an error", func() {
							Expect(executeErr).To(MatchError(actionerror.AllInstancesCrashedError{}))
						})

						It("should call GetProcessInstances twice", func() {
							Expect(processInstanceCallCount).To(Equal(2))
						})

						It("should return correct warnings", func() {
							Expect(warnings).To(ConsistOf("get-app-warning-1", "get-process-warning-1", "get-process-warning-2", "get-process-warning-3"))
						})
					})
				})
			})

			When("there are multiple processes", func() {
				var (
					processInstanceCallCount int
				)

				BeforeEach(func() {
					processInstanceCallCount = 0
					fakeConfig.StartupTimeoutReturns(time.Millisecond)
					fakeConfig.PollingIntervalReturns(time.Millisecond * 2)

					fakeCloudControllerClient.GetProcessInstancesStub = func(processGuid string) ([]ccv3.ProcessInstance, ccv3.Warnings, error) {
						defer func() { processInstanceCallCount++ }()
						if strings.HasPrefix(processGuid, "good") {
							return []ccv3.ProcessInstance{{State: constant.ProcessInstanceRunning}}, nil, nil
						} else {
							return []ccv3.ProcessInstance{{State: constant.ProcessInstanceStarting}}, nil, nil
						}
					}
				})

				When("none of the processes are ready", func() {
					BeforeEach(func() {
						processes = []ccv3.Process{{GUID: "bad-1"}, {GUID: "bad-2"}}
						fakeCloudControllerClient.GetApplicationProcessesReturns(
							processes,
							ccv3.Warnings{"get-app-warning-1"}, nil)
					})

					It("returns the timeout error", func() {
						Expect(executeErr).To(MatchError(actionerror.StartupTimeoutError{}))
					})
				})

				When("some of the processes are ready", func() {
					BeforeEach(func() {
						processes = []ccv3.Process{{GUID: "bad-1"}, {GUID: "good-1"}}
						fakeCloudControllerClient.GetApplicationProcessesReturns(
							processes,
							ccv3.Warnings{"get-app-warning-1"}, nil)
					})

					It("returns the timeout error", func() {
						Expect(executeErr).To(MatchError(actionerror.StartupTimeoutError{}))
					})
				})

				When("all of the processes are ready", func() {
					BeforeEach(func() {
						processes = []ccv3.Process{{GUID: "good-1"}, {GUID: "good-2"}}
						fakeCloudControllerClient.GetApplicationProcessesReturns(
							processes,
							ccv3.Warnings{"get-app-warning-1"}, nil)
					})

					It("returns nil", func() {
						Expect(executeErr).ToNot(HaveOccurred())
					})
				})
			})

			When("there are multiple processes, and noWait is true", func() {
				BeforeEach(func() {
					processes = []ccv3.Process{{GUID: "worker-guid", Type: "worker"}, {GUID: "web-guid", Type: "web"}}
					fakeCloudControllerClient.GetApplicationProcessesReturns(
						processes,
						ccv3.Warnings{"get-app-warning-1"}, nil)
					noWait = true
				})

				It("only polls the web process", func() {
					Expect(fakeCloudControllerClient.GetProcessInstancesCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetProcessInstancesArgsForCall(0)).To(Equal("web-guid"))
				})
			})
		})
	})

	Describe("PollStartForRolling", func() {
		var (
			appGUID        string
			deploymentGUID string
			noWait         bool

			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			appGUID = "some-rolling-app-guid"
			deploymentGUID = "some-deployment-guid"
			fakeConfig.StartupTimeoutReturns(10 * time.Second)
		})

		JustBeforeEach(func() {
			warnings, executeErr = actor.PollStartForRolling(appGUID, deploymentGUID, noWait)
		})

		When("getting the deployment fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetDeploymentReturns(
					ccv3.Deployment{},
					ccv3.Warnings{"get-deployment-warning"},
					errors.New("get-deployment-error"),
				)
			})

			It("returns warnings and the error", func() {
				Expect(executeErr).To(MatchError("get-deployment-error"))
				Expect(warnings).To(ConsistOf("get-deployment-warning"))

				Expect(fakeCloudControllerClient.GetDeploymentCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetDeploymentArgsForCall(0)).To(Equal(deploymentGUID))

				Expect(fakeCloudControllerClient.GetApplicationProcessesCallCount()).To(Equal(0))
				Expect(fakeCloudControllerClient.GetProcessInstancesCallCount()).To(Equal(0))

				Expect(fakeConfig.StartupTimeoutCallCount()).To(Equal(1))
			})
		})

		When("getting the deployment times out", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetDeploymentReturns(
					ccv3.Deployment{
						StatusValue:  constant.DeploymentStatusValueDeploying,
						NewProcesses: []ccv3.Process{{Type: "new-web", GUID: "new-web-guid"}},
					},
					ccv3.Warnings{"get-deployment-warning"},
					nil,
				)
				fakeConfig.StartupTimeoutReturns(time.Millisecond * 5)
				fakeConfig.PollingIntervalReturns(time.Millisecond * 30) // should run once
			})

			When("--no-wait is not specified", func() {
				BeforeEach(func() {
					noWait = false
				})

				It("returns the timeout error", func() {
					Expect(executeErr).To(MatchError(actionerror.StartupTimeoutError{}))
					Expect(warnings).To(ConsistOf("get-deployment-warning"))

					Expect(fakeConfig.StartupTimeoutCallCount()).To(Equal(1))
					Expect(fakeConfig.PollingIntervalCallCount()).To(Equal(1))
				})
			})
		})

		When("getting the deployment is successful", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetDeploymentReturnsOnCall(0,
					ccv3.Deployment{
						StatusValue:  constant.DeploymentStatusValueDeploying,
						NewProcesses: []ccv3.Process{{Type: "new-web", GUID: "new-web-guid"}},
					},
					ccv3.Warnings{"get-deployment-warning"},
					nil,
				)

				fakeCloudControllerClient.GetDeploymentReturnsOnCall(1,
					ccv3.Deployment{
						StatusValue:  constant.DeploymentStatusValueDeploying,
						NewProcesses: []ccv3.Process{{Type: "new-web", GUID: "new-web-guid"}},
					},
					ccv3.Warnings{"get-deployment-warning"},
					nil,
				)

				fakeCloudControllerClient.GetDeploymentReturnsOnCall(2,
					ccv3.Deployment{
						StatusValue:  constant.DeploymentStatusValueFinalized,
						StatusReason: constant.DeploymentStatusReasonDeployed,
						NewProcesses: []ccv3.Process{{Type: "new-web", GUID: "new-web-guid"}},
					},
					ccv3.Warnings{"get-deployment-warning"},
					nil,
				)
			})

			When("--no-wait is specified", func() {
				BeforeEach(func() {
					noWait = true
				})

				It("does not get the processes or poll the deployment", func() {
					Expect(fakeCloudControllerClient.GetApplicationProcessesCallCount()).To(Equal(0))
				})

				When("getting the instances fails", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.GetProcessInstancesReturnsOnCall(0,
							[]ccv3.ProcessInstance{
								{State: constant.ProcessInstanceStarting},
								{State: constant.ProcessInstanceStarting},
							},
							ccv3.Warnings{"poll-process-warning1"},
							errors.New("some-instance-error"),
						)
					})

					It("does not check the deployment a second time", func() {
						Expect(executeErr.Error()).To(Equal("some-instance-error"))
						Expect(warnings).To(ConsistOf("get-deployment-warning", "poll-process-warning1"))

						Expect(fakeCloudControllerClient.GetDeploymentCallCount()).To(Equal(1))
					})
				})

				When("the polling times out on the first loop", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.GetDeploymentReturns(
							ccv3.Deployment{
								StatusValue:  constant.DeploymentStatusValueFinalized,
								StatusReason: constant.DeploymentStatusReasonDeployed,
								NewProcesses: []ccv3.Process{{Type: "new-web", GUID: "new-web-guid"}},
							},
							ccv3.Warnings{"get-deployment-warning"},
							nil,
						)

						fakeCloudControllerClient.GetProcessInstancesReturns(
							[]ccv3.ProcessInstance{
								{State: constant.ProcessInstanceStarting},
								{State: constant.ProcessInstanceStarting},
							},
							ccv3.Warnings{"poll-process-warning"},
							nil,
						)

						fakeConfig.StartupTimeoutReturns(time.Millisecond * 5)
						fakeConfig.PollingIntervalReturns(time.Millisecond * 8)
					})

					It("returns the timeout error", func() {
						Expect(executeErr).To(MatchError(actionerror.StartupTimeoutError{}))
						Expect(warnings).To(ConsistOf("get-deployment-warning", "poll-process-warning"))

						Expect(fakeCloudControllerClient.GetDeploymentCallCount()).To(Equal(1))
						Expect(fakeCloudControllerClient.GetProcessInstancesCallCount()).To(Equal(1))

						Expect(fakeConfig.StartupTimeoutCallCount()).To(Equal(1))
						Expect(fakeConfig.PollingIntervalCallCount()).To(Equal(1))
					})
				})

				When("polling instances take time to become healthy", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.GetProcessInstancesReturnsOnCall(0,
							[]ccv3.ProcessInstance{
								{State: constant.ProcessInstanceStarting},
								{State: constant.ProcessInstanceStarting},
							},
							ccv3.Warnings{"poll-process-warning1"},
							nil,
						)

						fakeCloudControllerClient.GetProcessInstancesReturnsOnCall(1,
							[]ccv3.ProcessInstance{
								{State: constant.ProcessInstanceRunning},
								{State: constant.ProcessInstanceStarting},
							},
							ccv3.Warnings{"poll-process-warning2"},
							nil,
						)
					})

					It("continues polling until one instance is healthy", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(warnings).To(ConsistOf("get-deployment-warning", "poll-process-warning1", "get-deployment-warning", "poll-process-warning2"))
						Expect(fakeCloudControllerClient.GetProcessInstancesCallCount()).To(Equal(2))
						Expect(fakeCloudControllerClient.GetProcessInstancesArgsForCall(0)).To(Equal("new-web-guid"))
						Expect(fakeCloudControllerClient.GetProcessInstancesArgsForCall(1)).To(Equal("new-web-guid"))
						Expect(fakeCloudControllerClient.GetDeploymentCallCount()).To(Equal(2))
					})
				})

				When("the deployment is canceled", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.GetProcessInstancesReturnsOnCall(0,
							[]ccv3.ProcessInstance{
								{State: constant.ProcessInstanceStarting},
								{State: constant.ProcessInstanceStarting},
							},
							ccv3.Warnings{"poll-process-warning1"},
							nil,
						)

						fakeCloudControllerClient.GetProcessInstancesReturnsOnCall(1,
							[]ccv3.ProcessInstance{
								{State: constant.ProcessInstanceStarting},
								{State: constant.ProcessInstanceStarting},
							},
							ccv3.Warnings{"poll-process-warning2"},
							nil,
						)

						fakeCloudControllerClient.GetDeploymentReturnsOnCall(0,
							ccv3.Deployment{
								StatusValue:  constant.DeploymentStatusValueDeploying,
								NewProcesses: []ccv3.Process{{Type: "new-web", GUID: "new-web-guid"}},
							},
							ccv3.Warnings{"get-deployment-warning1"},
							nil,
						)

						fakeCloudControllerClient.GetDeploymentReturnsOnCall(1,
							ccv3.Deployment{
								StatusValue:  constant.DeploymentStatusValueDeploying,
								NewProcesses: []ccv3.Process{{Type: "new-web", GUID: "new-web-guid"}},
							},
							ccv3.Warnings{"get-deployment-warning2"},
							nil,
						)

						fakeCloudControllerClient.GetDeploymentReturnsOnCall(2,
							ccv3.Deployment{
								StatusValue:  constant.DeploymentStatusValueFinalized,
								StatusReason: constant.DeploymentStatusReasonCanceled,
								NewProcesses: []ccv3.Process{{Type: "new-web", GUID: "new-web-guid"}},
							},
							ccv3.Warnings{"get-deployment-warning3"},
							nil,
						)
					})

					It("stops polling and returns an error", func() {
						Expect(executeErr).To(MatchError("Deployment has been canceled"))
						Expect(warnings).To(ConsistOf("get-deployment-warning1", "poll-process-warning1",
							"get-deployment-warning2", "poll-process-warning2", "get-deployment-warning3"))

						Expect(fakeCloudControllerClient.GetDeploymentCallCount()).To(Equal(3))
						Expect(fakeCloudControllerClient.GetProcessInstancesCallCount()).To(Equal(2))
					})
				})

				When("the deployment is superseded", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.GetProcessInstancesReturnsOnCall(0,
							[]ccv3.ProcessInstance{
								{State: constant.ProcessInstanceStarting},
								{State: constant.ProcessInstanceStarting},
							},
							ccv3.Warnings{"poll-process-warning1"},
							nil,
						)

						fakeCloudControllerClient.GetProcessInstancesReturnsOnCall(1,
							[]ccv3.ProcessInstance{
								{State: constant.ProcessInstanceStarting},
								{State: constant.ProcessInstanceStarting},
							},
							ccv3.Warnings{"poll-process-warning2"},
							nil,
						)

						fakeCloudControllerClient.GetDeploymentReturnsOnCall(0,
							ccv3.Deployment{
								StatusValue:  constant.DeploymentStatusValueDeploying,
								NewProcesses: []ccv3.Process{{Type: "new-web", GUID: "new-web-guid"}},
							},
							ccv3.Warnings{"get-deployment-warning1"},
							nil,
						)

						fakeCloudControllerClient.GetDeploymentReturnsOnCall(1,
							ccv3.Deployment{
								StatusValue:  constant.DeploymentStatusValueDeploying,
								NewProcesses: []ccv3.Process{{Type: "new-web", GUID: "new-web-guid"}},
							},
							ccv3.Warnings{"get-deployment-warning2"},
							nil,
						)

						fakeCloudControllerClient.GetDeploymentReturnsOnCall(2,
							ccv3.Deployment{
								StatusValue:  constant.DeploymentStatusValueFinalized,
								StatusReason: constant.DeploymentStatusReasonSuperseded,
								NewProcesses: []ccv3.Process{{Type: "new-web", GUID: "new-web-guid"}},
							},
							ccv3.Warnings{"get-deployment-warning3"},
							nil,
						)
					})

					It("stops polling and returns an error", func() {
						Expect(executeErr).To(MatchError("Deployment has been superseded"))
						Expect(warnings).To(ConsistOf("get-deployment-warning1", "poll-process-warning1",
							"get-deployment-warning2", "poll-process-warning2", "get-deployment-warning3"))

						Expect(fakeCloudControllerClient.GetDeploymentCallCount()).To(Equal(3))
						Expect(fakeCloudControllerClient.GetProcessInstancesCallCount()).To(Equal(2))
					})
				})

				When("All instances are crashing", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.GetProcessInstancesReturnsOnCall(0,
							[]ccv3.ProcessInstance{
								{State: constant.ProcessInstanceStarting},
								{State: constant.ProcessInstanceStarting},
							},
							ccv3.Warnings{"poll-process-warning1"},
							nil,
						)

						fakeCloudControllerClient.GetProcessInstancesReturnsOnCall(1,
							[]ccv3.ProcessInstance{
								{State: constant.ProcessInstanceStarting},
								{State: constant.ProcessInstanceCrashed},
							},
							ccv3.Warnings{"poll-process-warning2"},
							nil,
						)

						fakeCloudControllerClient.GetProcessInstancesReturnsOnCall(2,
							[]ccv3.ProcessInstance{
								{State: constant.ProcessInstanceCrashed},
								{State: constant.ProcessInstanceCrashed},
							},
							ccv3.Warnings{"poll-process-warning3"},
							nil,
						)

						fakeCloudControllerClient.GetDeploymentReturns(
							ccv3.Deployment{
								StatusValue:  constant.DeploymentStatusValueDeploying,
								NewProcesses: []ccv3.Process{{Type: "new-web", GUID: "new-web-guid"}},
							},
							ccv3.Warnings{"get-deployment-warning"},
							nil,
						)
					})

					It("returns an AllInstancesCrashedError", func() {
						Expect(executeErr).To(MatchError(actionerror.AllInstancesCrashedError{}))
						Expect(warnings).To(ConsistOf("get-deployment-warning", "poll-process-warning1",
							"get-deployment-warning", "poll-process-warning2", "get-deployment-warning", "poll-process-warning3"))

						Expect(fakeCloudControllerClient.GetDeploymentCallCount()).To(Equal(3))
						Expect(fakeCloudControllerClient.GetProcessInstancesCallCount()).To(Equal(3))
					})
				})
			})

			When("--no-wait is not specified", func() {
				BeforeEach(func() {
					noWait = false
				})

				It("polls the deployment 3 times", func() {
					Expect(fakeCloudControllerClient.GetDeploymentCallCount()).To(Equal(3))
					Expect(fakeCloudControllerClient.GetApplicationProcessesCallCount()).To(Equal(1))
				})

				When("getting the application processes fails", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.GetApplicationProcessesReturns(
							nil,
							ccv3.Warnings{"get-app-warning-1", "get-app-warning-2"},
							errors.New("some-error"),
						)
					})

					It("returns the error and all warnings", func() {
						Expect(executeErr).To(MatchError(errors.New("some-error")))
						Expect(warnings).To(ConsistOf(
							"get-deployment-warning",
							"get-deployment-warning",
							"get-deployment-warning",
							"get-app-warning-1",
							"get-app-warning-2",
						))
					})
				})
			})
		})
	})

	Describe("SetApplicationProcessHealthCheckTypeByNameAndSpace", func() {
		var (
			healthCheckType     constant.HealthCheckType
			healthCheckEndpoint string

			warnings Warnings
			err      error
			app      Application
		)

		BeforeEach(func() {
			healthCheckType = constant.HTTP
			healthCheckEndpoint = "some-http-endpoint"
		})

		JustBeforeEach(func() {
			app, warnings, err = actor.SetApplicationProcessHealthCheckTypeByNameAndSpace(
				"some-app-name",
				"some-space-guid",
				healthCheckType,
				healthCheckEndpoint,
				"some-process-type",
				42,
			)
		})

		When("getting application returns an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("some-error")
				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv3.Application{},
					ccv3.Warnings{"some-warning"},
					expectedErr,
				)
			})

			It("returns the error and warnings", func() {
				Expect(err).To(Equal(expectedErr))
				Expect(warnings).To(ConsistOf("some-warning"))
			})
		})

		When("application exists", func() {
			var ccv3App ccv3.Application

			BeforeEach(func() {
				ccv3App = ccv3.Application{
					GUID: "some-app-guid",
				}

				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv3.Application{ccv3App},
					ccv3.Warnings{"some-warning"},
					nil,
				)
			})

			When("setting the health check returns an error", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("some-error")
					fakeCloudControllerClient.GetApplicationProcessByTypeReturns(
						ccv3.Process{},
						ccv3.Warnings{"some-process-warning"},
						expectedErr,
					)
				})

				It("returns the error and warnings", func() {
					Expect(err).To(Equal(expectedErr))
					Expect(warnings).To(ConsistOf("some-warning", "some-process-warning"))
				})
			})

			When("application process exists", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetApplicationProcessByTypeReturns(
						ccv3.Process{GUID: "some-process-guid"},
						ccv3.Warnings{"some-process-warning"},
						nil,
					)

					fakeCloudControllerClient.UpdateProcessReturns(
						ccv3.Process{GUID: "some-process-guid"},
						ccv3.Warnings{"some-health-check-warning"},
						nil,
					)
				})

				It("returns the application", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(warnings).To(ConsistOf("some-warning", "some-process-warning", "some-health-check-warning"))

					Expect(app).To(Equal(Application{
						GUID: ccv3App.GUID,
					}))

					Expect(fakeCloudControllerClient.GetApplicationProcessByTypeCallCount()).To(Equal(1))
					appGUID, processType := fakeCloudControllerClient.GetApplicationProcessByTypeArgsForCall(0)
					Expect(appGUID).To(Equal("some-app-guid"))
					Expect(processType).To(Equal("some-process-type"))

					Expect(fakeCloudControllerClient.UpdateProcessCallCount()).To(Equal(1))
					process := fakeCloudControllerClient.UpdateProcessArgsForCall(0)
					Expect(process.GUID).To(Equal("some-process-guid"))
					Expect(process.HealthCheckType).To(Equal(constant.HTTP))
					Expect(process.HealthCheckEndpoint).To(Equal("some-http-endpoint"))
					Expect(process.HealthCheckInvocationTimeout).To(BeEquivalentTo(42))
				})
			})
		})
	})

	Describe("StopApplication", func() {
		var (
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			warnings, executeErr = actor.StopApplication("some-app-guid")
		})

		When("there are no client errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.UpdateApplicationStopReturns(
					ccv3.Application{GUID: "some-app-guid"},
					ccv3.Warnings{"stop-application-warning"},
					nil,
				)
			})

			It("stops the application", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("stop-application-warning"))

				Expect(fakeCloudControllerClient.UpdateApplicationStopCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.UpdateApplicationStopArgsForCall(0)).To(Equal("some-app-guid"))
			})
		})

		When("stopping the application fails", func() {
			var expectedErr error
			BeforeEach(func() {
				expectedErr = errors.New("some set stop-application error")
				fakeCloudControllerClient.UpdateApplicationStopReturns(
					ccv3.Application{},
					ccv3.Warnings{"stop-application-warning"},
					expectedErr,
				)
			})

			It("returns the error", func() {
				Expect(executeErr).To(Equal(expectedErr))
				Expect(warnings).To(ConsistOf("stop-application-warning"))
			})
		})
	})

	Describe("StartApplication", func() {
		var (
			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			fakeConfig.StartupTimeoutReturns(time.Second)
			fakeConfig.PollingIntervalReturns(0)
		})

		JustBeforeEach(func() {
			warnings, executeErr = actor.StartApplication("some-app-guid")
		})

		When("there are no client errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.UpdateApplicationStartReturns(
					ccv3.Application{GUID: "some-app-guid"},
					ccv3.Warnings{"start-application-warning"},
					nil,
				)
			})

			It("starts the application", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("start-application-warning"))

				Expect(fakeCloudControllerClient.UpdateApplicationStartCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.UpdateApplicationStartArgsForCall(0)).To(Equal("some-app-guid"))
			})
		})

		When("starting the application fails", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("some set start-application error")
				fakeCloudControllerClient.UpdateApplicationStartReturns(
					ccv3.Application{},
					ccv3.Warnings{"start-application-warning"},
					expectedErr,
				)
			})

			It("returns the error", func() {
				warnings, err := actor.StartApplication("some-app-guid")

				Expect(err).To(Equal(expectedErr))
				Expect(warnings).To(ConsistOf("start-application-warning"))
			})
		})
	})

	Describe("RestartApplication", func() {
		var (
			warnings   Warnings
			executeErr error
			noWait     bool
		)

		BeforeEach(func() {
			fakeConfig.StartupTimeoutReturns(time.Second)
			fakeConfig.PollingIntervalReturns(0)
			noWait = false
		})

		JustBeforeEach(func() {
			warnings, executeErr = actor.RestartApplication("some-app-guid", noWait)
		})

		When("restarting the application is successful", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.UpdateApplicationRestartReturns(
					ccv3.Application{GUID: "some-app-guid"},
					ccv3.Warnings{"restart-application-warning"},
					nil,
				)
			})

			When("the noWait flag is passed", func() {
				BeforeEach(func() {
					processes := []ccv3.Process{{GUID: "some-web-process-guid", Type: "web"}, {GUID: "some-worker-process-guid", Type: "worker"}}
					fakeCloudControllerClient.GetApplicationProcessesReturns(processes, ccv3.Warnings{"get-process-warnings"}, nil)

					noWait = true
				})

				It("only polls the web process", func() {
					Expect(fakeCloudControllerClient.GetProcessInstancesCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetProcessInstancesArgsForCall(0)).To(Equal("some-web-process-guid"))
				})
			})

			When("polling the application start is successful", func() {
				BeforeEach(func() {
					processes := []ccv3.Process{{GUID: "some-process-guid"}}
					fakeCloudControllerClient.GetApplicationProcessesReturns(processes, ccv3.Warnings{"get-process-warnings"}, nil)
					fakeCloudControllerClient.GetProcessInstancesReturns(nil, ccv3.Warnings{"some-process-instance-warnings"}, nil)
				})

				It("returns all the warnings", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("restart-application-warning", "get-process-warnings", "some-process-instance-warnings"))
				})

				It("calls restart", func() {
					Expect(fakeCloudControllerClient.UpdateApplicationRestartCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.UpdateApplicationRestartArgsForCall(0)).To(Equal("some-app-guid"))
				})

				It("polls for the application's process to start", func() {
					Expect(fakeCloudControllerClient.GetApplicationProcessesCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetApplicationProcessesArgsForCall(0)).To(Equal("some-app-guid"))

					Expect(fakeCloudControllerClient.GetProcessInstancesCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetProcessInstancesArgsForCall(0)).To(Equal("some-process-guid"))
				})
			})

			When("polling the application start errors", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("some polling error")
					fakeCloudControllerClient.GetApplicationProcessesReturns(nil, ccv3.Warnings{"get-process-warnings"}, expectedErr)
				})

				It("returns the warnings and error", func() {
					Expect(executeErr).To(Equal(expectedErr))
					Expect(warnings).To(ConsistOf("restart-application-warning", "get-process-warnings"))
				})
			})
		})

		When("restarting the application fails", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("some set restart-application error")
				fakeCloudControllerClient.UpdateApplicationRestartReturns(
					ccv3.Application{},
					ccv3.Warnings{"restart-application-warning"},
					expectedErr,
				)
			})

			It("returns the warnings and error", func() {
				Expect(executeErr).To(Equal(expectedErr))
				Expect(warnings).To(ConsistOf("restart-application-warning"))
			})
		})
	})
})
