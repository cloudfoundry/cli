package v2v3action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/v2action"
	. "code.cloudfoundry.org/cli/actor/v2v3action"
	"code.cloudfoundry.org/cli/actor/v2v3action/v2v3actionfakes"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Application Summary Actions", func() {
	var (
		actor       *Actor
		fakeV2Actor *v2v3actionfakes.FakeV2Actor
		fakeV3Actor *v2v3actionfakes.FakeV3Actor
	)

	BeforeEach(func() {
		fakeV2Actor = new(v2v3actionfakes.FakeV2Actor)
		fakeV3Actor = new(v2v3actionfakes.FakeV3Actor)
		actor = NewActor(fakeV2Actor, fakeV3Actor)
	})

	Describe("ApplicationSummary", func() {
		DescribeTable("GetIsolationSegmentName",
			func(summary ApplicationSummary, isoName string, exists bool) {
				name, ok := summary.GetIsolationSegmentName()
				Expect(ok).To(Equal(exists))
				Expect(name).To(Equal(isoName))
			},

			Entry("when the there are application instances and the isolationSegmentName is set",
				ApplicationSummary{
					ApplicationInstanceWithStats: []v2action.ApplicationInstanceWithStats{{IsolationSegment: "some-name"}},
				},
				"some-name",
				true,
			),

			Entry("when the there are application instances and the isolationSegmentName is blank",
				ApplicationSummary{
					ApplicationInstanceWithStats: []v2action.ApplicationInstanceWithStats{{}},
				},
				"",
				false,
			),

			Entry("when the there are no application instances", ApplicationSummary{}, "", false),
		)
	})

	Describe("GetApplicationSummaryByNameAndSpace", func() {
		var (
			appName              string
			spaceGUID            string
			withObfuscatedValues bool

			summary    ApplicationSummary
			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			appName = "some-app-name"
			spaceGUID = "some-space-guid"
			withObfuscatedValues = true
		})

		JustBeforeEach(func() {
			summary, warnings, executeErr = actor.GetApplicationSummaryByNameAndSpace(appName, spaceGUID, withObfuscatedValues)
		})

		Context("when getting the V3 Application Summary is successful", func() {
			Context("regardless of the application state", func() {
				BeforeEach(func() {
					v3Summary := v3action.ApplicationSummary{
						Application: v3action.Application{
							GUID: "some-app-guid",
						},
						ProcessSummaries: v3action.ProcessSummaries{
							{Process: v3action.Process{Type: "console"}},
							{Process: v3action.Process{Type: constant.ProcessTypeWeb}},
						},
					}
					fakeV3Actor.GetApplicationSummaryByNameAndSpaceReturns(v3Summary, v3action.Warnings{"v3-summary-warning"}, nil)
				})

				It("returns the v3 application summary with sorted processes and warnings", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("v3-summary-warning"))
					Expect(summary.ApplicationSummary).To(Equal(v3action.ApplicationSummary{
						Application: v3action.Application{
							GUID: "some-app-guid",
						},
						ProcessSummaries: v3action.ProcessSummaries{
							{Process: v3action.Process{Type: constant.ProcessTypeWeb}},
							{Process: v3action.Process{Type: "console"}},
						},
					}))

					Expect(fakeV3Actor.GetApplicationSummaryByNameAndSpaceCallCount()).To(Equal(1))
					passedAppName, passedSpaceGUID, passedWithObfuscatedValues := fakeV3Actor.GetApplicationSummaryByNameAndSpaceArgsForCall(0)
					Expect(passedAppName).To(Equal(appName))
					Expect(passedSpaceGUID).To(Equal(spaceGUID))
					Expect(passedWithObfuscatedValues).To(Equal(withObfuscatedValues))
				})

				Context("when getting the routes is successful", func() {
					BeforeEach(func() {
						routes := v2action.Routes{
							{GUID: "some-route-guid"},
							{GUID: "some-other-route-guid"},
						}

						fakeV2Actor.GetApplicationRoutesReturns(routes, v2action.Warnings{"v2-routes-warnings"}, nil)
					})

					It("adds the routes to the summary", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(warnings).To(ConsistOf("v2-routes-warnings", "v3-summary-warning"))

						Expect(summary.Routes).To(Equal(v2action.Routes{
							{GUID: "some-route-guid"},
							{GUID: "some-other-route-guid"},
						}))

						Expect(fakeV2Actor.GetApplicationRoutesCallCount()).To(Equal(1))
						Expect(fakeV2Actor.GetApplicationRoutesArgsForCall(0)).To(Equal("some-app-guid"))
					})
				})

				Context("when getting the application routes errors", func() {
					Context("when a generic error is returned", func() {
						BeforeEach(func() {
							fakeV2Actor.GetApplicationRoutesReturns(nil, v2action.Warnings{"v2-routes-warnings"}, errors.New("some-error"))
						})

						It("returns warnings and the error", func() {
							Expect(executeErr).To(MatchError("some-error"))
							Expect(warnings).To(ConsistOf("v2-routes-warnings", "v3-summary-warning"))
						})
					})

					Context("when a ResourceNotFoundError is returned", func() {
						BeforeEach(func() {
							fakeV2Actor.GetApplicationRoutesReturns(nil, v2action.Warnings{"v2-routes-warnings"}, ccerror.ResourceNotFoundError{})
						})

						It("adds warnings and continues", func() {
							Expect(executeErr).ToNot(HaveOccurred())
							Expect(warnings).To(ConsistOf("v2-routes-warnings", "v3-summary-warning"))
						})
					})
				})
			})

			Context("when the application is running", func() {
				BeforeEach(func() {
					v3Summary := v3action.ApplicationSummary{
						Application: v3action.Application{
							GUID:  "some-app-guid",
							State: constant.ApplicationStarted,
						},
					}
					fakeV3Actor.GetApplicationSummaryByNameAndSpaceReturns(v3Summary, v3action.Warnings{"v3-summary-warning"}, nil)
				})

				Context("when getting the application instances with stats is successful", func() {
					BeforeEach(func() {
						stats := []v2action.ApplicationInstanceWithStats{{ID: 0}, {ID: 1}}
						fakeV2Actor.GetApplicationInstancesWithStatsByApplicationReturns(stats, v2action.Warnings{"v2-app-instances-warning"}, nil)
					})

					It("returns the application summary with application instances with stats", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(warnings).To(ConsistOf("v2-app-instances-warning", "v3-summary-warning"))
						Expect(summary.ApplicationInstanceWithStats).To(Equal([]v2action.ApplicationInstanceWithStats{{ID: 0}, {ID: 1}}))

						Expect(fakeV2Actor.GetApplicationInstancesWithStatsByApplicationCallCount()).To(Equal(1))
					})
				})

				Context("when getting the application instances with stats returns an error", func() {
					Context("when a generic error is returned", func() {
						BeforeEach(func() {
							fakeV2Actor.GetApplicationInstancesWithStatsByApplicationReturns(nil, v2action.Warnings{"v2-app-instances-warning"}, errors.New("boom"))
						})

						It("returns error and warnings", func() {
							Expect(executeErr).To(MatchError("boom"))
							Expect(warnings).To(ConsistOf("v2-app-instances-warning", "v3-summary-warning"))
						})
					})

					Context("when a ResourceNotFoundError is returned", func() {
						BeforeEach(func() {
							fakeV2Actor.GetApplicationInstancesWithStatsByApplicationReturns(nil, v2action.Warnings{"v2-app-instances-warning"}, ccerror.ResourceNotFoundError{})
						})

						It("adds warnings and continues", func() {
							Expect(executeErr).ToNot(HaveOccurred())
							Expect(warnings).To(ConsistOf("v2-app-instances-warning", "v3-summary-warning"))
						})
					})
				})
			})

			Context("when the application is stopped", func() {
				BeforeEach(func() {
					v3Summary := v3action.ApplicationSummary{
						Application: v3action.Application{
							GUID:  "some-app-guid",
							State: constant.ApplicationStopped,
						},
					}
					fakeV3Actor.GetApplicationSummaryByNameAndSpaceReturns(v3Summary, v3action.Warnings{"v3-summary-warning"}, nil)
				})

				It("does not get application instances with stats", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("v3-summary-warning"))

					Expect(fakeV2Actor.GetApplicationInstancesWithStatsByApplicationCallCount()).To(Equal(0))
				})
			})
		})

		Context("when getting the V3 Application Summary returns an error", func() {
			BeforeEach(func() {
				fakeV3Actor.GetApplicationSummaryByNameAndSpaceReturns(v3action.ApplicationSummary{}, v3action.Warnings{"v3-summary-warning"}, errors.New("CRAZY!"))
			})

			It("returns back error and warnings", func() {
				Expect(executeErr).To(MatchError("CRAZY!"))
				Expect(warnings).To(ConsistOf("v3-summary-warning"))
			})
		})
	})
})
