package v7action_test

import (
	"errors"
	"fmt"
	"time"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/batcher"
	"code.cloudfoundry.org/clock/fakeclock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Application Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
		fakeConfig                *v7actionfakes.FakeConfig
		fakeClock                 *fakeclock.FakeClock
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v7actionfakes.FakeCloudControllerClient)
		fakeConfig = new(v7actionfakes.FakeConfig)
		fakeClock = fakeclock.NewFakeClock(time.Now())
		actor = NewActor(fakeCloudControllerClient, fakeConfig, nil, nil, nil, fakeClock)
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
				fakeCloudControllerClient.GetApplicationsReturns([]resources.Application{}, ccv3.Warnings{"some-get-app-warning"}, errors.New("some-get-app-error"))
			})

			It("returns the warnings and error", func() {
				Expect(warnings).To(ConsistOf("some-get-app-warning"))
				Expect(executeErr).To(MatchError("some-get-app-error"))
			})
		})

		When("looking up the app guid succeeds without routes", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns([]resources.Application{{Name: "some-app", GUID: "abc123"}}, ccv3.Warnings{"some-get-app-warning"}, nil)
				deleteMappedRoutes = false
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

		When("looking up the app guid succeeds with routes", func() {
			BeforeEach(func() {
				deleteMappedRoutes = true
				fakeCloudControllerClient.GetApplicationsReturns([]resources.Application{{Name: "some-app", GUID: "abc123"}}, nil, nil)
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
						fakeCloudControllerClient.GetApplicationRoutesReturns([]resources.Route{}, nil, nil)
					})

					It("does not delete any routes", func() {
						Expect(fakeCloudControllerClient.DeleteRouteCallCount()).To(Equal(0))
					})
				})

				When("there are routes", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.GetApplicationRoutesReturns([]resources.Route{{GUID: "route-1-guid"}, {GUID: "route-2-guid", URL: "route-2.example.com"}}, nil, nil)
					})

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
							fakeCloudControllerClient.GetApplicationRoutesReturns(
								[]resources.Route{
									{GUID: "route-1-guid"},
									{GUID: "route-2-guid",
										URL: "route-2.example.com",
										Destinations: []resources.RouteDestination{
											{App: resources.RouteDestinationApp{GUID: "abc123"}},
											{App: resources.RouteDestinationApp{GUID: "different-app-guid"}},
										},
									},
								},
								nil,
								nil,
							)
						})

						It("refuses the entire operation", func() {
							Expect(executeErr).To(MatchError(actionerror.RouteBoundToMultipleAppsError{AppName: "some-app", RouteURL: "route-2.example.com"}))
							Expect(warnings).To(BeEmpty())
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
						})

						It("returns the error", func() {
							Expect(executeErr).To(MatchError("delete-route-2-error"))
							Expect(warnings).To(ConsistOf("delete-route-1-warning", "delete-route-2-warning"))
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
					[]resources.Application{
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
					resources.Application{
						Name: "some-app-name",
						GUID: "some-app-guid",
					},
					resources.Application{
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
					[]resources.Application{
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
				_, warnings, err := actor.GetApplicationsByGUIDs([]string{"some-app-guid", "nonexistent-app-guid"})
				Expect(warnings).To(ConsistOf("some-warning"))
				Expect(err).To(MatchError(actionerror.ApplicationsNotFoundError{}))
			})
		})

		When("a single app has two routes", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]resources.Application{
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
					[]resources.Application{},
					ccv3.Warnings{"some-warning"},
					expectedError)
			})

			It("returns the warnings and the error", func() {
				_, warnings, err := actor.GetApplicationsByGUIDs([]string{"some-app-guid"})
				Expect(warnings).To(ConsistOf("some-warning"))
				Expect(err).To(MatchError(expectedError))
			})
		})

		When("there are many guids", func() {
			const batches = 10
			var guids []string

			BeforeEach(func() {
				var apps []resources.Application

				for i := 0; i < batcher.BatchSize*batches; i++ {
					guids = append(guids, fmt.Sprintf("app-%d-guid", i))
					apps = append(apps, resources.Application{
						Name: fmt.Sprintf("app-%d-name", i),
						GUID: fmt.Sprintf("app-%d-guid", i),
					})
				}

				for b := 0; b < batches; b++ {
					fakeCloudControllerClient.GetApplicationsReturnsOnCall(
						b,
						apps[:batcher.BatchSize],
						ccv3.Warnings{"some-warning"},
						nil,
					)
					apps = apps[batcher.BatchSize:]
				}
			})

			It("makes many calls", func() {
				apps, warnings, err := actor.GetApplicationsByGUIDs(guids)
				Expect(len(apps)).To(Equal(batches * batcher.BatchSize))
				Expect(apps).To(HaveLen(batcher.BatchSize * 10))
				Expect(warnings).To(HaveLen(10))
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeCloudControllerClient.GetApplicationsCallCount()).To(Equal(10))
				Expect(fakeCloudControllerClient.GetApplicationsArgsForCall(0)).
					NotTo(Equal(fakeCloudControllerClient.GetApplicationsArgsForCall(1)))
			})
		})
	})

	Describe("GetApplicationsByNameAndSpace", func() {
		When("all of the requested apps exist", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]resources.Application{
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
					resources.Application{
						Name: "some-app-name",
						GUID: "some-app-guid",
					},
					resources.Application{
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
					[]resources.Application{
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
					[]resources.Application{
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
					[]resources.Application{},
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
					[]resources.Application{
						{
							Name: "some-app-name",
							GUID: "some-app-guid",
							Metadata: &resources.Metadata{
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
				Expect(app).To(Equal(resources.Application{
					Name: "some-app-name",
					GUID: "some-app-guid",
					Metadata: &resources.Metadata{
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
					[]resources.Application{},
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
					[]resources.Application{},
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
					[]resources.Application{
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
					resources.Application{
						GUID: "some-app-guid-1",
						Name: "some-app-1",
					},
					resources.Application{
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
					[]resources.Application{},
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
			application resources.Application
			warnings    Warnings
			err         error
		)

		JustBeforeEach(func() {
			application, warnings, err = actor.CreateApplicationInSpace(resources.Application{
				Name:                "some-app-name",
				LifecycleType:       constant.AppLifecycleTypeBuildpack,
				LifecycleBuildpacks: []string{"buildpack-1", "buildpack-2"},
			}, "some-space-guid")
		})

		When("the app successfully gets created", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CreateApplicationReturns(
					resources.Application{
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
				Expect(application).To(Equal(resources.Application{
					Name:                "some-app-name",
					GUID:                "some-app-guid",
					LifecycleType:       constant.AppLifecycleTypeBuildpack,
					LifecycleBuildpacks: []string{"buildpack-1", "buildpack-2"},
				}))
				Expect(warnings).To(ConsistOf("some-warning"))

				Expect(fakeCloudControllerClient.CreateApplicationCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.CreateApplicationArgsForCall(0)).To(Equal(resources.Application{
					Name:                "some-app-name",
					SpaceGUID:           "some-space-guid",
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
					resources.Application{},
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
					resources.Application{},
					ccv3.Warnings{"some-warning"},
					ccerror.NameNotUniqueInSpaceError{},
				)
			})

			It("returns the NameNotUniqueInSpaceError and warnings", func() {
				Expect(err).To(MatchError(ccerror.NameNotUniqueInSpaceError{}))
				Expect(warnings).To(ConsistOf("some-warning"))
			})
		})
	})

	Describe("UpdateApplication", func() {
		var (
			submitApp, resultApp resources.Application
			warnings             Warnings
			err                  error
		)

		JustBeforeEach(func() {
			submitApp = resources.Application{
				GUID:                "some-app-guid",
				StackName:           "some-stack-name",
				Name:                "some-app-name",
				LifecycleType:       constant.AppLifecycleTypeBuildpack,
				LifecycleBuildpacks: []string{"buildpack-1", "buildpack-2"},
				Metadata: &resources.Metadata{Labels: map[string]types.NullString{
					"some-label":  types.NewNullString("some-value"),
					"other-label": types.NewNullString("other-value"),
				}},
			}

			resultApp, warnings, err = actor.UpdateApplication(submitApp)
		})

		When("the app successfully gets updated", func() {
			var apiResponseApp resources.Application

			BeforeEach(func() {
				apiResponseApp = resources.Application{
					GUID:                "response-app-guid",
					StackName:           "response-stack-name",
					Name:                "response-app-name",
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
				Expect(resultApp).To(Equal(resources.Application{
					Name:                apiResponseApp.Name,
					GUID:                apiResponseApp.GUID,
					StackName:           apiResponseApp.StackName,
					LifecycleType:       apiResponseApp.LifecycleType,
					LifecycleBuildpacks: apiResponseApp.LifecycleBuildpacks,
				}))
				Expect(warnings).To(ConsistOf("some-warning"))

				Expect(fakeCloudControllerClient.UpdateApplicationCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.UpdateApplicationArgsForCall(0)).To(Equal(resources.Application{
					GUID:                submitApp.GUID,
					StackName:           submitApp.StackName,
					LifecycleType:       submitApp.LifecycleType,
					LifecycleBuildpacks: submitApp.LifecycleBuildpacks,
					Name:                submitApp.Name,
					Metadata:            submitApp.Metadata,
				}))
			})
		})

		When("the cc client returns an error", func() {
			var expectedError error

			BeforeEach(func() {
				expectedError = errors.New("I am a CloudControllerClient Error")
				fakeCloudControllerClient.UpdateApplicationReturns(
					resources.Application{},
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

	Describe("UpdateApplicationName", func() {
		var (
			resultApp           resources.Application
			newAppName, appGUID string
			warnings            Warnings
			err                 error
		)

		JustBeforeEach(func() {
			newAppName = "some-new-app-name"
			appGUID = "some-app-guid"

			resultApp, warnings, err = actor.UpdateApplicationName(newAppName, appGUID)
		})

		When("the app successfully gets updated", func() {
			var apiResponseApp resources.Application

			BeforeEach(func() {
				apiResponseApp = resources.Application{
					GUID:                "response-app-guid",
					StackName:           "response-stack-name",
					Name:                "response-app-name",
					LifecycleType:       constant.AppLifecycleTypeBuildpack,
					LifecycleBuildpacks: []string{"response-buildpack-1", "response-buildpack-2"},
				}
				fakeCloudControllerClient.UpdateApplicationNameReturns(
					apiResponseApp,
					ccv3.Warnings{"some-warning"},
					nil,
				)
			})

			It("creates and returns the application and warnings", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(resultApp).To(Equal(resources.Application{
					Name:                apiResponseApp.Name,
					GUID:                apiResponseApp.GUID,
					StackName:           apiResponseApp.StackName,
					LifecycleType:       apiResponseApp.LifecycleType,
					LifecycleBuildpacks: apiResponseApp.LifecycleBuildpacks,
				}))
				Expect(warnings).To(ConsistOf("some-warning"))

				Expect(fakeCloudControllerClient.UpdateApplicationNameCallCount()).To(Equal(1))
				appName, appGuid := fakeCloudControllerClient.UpdateApplicationNameArgsForCall(0)
				Expect(appName).To(Equal("some-new-app-name"))
				Expect(appGuid).To(Equal("some-app-guid"))
			})
		})

		When("the cc client returns an error", func() {
			var expectedError error

			BeforeEach(func() {
				expectedError = errors.New("I am a CloudControllerClient Error")
				fakeCloudControllerClient.UpdateApplicationNameReturns(
					resources.Application{},
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
			app                   resources.Application
			noWait                bool
			handleInstanceDetails func(string)

			done chan bool

			warnings                Warnings
			executeErr              error
			reportedInstanceDetails []string
		)

		BeforeEach(func() {
			done = make(chan bool)
			fakeConfig.StartupTimeoutReturns(2 * time.Second)
			fakeConfig.PollingIntervalReturns(1 * time.Second)
			app = resources.Application{GUID: "some-guid"}
			noWait = false

			reportedInstanceDetails = []string{}
			handleInstanceDetails = func(instanceDetails string) {
				reportedInstanceDetails = append(reportedInstanceDetails, instanceDetails)
			}
		})

		JustBeforeEach(func() {
			go func() {
				defer close(done)
				warnings, executeErr = actor.PollStart(app, noWait, handleInstanceDetails)
				done <- true
			}()
		})

		It("gets the apps processes", func() {
			// advanced clock so function exits
			fakeClock.WaitForNWatchersAndIncrement(1*time.Millisecond, 2)

			// wait for function to finish
			Eventually(done).Should(Receive(BeTrue()))

			Expect(fakeCloudControllerClient.GetApplicationProcessesCallCount()).To(Equal(1))
			Expect(fakeCloudControllerClient.GetApplicationProcessesArgsForCall(0)).To(Equal("some-guid"))

		})

		When("getting the application processes fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationProcessesReturns(nil, ccv3.Warnings{"get-app-warning-1", "get-app-warning-2"}, errors.New("some-error"))
			})

			It("returns the error and all warnings", func() {
				// wait for function to finish
				Eventually(done).Should(Receive(BeTrue()))

				Expect(executeErr).To(MatchError(errors.New("some-error")))
				Expect(warnings).To(ConsistOf("get-app-warning-1", "get-app-warning-2"))
			})
		})

		When("getting the application process succeeds", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationProcessesReturns(
					[]resources.Process{
						{GUID: "process1", Type: "web"},
					},
					ccv3.Warnings{"get-app-warning-1"},
					nil,
				)

			})

			It("gets the startup timeout", func() {
				// advanced clock so function exits
				fakeClock.WaitForNWatchersAndIncrement(1*time.Millisecond, 2)

				// wait for function to finish
				Eventually(done).Should(Receive(BeTrue()))

				Expect(fakeConfig.StartupTimeoutCallCount()).To(Equal(1))
			})

			When("the no-wait flag is provided", func() {
				BeforeEach(func() {
					noWait = true
					fakeCloudControllerClient.GetApplicationProcessesReturns(
						[]resources.Process{
							{GUID: "process1", Type: "web"},
							{GUID: "process2", Type: "worker"},
						},
						ccv3.Warnings{"get-app-warning-1"},
						nil,
					)
				})

				It("filters out the non web processes", func() {
					// send something on the timer channel
					fakeClock.WaitForNWatchersAndIncrement(1*time.Millisecond, 2)

					// Wait for function to finish
					Eventually(done).Should(Receive(BeTrue()))

					// assert on the cc call made within poll processes to make sure there is only the web process
					Expect(fakeCloudControllerClient.GetProcessInstancesCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetProcessInstancesArgsForCall(0)).To(Equal("process1"))

				})
			})

			When("polling processes returns an error", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetProcessInstancesReturns(nil, ccv3.Warnings{"poll-process-warning"}, errors.New("poll-process-error"))
				})

				It("returns the error and warnings", func() {
					// send something on the timer channel
					fakeClock.WaitForNWatchersAndIncrement(1*time.Millisecond, 2)

					// Wait for function to finish
					Eventually(done).Should(Receive(BeTrue()))

					Expect(executeErr).Should(MatchError("poll-process-error"))
					Expect(warnings).Should(ConsistOf("poll-process-warning", "get-app-warning-1"))
				})
			})

			When("polling start times out", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetProcessInstancesReturns(
						[]ccv3.ProcessInstance{
							{State: constant.ProcessInstanceStarting},
						},
						ccv3.Warnings{"poll-process-warning"},
						nil,
					)

					fakeConfig.StartupTimeoutReturns(2 * time.Millisecond)
				})

				It("returns a timeout error and any warnings", func() {
					// send something on the timer channel for first tick
					fakeClock.WaitForNWatchersAndIncrement(1*time.Millisecond, 2)

					fakeClock.Increment(1 * time.Millisecond)

					// Wait for function to finish
					Eventually(done).Should(Receive(BeTrue()))

					Expect(executeErr).To(MatchError(actionerror.StartupTimeoutError{}))
					Expect(warnings).To(ConsistOf("poll-process-warning", "get-app-warning-1"))
				})
			})

			When("polling process eventually returns we should stop polling", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetProcessInstancesReturnsOnCall(0,
						[]ccv3.ProcessInstance{
							{State: constant.ProcessInstanceStarting},
						},
						ccv3.Warnings{"poll-process-warning1"},
						nil,
					)

					fakeCloudControllerClient.GetProcessInstancesReturnsOnCall(1,
						[]ccv3.ProcessInstance{
							{State: constant.ProcessInstanceRunning},
						},
						ccv3.Warnings{"poll-process-warning2"},
						nil,
					)
				})

				It("returns success and any warnings", func() {
					// send something on the timer channel
					fakeClock.WaitForNWatchersAndIncrement(1*time.Millisecond, 2)

					Eventually(fakeConfig.PollingIntervalCallCount).Should(Equal(1))

					fakeClock.Increment(1 * time.Second)

					// Wait for function to finish
					Eventually(done).Should(Receive(BeTrue()))
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(warnings).To(ConsistOf("poll-process-warning1", "get-app-warning-1", "poll-process-warning2"))
				})

			})
		})
	})

	Describe("PollStartForRolling", func() {
		var (
			app                   resources.Application
			deploymentGUID        string
			noWait                bool
			handleInstanceDetails func(string)

			done chan bool

			warnings                Warnings
			executeErr              error
			reportedInstanceDetails []string
		)

		BeforeEach(func() {
			reportedInstanceDetails = []string{}
			handleInstanceDetails = func(instanceDetails string) {
				reportedInstanceDetails = append(reportedInstanceDetails, instanceDetails)
			}

			app = resources.Application{GUID: "some-rolling-app-guid"}
			deploymentGUID = "some-deployment-guid"
			noWait = false

			done = make(chan bool)

			fakeConfig.StartupTimeoutReturns(5 * time.Second)
			fakeConfig.PollingIntervalReturns(1 * time.Second)
		})

		JustBeforeEach(func() {
			go func() {
				warnings, executeErr = actor.PollStartForRolling(app, deploymentGUID, noWait, handleInstanceDetails)
				done <- true
			}()
		})

		When("There is a non-timeout failure in the loop", func() {
			// this may need to be expanded to also include when the deployment is superseded or cancelled
			When("getting the deployment fails", func() {
				When("it is because the deployment was cancelled", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.GetDeploymentReturns(
							resources.Deployment{
								StatusValue:  constant.DeploymentStatusValueFinalized,
								StatusReason: constant.DeploymentStatusReasonCanceled,
							},
							ccv3.Warnings{"get-deployment-warning"},
							nil,
						)
					})

					It("returns warnings and the error", func() {
						// initial tick
						fakeClock.WaitForNWatchersAndIncrement(1*time.Millisecond, 2)

						// wait for func to finish
						Eventually(done).Should(Receive(BeTrue()))

						Expect(executeErr).To(MatchError("Deployment has been canceled"))
						Expect(warnings).To(ConsistOf("get-deployment-warning"))

						Expect(fakeCloudControllerClient.GetDeploymentCallCount()).To(Equal(1))
						Expect(fakeCloudControllerClient.GetDeploymentArgsForCall(0)).To(Equal(deploymentGUID))

						Expect(fakeCloudControllerClient.GetApplicationProcessesCallCount()).To(Equal(0))
						Expect(fakeCloudControllerClient.GetProcessInstancesCallCount()).To(Equal(0))

						Expect(fakeConfig.StartupTimeoutCallCount()).To(Equal(1))
					})

				})

				When("it is because the deployment was superseded", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.GetDeploymentReturns(
							resources.Deployment{
								StatusValue:  constant.DeploymentStatusValueFinalized,
								StatusReason: constant.DeploymentStatusReasonSuperseded,
							},
							ccv3.Warnings{"get-deployment-warning"},
							nil,
						)
					})

					It("returns warnings and the error", func() {
						// initial tick
						fakeClock.WaitForNWatchersAndIncrement(1*time.Millisecond, 2)

						// wait for func to finish
						Eventually(done).Should(Receive(BeTrue()))

						Expect(executeErr).To(MatchError("Deployment has been superseded"))
						Expect(warnings).To(ConsistOf("get-deployment-warning"))

						Expect(fakeCloudControllerClient.GetDeploymentCallCount()).To(Equal(1))
						Expect(fakeCloudControllerClient.GetDeploymentArgsForCall(0)).To(Equal(deploymentGUID))

						Expect(fakeCloudControllerClient.GetApplicationProcessesCallCount()).To(Equal(0))
						Expect(fakeCloudControllerClient.GetProcessInstancesCallCount()).To(Equal(0))

						Expect(fakeConfig.StartupTimeoutCallCount()).To(Equal(1))
					})

				})

				When("it is because of an API error", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.GetDeploymentReturns(
							resources.Deployment{},
							ccv3.Warnings{"get-deployment-warning"},
							errors.New("get-deployment-error"),
						)
					})

					It("returns warnings and the error", func() {
						// initial tick
						fakeClock.WaitForNWatchersAndIncrement(1*time.Millisecond, 2)

						// wait for func to finish
						Eventually(done).Should(Receive(BeTrue()))

						Expect(executeErr).To(MatchError("get-deployment-error"))
						Expect(warnings).To(ConsistOf("get-deployment-warning"))

						Expect(fakeCloudControllerClient.GetDeploymentCallCount()).To(Equal(1))
						Expect(fakeCloudControllerClient.GetDeploymentArgsForCall(0)).To(Equal(deploymentGUID))

						Expect(fakeCloudControllerClient.GetApplicationProcessesCallCount()).To(Equal(0))
						Expect(fakeCloudControllerClient.GetProcessInstancesCallCount()).To(Equal(0))

						Expect(fakeConfig.StartupTimeoutCallCount()).To(Equal(1))
					})

				})
			})

			When("getting the deployment succeeds", func() {
				BeforeEach(func() {
					// get processes requires the deployment to be deployed so we need this to indirectly test the error case
					fakeCloudControllerClient.GetDeploymentReturns(
						resources.Deployment{StatusValue: constant.DeploymentStatusValueFinalized, StatusReason: constant.DeploymentStatusReasonDeployed},
						ccv3.Warnings{"get-deployment-warning"},
						nil,
					)

				})

				When("getting the processes fails", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.GetApplicationProcessesReturns(
							[]resources.Process{},
							ccv3.Warnings{"get-processes-warning"},
							errors.New("get-processes-error"),
						)
					})

					It("returns warnings and the error", func() {
						// initial tick
						fakeClock.WaitForNWatchersAndIncrement(1*time.Millisecond, 2)

						// wait for func to finish
						Eventually(done).Should(Receive(BeTrue()))

						Expect(executeErr).To(MatchError("get-processes-error"))
						Expect(warnings).To(ConsistOf("get-deployment-warning", "get-processes-warning"))

						Expect(fakeCloudControllerClient.GetDeploymentCallCount()).To(Equal(1))
						Expect(fakeCloudControllerClient.GetDeploymentArgsForCall(0)).To(Equal(deploymentGUID))

						Expect(fakeCloudControllerClient.GetApplicationProcessesCallCount()).To(Equal(1))
						Expect(fakeCloudControllerClient.GetApplicationProcessesArgsForCall(0)).To(Equal(app.GUID))

						Expect(fakeCloudControllerClient.GetProcessInstancesCallCount()).To(Equal(0))

					})
				})

				When("getting the processes succeeds", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.GetApplicationProcessesReturns(
							[]resources.Process{{GUID: "process-guid"}},
							ccv3.Warnings{"get-processes-warning"},
							nil,
						)
					})

					When("polling the processes fails", func() {
						BeforeEach(func() {
							fakeCloudControllerClient.GetProcessInstancesReturns(
								[]ccv3.ProcessInstance{},
								ccv3.Warnings{"poll-processes-warning"},
								errors.New("poll-processes-error"),
							)
						})

						It("returns all warnings and errors", func() {
							// initial tick
							fakeClock.WaitForNWatchersAndIncrement(1*time.Millisecond, 2)

							// wait for func to finish
							Eventually(done).Should(Receive(BeTrue()))

							Expect(executeErr).To(MatchError("poll-processes-error"))
							Expect(warnings).To(ConsistOf("get-deployment-warning", "get-processes-warning", "poll-processes-warning"))

							Expect(fakeCloudControllerClient.GetDeploymentCallCount()).To(Equal(1))
							Expect(fakeCloudControllerClient.GetDeploymentArgsForCall(0)).To(Equal(deploymentGUID))

							Expect(fakeCloudControllerClient.GetApplicationProcessesCallCount()).To(Equal(1))
							Expect(fakeCloudControllerClient.GetApplicationProcessesArgsForCall(0)).To(Equal(app.GUID))

							Expect(fakeCloudControllerClient.GetProcessInstancesCallCount()).To(Equal(1))
							Expect(fakeCloudControllerClient.GetProcessInstancesArgsForCall(0)).To(Equal("process-guid"))
						})

					})
				})

			})

		})

		// intentionally ignore the no-wait flag here for simplicity. One of these two things must cause timeout regardless of no-wait state
		When("there is a timeout error", func() {
			BeforeEach(func() {
				// 1 millisecond for initial tick then 1 to trigger timeout
				fakeConfig.StartupTimeoutReturns(2 * time.Millisecond)
			})

			When("the deployment never deploys", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetDeploymentReturns(
						resources.Deployment{StatusValue: constant.DeploymentStatusValueActive},
						ccv3.Warnings{"get-deployment-warning"},
						nil,
					)
					fakeCloudControllerClient.CancelDeploymentReturns(
						ccv3.Warnings{"cancel-deployment-warning"},
						nil,
					)
				})

				It("returns a timeout error and any warnings and cancels the deployment", func() {
					// initial tick
					fakeClock.WaitForNWatchersAndIncrement(1*time.Millisecond, 2)

					Eventually(fakeCloudControllerClient.GetDeploymentCallCount).Should(Equal(1))

					// timeout tick
					fakeClock.Increment(1 * time.Millisecond)

					Eventually(fakeCloudControllerClient.CancelDeploymentCallCount).Should(Equal(1))

					// wait for func to finish
					Eventually(done).Should(Receive(BeTrue()))

					Expect(executeErr).To(MatchError(actionerror.StartupTimeoutError{}))
					Expect(warnings).To(ConsistOf("get-deployment-warning", "cancel-deployment-warning"))
				})

				When("the cancel deployment fails", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.CancelDeploymentReturns(
							ccv3.Warnings{"cancel-deployment-warning"},
							errors.New("cancel-deployment-error"),
						)
					})

					It("returns a timeout error and any warnings and cancels the deployment", func() {
						// initial tick
						fakeClock.WaitForNWatchersAndIncrement(1*time.Millisecond, 2)

						Eventually(fakeCloudControllerClient.GetDeploymentCallCount).Should(Equal(1))

						// timeout tick
						fakeClock.Increment(1 * time.Millisecond)

						Eventually(fakeCloudControllerClient.CancelDeploymentCallCount).Should(Equal(1))

						// wait for func to finish
						Eventually(done).Should(Receive(BeTrue()))

						Expect(executeErr).To(MatchError("cancel-deployment-error"))
						Expect(warnings).To(ConsistOf("get-deployment-warning", "cancel-deployment-warning"))
					})

				})

			})

			When("the processes dont become healthy", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetDeploymentReturns(
						resources.Deployment{StatusValue: constant.DeploymentStatusValueFinalized, StatusReason: constant.DeploymentStatusReasonDeployed},
						ccv3.Warnings{"get-deployment-warning"},
						nil,
					)

					fakeCloudControllerClient.GetApplicationProcessesReturns(
						[]resources.Process{{GUID: "process-guid"}},
						ccv3.Warnings{"get-processes-warning"},
						nil,
					)

					fakeCloudControllerClient.GetProcessInstancesReturns(
						[]ccv3.ProcessInstance{{State: constant.ProcessInstanceStarting}},
						ccv3.Warnings{"poll-processes-warning"},
						nil,
					)
				})

				It("returns a timeout error and any warnings", func() {
					// initial tick
					fakeClock.WaitForNWatchersAndIncrement(1*time.Millisecond, 2)

					Eventually(fakeCloudControllerClient.GetDeploymentCallCount).Should(Equal(1))
					Eventually(fakeCloudControllerClient.GetApplicationProcessesCallCount).Should(Equal(1))
					Eventually(fakeCloudControllerClient.GetProcessInstancesCallCount).Should(Equal(1))

					// timeout tick
					fakeClock.Increment(1 * time.Millisecond)

					// wait for func to finish
					Eventually(done).Should(Receive(BeTrue()))

					Expect(executeErr).To(MatchError(actionerror.StartupTimeoutError{}))
					Expect(warnings).To(ConsistOf("get-deployment-warning", "get-processes-warning", "poll-processes-warning"))
				})

			})
		})

		When("things eventually become healthy", func() {
			When("the no wait flag is given", func() {
				BeforeEach(func() {
					// in total three loops 1: deployment still deploying 2: deployment deployed processes starting 3: processes started
					noWait = true

					// Always return deploying as a way to check we respect no wait
					fakeCloudControllerClient.GetDeploymentReturns(
						resources.Deployment{
							StatusValue:  constant.DeploymentStatusValueActive,
							NewProcesses: []resources.Process{{GUID: "new-deployment-process"}},
						},
						ccv3.Warnings{"get-deployment-warning"},
						nil,
					)

					// We only poll the processes. Two loops for fun
					fakeCloudControllerClient.GetProcessInstancesReturnsOnCall(0,
						[]ccv3.ProcessInstance{{State: constant.ProcessInstanceStarting}},
						ccv3.Warnings{"poll-processes-warning-1"},
						nil,
					)

					fakeCloudControllerClient.GetProcessInstancesReturnsOnCall(1,
						[]ccv3.ProcessInstance{{State: constant.ProcessInstanceRunning}},
						ccv3.Warnings{"poll-processes-warning-2"},
						nil,
					)
				})

				It("polls the start of the application correctly and returns warnings and no error", func() {
					// Initial tick
					fakeClock.WaitForNWatchersAndIncrement(1*time.Millisecond, 2)

					// assert one of our watcher is the timeout
					Expect(fakeConfig.StartupTimeoutCallCount()).To(Equal(1))

					// the first time through we always get the deployment regardless of no-wait
					Eventually(fakeCloudControllerClient.GetDeploymentCallCount).Should(Equal(1))
					Expect(fakeCloudControllerClient.GetDeploymentArgsForCall(0)).To(Equal(deploymentGUID))
					Eventually(fakeCloudControllerClient.GetProcessInstancesCallCount).Should(Equal(1))
					Expect(fakeCloudControllerClient.GetProcessInstancesArgsForCall(0)).To(Equal("new-deployment-process"))
					Eventually(fakeConfig.PollingIntervalCallCount).Should(Equal(1))

					fakeClock.Increment(1 * time.Second)

					Eventually(fakeCloudControllerClient.GetDeploymentCallCount).Should(Equal(2))
					Expect(fakeCloudControllerClient.GetDeploymentArgsForCall(0)).To(Equal(deploymentGUID))
					Eventually(fakeCloudControllerClient.GetProcessInstancesCallCount).Should(Equal(2))
					Expect(fakeCloudControllerClient.GetProcessInstancesArgsForCall(0)).To(Equal("new-deployment-process"))

					Eventually(done).Should(Receive(BeTrue()))

					Expect(executeErr).NotTo(HaveOccurred())
					Expect(warnings).To(ConsistOf(
						"get-deployment-warning",
						"poll-processes-warning-1",
						"get-deployment-warning",
						"poll-processes-warning-2",
					))

					Expect(fakeCloudControllerClient.GetDeploymentCallCount()).To(Equal(2))
					Expect(fakeCloudControllerClient.GetApplicationProcessesCallCount()).To(Equal(0))
					Expect(fakeCloudControllerClient.GetProcessInstancesCallCount()).To(Equal(2))
					Expect(fakeConfig.PollingIntervalCallCount()).To(Equal(1))

				})

			})

			When("the no wait flag is not given", func() {
				BeforeEach(func() {
					// in total three loops 1: deployment still deploying 2: deployment deployed processes starting 3: processes started
					fakeCloudControllerClient.GetDeploymentReturnsOnCall(0,
						resources.Deployment{StatusValue: constant.DeploymentStatusValueActive},
						ccv3.Warnings{"get-deployment-warning-1"},
						nil,
					)

					// Poll the deployment twice to make sure we are polling (one in the above before each)
					fakeCloudControllerClient.GetDeploymentReturnsOnCall(1,
						resources.Deployment{StatusValue: constant.DeploymentStatusValueFinalized, StatusReason: constant.DeploymentStatusReasonDeployed},
						ccv3.Warnings{"get-deployment-warning-2"},
						nil,
					)

					// then we get the processes. This should only be called once
					fakeCloudControllerClient.GetApplicationProcessesReturns(
						[]resources.Process{{GUID: "process-guid"}},
						ccv3.Warnings{"get-processes-warning"},
						nil,
					)

					// then we poll the processes. Two loops for fun
					fakeCloudControllerClient.GetProcessInstancesReturnsOnCall(0,
						[]ccv3.ProcessInstance{{State: constant.ProcessInstanceStarting}},
						ccv3.Warnings{"poll-processes-warning-1"},
						nil,
					)

					fakeCloudControllerClient.GetProcessInstancesReturnsOnCall(1,
						[]ccv3.ProcessInstance{{State: constant.ProcessInstanceRunning}},
						ccv3.Warnings{"poll-processes-warning-2"},
						nil,
					)
				})

				It("polls the start of the application correctly and returns warnings and no error", func() {
					// Initial tick
					fakeClock.WaitForNWatchersAndIncrement(1*time.Millisecond, 2)

					// assert one of our watchers is for the timeout
					Expect(fakeConfig.StartupTimeoutCallCount()).To(Equal(1))

					Eventually(fakeCloudControllerClient.GetDeploymentCallCount).Should(Equal(1))
					Expect(fakeCloudControllerClient.GetDeploymentArgsForCall(0)).To(Equal(deploymentGUID))
					Eventually(fakeConfig.PollingIntervalCallCount).Should(Equal(1))

					// start the second loop where the deployment is deployed so we poll processes
					fakeClock.Increment(1 * time.Second)

					Eventually(fakeCloudControllerClient.GetDeploymentCallCount).Should(Equal(2))
					Expect(fakeCloudControllerClient.GetDeploymentArgsForCall(1)).To(Equal(deploymentGUID))
					Eventually(fakeCloudControllerClient.GetApplicationProcessesCallCount).Should(Equal(1))
					Expect(fakeCloudControllerClient.GetApplicationProcessesArgsForCall(0)).To(Equal(app.GUID))
					Eventually(fakeCloudControllerClient.GetProcessInstancesCallCount).Should(Equal(1))
					Expect(fakeCloudControllerClient.GetProcessInstancesArgsForCall(0)).To(Equal("process-guid"))
					Eventually(fakeConfig.PollingIntervalCallCount).Should(Equal(2))

					fakeClock.Increment(1 * time.Second)

					// we should stop polling because it is deployed
					Eventually(fakeCloudControllerClient.GetProcessInstancesCallCount).Should(Equal(2))
					Expect(fakeCloudControllerClient.GetProcessInstancesArgsForCall(0)).To(Equal("process-guid"))

					Eventually(done).Should(Receive(BeTrue()))

					Expect(executeErr).NotTo(HaveOccurred())
					Expect(warnings).To(ConsistOf(
						"get-deployment-warning-1",
						"get-deployment-warning-2",
						"get-processes-warning",
						"poll-processes-warning-1",
						"poll-processes-warning-2",
					))

					Expect(fakeCloudControllerClient.GetDeploymentCallCount()).To(Equal(2))
					Expect(fakeCloudControllerClient.GetApplicationProcessesCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetProcessInstancesCallCount()).To(Equal(2))
					Expect(fakeConfig.PollingIntervalCallCount()).To(Equal(2))

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
			app      resources.Application
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
					[]resources.Application{},
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
			var ccv3App resources.Application

			BeforeEach(func() {
				ccv3App = resources.Application{
					GUID: "some-app-guid",
				}

				fakeCloudControllerClient.GetApplicationsReturns(
					[]resources.Application{ccv3App},
					ccv3.Warnings{"some-warning"},
					nil,
				)
			})

			When("setting the health check returns an error", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("some-error")
					fakeCloudControllerClient.GetApplicationProcessByTypeReturns(
						resources.Process{},
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
						resources.Process{GUID: "some-process-guid"},
						ccv3.Warnings{"some-process-warning"},
						nil,
					)

					fakeCloudControllerClient.UpdateProcessReturns(
						resources.Process{GUID: "some-process-guid"},
						ccv3.Warnings{"some-health-check-warning"},
						nil,
					)
				})

				It("returns the application", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(warnings).To(ConsistOf("some-warning", "some-process-warning", "some-health-check-warning"))

					Expect(app).To(Equal(resources.Application{
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
					resources.Application{GUID: "some-app-guid"},
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
					resources.Application{},
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
					resources.Application{GUID: "some-app-guid"},
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
					resources.Application{},
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
					resources.Application{GUID: "some-app-guid"},
					ccv3.Warnings{"restart-application-warning"},
					nil,
				)
			})

			It("does not error", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("restart-application-warning"))
			})
		})

		When("restarting the application fails", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("some set restart-application error")
				fakeCloudControllerClient.UpdateApplicationRestartReturns(
					resources.Application{},
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

	Describe("PollProcesses", func() {
		var (
			processes               []resources.Process
			handleInstanceDetails   func(string)
			reportedInstanceDetails []string

			keepPolling bool
			warnings    Warnings
			executeErr  error
		)

		BeforeEach(func() {
			reportedInstanceDetails = []string{}
			handleInstanceDetails = func(instanceDetails string) {
				reportedInstanceDetails = append(reportedInstanceDetails, instanceDetails)
			}

			processes = []resources.Process{
				{GUID: "process-1"},
				{GUID: "process-2"},
			}
		})

		JustBeforeEach(func() {
			keepPolling, warnings, executeErr = actor.PollProcesses(processes, handleInstanceDetails)
		})

		It("gets process instances for each process", func() {
			Expect(executeErr).NotTo(HaveOccurred())
			Expect(fakeCloudControllerClient.GetProcessInstancesCallCount()).To(Equal(2))
			Expect(fakeCloudControllerClient.GetProcessInstancesArgsForCall(0)).To(Equal("process-1"))
			Expect(fakeCloudControllerClient.GetProcessInstancesArgsForCall(1)).To(Equal("process-2"))
		})

		When("getting the process instances fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetProcessInstancesReturns(nil, ccv3.Warnings{"get-instances-warning"}, errors.New("get-instances-error"))
			})

			It("returns an error and warnings and terminates the loop", func() {
				Expect(executeErr).To(MatchError("get-instances-error"))
				Expect(warnings).To(ConsistOf("get-instances-warning"))
				Expect(keepPolling).To(BeTrue())

				Expect(fakeCloudControllerClient.GetProcessInstancesCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetProcessInstancesArgsForCall(0)).To(Equal("process-1"))
			})
		})

		When("getting the process instances is always successful", func() {
			When("a process has all instances crashed", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetProcessInstancesReturns(
						[]ccv3.ProcessInstance{
							{State: constant.ProcessInstanceCrashed, Details: "details1"},
						},
						ccv3.Warnings{"get-process1-instances-warning"},
						nil,
					)
				})

				It("calls the callback function with the retrieved instances", func() {
					Expect(reportedInstanceDetails).To(Equal([]string{
						"Error starting instances: 'details1'",
					}))
				})

				It("returns an all instances crashed error", func() {
					Expect(executeErr).To(MatchError(actionerror.AllInstancesCrashedError{}))
					Expect(warnings).To(ConsistOf("get-process1-instances-warning"))
					Expect(keepPolling).To(BeTrue())
				})
			})

			When("there are still instances in the starting state for a process", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetProcessInstancesReturnsOnCall(0,
						[]ccv3.ProcessInstance{
							{State: constant.ProcessInstanceRunning},
						},
						ccv3.Warnings{"get-process1-instances-warning"},
						nil,
					)

					fakeCloudControllerClient.GetProcessInstancesReturnsOnCall(1,
						[]ccv3.ProcessInstance{
							{State: constant.ProcessInstanceStarting, Details: "details2"},
						},
						ccv3.Warnings{"get-process2-instances-warning"},
						nil,
					)
				})

				It("calls the callback function with the retrieved instances", func() {
					Expect(reportedInstanceDetails).To(Equal([]string{
						"Instances starting...",
						"Error starting instances: 'details2'",
					}))
				})

				It("returns success and that we should keep polling", func() {
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(warnings).To(ConsistOf("get-process1-instances-warning", "get-process2-instances-warning"))
					Expect(keepPolling).To(BeFalse())
				})
			})

			When("all the instances of all processes are stable", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetProcessInstancesReturnsOnCall(0,
						[]ccv3.ProcessInstance{
							{State: constant.ProcessInstanceRunning, Details: "details1"},
						},
						ccv3.Warnings{"get-process1-instances-warning"},
						nil,
					)

					fakeCloudControllerClient.GetProcessInstancesReturnsOnCall(1,
						[]ccv3.ProcessInstance{
							{State: constant.ProcessInstanceRunning},
						},
						ccv3.Warnings{"get-process2-instances-warning"},
						nil,
					)
				})

				It("calls the callback function with the retrieved instances", func() {
					Expect(reportedInstanceDetails).To(Equal([]string{
						"Error starting instances: 'details1'",
						"Instances starting...",
					}))
				})

				It("returns success and that we should keep polling", func() {
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(warnings).To(ConsistOf("get-process1-instances-warning", "get-process2-instances-warning"))
					Expect(keepPolling).To(BeTrue())
				})

			})
		})

	})

	Describe("GetUnstagedNewestPackageGUID", func() {
		var (
			packageToStage string
			warnings       Warnings
			executeErr     error
		)

		JustBeforeEach(func() {
			packageToStage, warnings, executeErr = actor.GetUnstagedNewestPackageGUID("some-app-guid")
		})

		// Nothing to stage.
		When("There are no packages on the app", func() {
			When("getting the packages succeeds", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetPackagesReturns([]resources.Package{}, ccv3.Warnings{"get-packages-warnings"}, nil)
				})

				It("checks for packages", func() {
					Expect(fakeCloudControllerClient.GetPackagesCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetPackagesArgsForCall(0)).To(ConsistOf(
						ccv3.Query{Key: ccv3.AppGUIDFilter, Values: []string{"some-app-guid"}},
						ccv3.Query{Key: ccv3.OrderBy, Values: []string{ccv3.CreatedAtDescendingOrder}},
						ccv3.Query{Key: ccv3.PerPage, Values: []string{"1"}},
					))
				})

				It("returns empty string", func() {
					Expect(packageToStage).To(Equal(""))
					Expect(warnings).To(ConsistOf("get-packages-warnings"))
					Expect(executeErr).To(BeNil())
				})
			})

			When("getting the packages fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetPackagesReturns(
						nil,
						ccv3.Warnings{"get-packages-warnings"},
						errors.New("get-packages-error"),
					)
				})

				It("returns the error", func() {
					Expect(warnings).To(ConsistOf("get-packages-warnings"))
					Expect(executeErr).To(MatchError("get-packages-error"))
				})
			})
		})

		When("there are packages", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetPackagesReturns(
					[]resources.Package{{GUID: "package-guid", CreatedAt: "2019-01-01T06:00:00Z"}},
					ccv3.Warnings{"get-packages-warning"},
					nil)
			})

			It("checks for the packages latest droplet", func() {
				Expect(fakeCloudControllerClient.GetPackageDropletsCallCount()).To(Equal(1))
				packageGuid, queries := fakeCloudControllerClient.GetPackageDropletsArgsForCall(0)
				Expect(packageGuid).To(Equal("package-guid"))
				Expect(queries).To(ConsistOf(
					ccv3.Query{Key: ccv3.PerPage, Values: []string{"1"}},
					ccv3.Query{Key: ccv3.StatesFilter, Values: []string{"STAGED"}},
				))
			})

			When("the newest package's has a STAGED droplet", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetPackageDropletsReturns(
						[]resources.Droplet{{State: constant.DropletStaged}},
						ccv3.Warnings{"get-package-droplet-warning"},
						nil,
					)
				})

				It("returns empty string", func() {
					Expect(packageToStage).To(Equal(""))
					Expect(warnings).To(ConsistOf("get-packages-warning", "get-package-droplet-warning"))
					Expect(executeErr).To(BeNil())
				})
			})

			When("the package has no STAGED droplets", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetPackageDropletsReturns(
						[]resources.Droplet{},
						ccv3.Warnings{"get-package-droplet-warning"},
						nil,
					)
				})

				It("returns the guid of the newest package", func() {
					Expect(packageToStage).To(Equal("package-guid"))
					Expect(warnings).To(ConsistOf("get-packages-warning", "get-package-droplet-warning"))
					Expect(executeErr).To(BeNil())
				})
			})
		})
	})

	Describe("RenameApplicationByNameAndSpaceGUID", func() {
		When("the app does not exist", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]resources.Application{},
					ccv3.Warnings{"some-warning"},
					nil,
				)
			})

			It("returns an ApplicationNotFoundError and the warnings", func() {
				_, warnings, err := actor.RenameApplicationByNameAndSpaceGUID("old-app-name", "new-app-name", "space-guid")
				Expect(warnings).To(ConsistOf("some-warning"))
				Expect(err).To(MatchError(actionerror.ApplicationNotFoundError{Name: "old-app-name"}))
			})
		})

		When("the cloud controller client returns an error on application find", func() {
			var expectedError error

			BeforeEach(func() {
				expectedError = errors.New("I am a CloudControllerClient Error")
				fakeCloudControllerClient.GetApplicationsReturns(
					[]resources.Application{},
					ccv3.Warnings{"some-warning"},
					expectedError)
			})

			It("returns the warnings and the error", func() {
				_, warnings, err := actor.RenameApplicationByNameAndSpaceGUID("old-app-name", "new-app-name", "space-guid")
				Expect(warnings).To(ConsistOf("some-warning"))
				Expect(err).To(MatchError(expectedError))
			})
		})

		When("the cloud controller client returns an error on application update", func() {
			var expectedError error

			BeforeEach(func() {
				expectedError = errors.New("I am a CloudControllerClient Error")
				fakeCloudControllerClient.GetApplicationsReturns(
					[]resources.Application{
						{
							Name: "old-app-name",
							GUID: "old-app-guid",
						},
					},
					ccv3.Warnings{"get-app-warning"},
					nil)
				fakeCloudControllerClient.UpdateApplicationNameReturns(
					resources.Application{},
					ccv3.Warnings{"update-app-warning"},
					expectedError)
			})

			It("returns the warnings and the error", func() {
				_, warnings, err := actor.RenameApplicationByNameAndSpaceGUID("old-app-name", "new-app-name", "space-guid")
				Expect(warnings).To(ConsistOf("get-app-warning", "update-app-warning"))
				Expect(err).To(MatchError(expectedError))
			})
		})

		When("the app exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]resources.Application{
						{
							Name: "old-app-name",
							GUID: "old-app-guid",
						},
					},
					ccv3.Warnings{"get-app-warning"},
					nil,
				)

				fakeCloudControllerClient.UpdateApplicationNameReturns(
					resources.Application{
						Name: "new-app-name",
						GUID: "old-app-guid",
					},
					ccv3.Warnings{"update-app-warning"},
					nil,
				)
			})

			It("changes the app name and returns the application and warnings", func() {
				app, warnings, err := actor.RenameApplicationByNameAndSpaceGUID("old-app-name", "new-app-name", "some-space-guid")
				Expect(err).ToNot(HaveOccurred())
				Expect(app).To(Equal(resources.Application{
					Name: "new-app-name",
					GUID: "old-app-guid",
				}))
				Expect(warnings).To(ConsistOf("get-app-warning", "update-app-warning"))
				appName, appGuid := fakeCloudControllerClient.UpdateApplicationNameArgsForCall(0)
				Expect(appName).To(Equal("new-app-name"))
				Expect(appGuid).To(Equal("old-app-guid"))
			})
		})

	})
})
