package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/v8/actor/actionerror"
	"code.cloudfoundry.org/cli/v8/actor/v7action"
	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/v8/command/commandfakes"
	v7 "code.cloudfoundry.org/cli/v8/command/v7"
	"code.cloudfoundry.org/cli/v8/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/v8/resources"
	"code.cloudfoundry.org/cli/v8/util/configv3"
	"code.cloudfoundry.org/cli/v8/util/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("apps Command", func() {
	var (
		cmd             v7.AppsCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		binaryName      string
		executeErr      error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)

		cmd = v7.AppsCommand{
			BaseCommand: v7.BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				Actor:       fakeActor,
				SharedActor: fakeSharedActor,
			},
		}

		fakeConfig.TargetedOrganizationReturns(configv3.Organization{
			Name: "some-org",
			GUID: "some-org-guid",
		})
		fakeConfig.TargetedSpaceReturns(configv3.Space{
			Name: "some-space",
			GUID: "some-space-guid",
		})

		fakeActor.GetCurrentUserReturns(configv3.User{Name: "steve"}, nil)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	It("displays the experimental warning", func() {
		Expect(testUI.Err).NotTo(Say("This command is in EXPERIMENTAL stage and may change without notice"))
	})

	When("checking target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(actionerror.NoOrganizationTargetedError{BinaryName: binaryName})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(actionerror.NoOrganizationTargetedError{BinaryName: binaryName}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrg).To(BeTrue())
			Expect(checkTargetedSpace).To(BeTrue())
		})
	})

	When("the user is not logged in", func() {
		var expectedErr error

		BeforeEach(func() {
			expectedErr = errors.New("some current user error")
			fakeActor.GetCurrentUserReturns(configv3.User{}, expectedErr)
		})

		It("return an error", func() {
			Expect(executeErr).To(Equal(expectedErr))
		})
	})

	When("getting the applications returns an error", func() {
		var expectedErr error

		BeforeEach(func() {
			expectedErr = ccerror.RequestError{}
			fakeActor.GetAppSummariesForSpaceReturns([]v7action.ApplicationSummary{}, v7action.Warnings{"warning-1", "warning-2"}, expectedErr)
		})

		It("returns the error and prints warnings", func() {
			Expect(executeErr).To(Equal(ccerror.RequestError{}))

			Expect(testUI.Out).To(Say(`Getting apps in org some-org / space some-space as steve\.\.\.`))

			Expect(testUI.Err).To(Say("warning-1"))
			Expect(testUI.Err).To(Say("warning-2"))
		})
	})

	When("getting routes returns an error", func() {
		var expectedErr error

		BeforeEach(func() {
			expectedErr = ccerror.RequestError{}
			fakeActor.GetAppSummariesForSpaceReturns([]v7action.ApplicationSummary{
				{
					Application: resources.Application{
						GUID:  "app-guid",
						Name:  "some-app",
						State: constant.ApplicationStarted,
					},
					ProcessSummaries: []v7action.ProcessSummary{{Process: resources.Process{Type: "process-type"}}},
					Routes:           []resources.Route{},
				},
			}, v7action.Warnings{"warning-1", "warning-2"}, expectedErr)
		})

		It("returns the error and prints warnings", func() {
			Expect(executeErr).To(Equal(ccerror.RequestError{}))

			Expect(testUI.Out).To(Say(`Getting apps in org some-org / space some-space as steve\.\.\.`))

			Expect(testUI.Err).To(Say("warning-1"))
			Expect(testUI.Err).To(Say("warning-2"))
		})
	})

	When("the route actor does not return any errors", func() {
		Context("with existing apps", func() {
			BeforeEach(func() {
				lbLCVal := "least-connection"
				lbLeastConnections := &lbLCVal

				appSummaries := []v7action.ApplicationSummary{
					{
						Application: resources.Application{
							GUID:  "app-guid-1",
							Name:  "some-app-1",
							State: constant.ApplicationStarted,
						},
						ProcessSummaries: []v7action.ProcessSummary{
							{
								Process: resources.Process{
									Type: "console",
								},
								InstanceDetails: []v7action.ProcessInstance{},
							},
							{
								Process: resources.Process{
									Type: "worker",
								},
								InstanceDetails: []v7action.ProcessInstance{
									{
										Index: 0,
										State: constant.ProcessInstanceDown,
									},
								},
							},
							{
								Process: resources.Process{
									Type: constant.ProcessTypeWeb,
								},
								InstanceDetails: []v7action.ProcessInstance{
									v7action.ProcessInstance{
										Index: 0,
										State: constant.ProcessInstanceRunning,
									},
									v7action.ProcessInstance{
										Index: 1,
										State: constant.ProcessInstanceRunning,
									},
								},
							},
						},
						Routes: []resources.Route{
							{
								Host:    "some-app-1",
								URL:     "some-app-1.some-other-domain",
								Options: map[string]*string{"loadbalancing": lbLeastConnections},
							},
							{
								Host: "some-app-1",
								URL:  "some-app-1.some-domain",
							},
						},
					},
					{
						Application: resources.Application{
							GUID:  "app-guid-2",
							Name:  "some-app-2",
							State: constant.ApplicationStopped,
						},
						ProcessSummaries: []v7action.ProcessSummary{
							{
								Process: resources.Process{
									Type: constant.ProcessTypeWeb,
								},
								InstanceDetails: []v7action.ProcessInstance{
									v7action.ProcessInstance{
										Index: 0,
										State: constant.ProcessInstanceDown,
									},
									v7action.ProcessInstance{
										Index: 1,
										State: constant.ProcessInstanceDown,
									},
								},
							},
						},
						Routes: []resources.Route{
							{
								Host: "some-app-2",
								URL:  "some-app-2.some-domain",
							},
						},
					},
				}
				fakeActor.GetAppSummariesForSpaceReturns(appSummaries, v7action.Warnings{"warning-1", "warning-2"}, nil)
			})

			It("prints the application summary and outputs warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(testUI.Out).To(Say(`Getting apps in org some-org / space some-space as steve\.\.\.`))

				Expect(testUI.Out).To(Say(`name\s+requested state\s+processes\s+routes`))
				Expect(testUI.Out).To(Say(`some-app-1\s+started\s+web:2/2, console:0/0, worker:0/1\s+some-app-1.some-other-domain, some-app-1.some-domain`))

				Expect(testUI.Err).To(Say("warning-1"))
				Expect(testUI.Err).To(Say("warning-2"))

				Expect(fakeActor.GetAppSummariesForSpaceCallCount()).To(Equal(1))
				spaceGUID, labels, omitStats := fakeActor.GetAppSummariesForSpaceArgsForCall(0)
				Expect(spaceGUID).To(Equal("some-space-guid"))
				Expect(labels).To(Equal(""))
				Expect(omitStats).To(Equal(false))
			})
		})

		When("app does not have processes", func() {
			BeforeEach(func() {
				appSummaries := []v7action.ApplicationSummary{
					{
						Application: resources.Application{
							GUID:  "app-guid",
							Name:  "some-app",
							State: constant.ApplicationStarted,
						},
						ProcessSummaries: []v7action.ProcessSummary{},
					},
				}
				fakeActor.GetAppSummariesForSpaceReturns(appSummaries, v7action.Warnings{"warning"}, nil)
			})

			It("it does not request or display routes information for app", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(testUI.Out).To(Say(`Getting apps in org some-org / space some-space as steve\.\.\.`))

				Expect(testUI.Out).To(Say(`name\s+requested state\s+processes\s+routes`))
				Expect(testUI.Out).To(Say(`some-app\s+started\s+$`))
				Expect(testUI.Err).To(Say("warning"))

				Expect(fakeActor.GetAppSummariesForSpaceCallCount()).To(Equal(1))
				spaceGUID, labelSelector, omitStats := fakeActor.GetAppSummariesForSpaceArgsForCall(0)
				Expect(spaceGUID).To(Equal("some-space-guid"))
				Expect(labelSelector).To(Equal(""))
				Expect(omitStats).To(Equal(false))
			})
		})

		Context("with no apps", func() {
			BeforeEach(func() {
				fakeActor.GetAppSummariesForSpaceReturns([]v7action.ApplicationSummary{}, v7action.Warnings{"warning-1", "warning-2"}, nil)
			})

			It("displays there are no apps", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(testUI.Out).To(Say(`Getting apps in org some-org / space some-space as steve\.\.\.`))
				Expect(testUI.Out).To(Say("No apps found"))
			})

		})
	})
	Context("when a labels flag is set", func() {
		BeforeEach(func() {
			cmd.Labels = "fish=moose"
		})

		It("passes the flag to the API", func() {
			Expect(fakeActor.GetAppSummariesForSpaceCallCount()).To(Equal(1))
			_, labelSelector, _ := fakeActor.GetAppSummariesForSpaceArgsForCall(0)
			Expect(labelSelector).To(Equal("fish=moose"))
		})
	})

	Context("when '--skip-stats' flag is set", func() {
		BeforeEach(func() {
			cmd.OmitStats = true

			appSummaries := []v7action.ApplicationSummary{
				{
					Application: resources.Application{
						GUID:  "app-guid-1",
						Name:  "some-app-1",
						State: constant.ApplicationStarted,
					},
					ProcessSummaries: []v7action.ProcessSummary{},
					Routes: []resources.Route{
						{
							Host: "some-app-1",
							URL:  "some-app-1.some-domain",
						},
					},
				},
			}
			fakeActor.GetAppSummariesForSpaceReturns(appSummaries, v7action.Warnings{}, nil)

		})

		It("prints the application summary without process information", func() {
			Expect(executeErr).ToNot(HaveOccurred())

			Expect(testUI.Out).To(Say(`Getting apps in org some-org / space some-space as steve\.\.\.`))

			Expect(testUI.Out).To(Say(`name\s+requested state\s+routes`))
			Expect(testUI.Out).To(Say(`some-app-1\s+started\s+some-app-1.some-domain`))
			Expect(fakeActor.GetAppSummariesForSpaceCallCount()).To(Equal(1))
			_, _, omitStats := fakeActor.GetAppSummariesForSpaceArgsForCall(0)
			Expect(omitStats).To(Equal(true))
		})
	})

})
