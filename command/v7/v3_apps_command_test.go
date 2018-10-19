package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/shared/sharedfakes"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("v3-apps Command", func() {
	var (
		cmd             v7.V3AppsCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeV3AppsActor
		fakeV2Actor     *sharedfakes.FakeV2AppActor
		binaryName      string
		executeErr      error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeV3AppsActor)
		fakeV2Actor = new(sharedfakes.FakeV2AppActor)

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)

		cmd = v7.V3AppsCommand{
			UI:          testUI,
			Config:      fakeConfig,
			Actor:       fakeActor,
			V2AppActor:  fakeV2Actor,
			SharedActor: fakeSharedActor,
		}

		fakeConfig.TargetedOrganizationReturns(configv3.Organization{
			Name: "some-org",
			GUID: "some-org-guid",
		})
		fakeConfig.TargetedSpaceReturns(configv3.Space{
			Name: "some-space",
			GUID: "some-space-guid",
		})

		fakeConfig.CurrentUserReturns(configv3.User{Name: "steve"}, nil)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	When("the API version is below the minimum", func() {
		BeforeEach(func() {
			fakeActor.CloudControllerAPIVersionReturns(ccversion.MinV3ClientVersion)
		})

		It("returns a MinimumAPIVersionNotMetError", func() {
			Expect(executeErr).To(MatchError(translatableerror.MinimumCFAPIVersionNotMetError{
				CurrentVersion: ccversion.MinV3ClientVersion,
				MinimumVersion: ccversion.MinVersionApplicationFlowV3,
			}))
		})

		It("displays the experimental warning", func() {
			Expect(testUI.Err).To(Say("This command is in EXPERIMENTAL stage and may change without notice"))
		})
	})

	When("checking target fails", func() {
		BeforeEach(func() {
			fakeActor.CloudControllerAPIVersionReturns(ccversion.MinVersionApplicationFlowV3)
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
			fakeActor.CloudControllerAPIVersionReturns(ccversion.MinVersionApplicationFlowV3)
			expectedErr = errors.New("some current user error")
			fakeConfig.CurrentUserReturns(configv3.User{}, expectedErr)
		})

		It("return an error", func() {
			Expect(executeErr).To(Equal(expectedErr))
		})
	})

	When("getting the applications returns an error", func() {
		var expectedErr error

		BeforeEach(func() {
			fakeActor.CloudControllerAPIVersionReturns(ccversion.MinVersionApplicationFlowV3)
			expectedErr = ccerror.RequestError{}
			fakeActor.GetApplicationsWithProcessesBySpaceReturns([]v7action.ApplicationWithProcessSummary{}, v7action.Warnings{"warning-1", "warning-2"}, expectedErr)
		})

		It("returns the error and prints warnings", func() {
			Expect(executeErr).To(Equal(ccerror.RequestError{}))

			Expect(testUI.Out).To(Say("Getting apps in org some-org / space some-space as steve\\.\\.\\."))

			Expect(testUI.Err).To(Say("warning-1"))
			Expect(testUI.Err).To(Say("warning-2"))
		})
	})

	When("getting routes returns an error", func() {
		var expectedErr error

		BeforeEach(func() {
			fakeActor.CloudControllerAPIVersionReturns(ccversion.MinVersionApplicationFlowV3)
			expectedErr = ccerror.RequestError{}
			fakeActor.GetApplicationsWithProcessesBySpaceReturns([]v7action.ApplicationWithProcessSummary{
				{
					Application: v7action.Application{
						GUID:  "app-guid",
						Name:  "some-app",
						State: constant.ApplicationStarted,
					},
					ProcessSummaries: []v7action.ProcessSummary{{Process: v7action.Process{Type: "process-type"}}},
				},
			}, v7action.Warnings{"warning-1", "warning-2"}, nil)

			fakeV2Actor.GetApplicationRoutesReturns([]v2action.Route{}, v2action.Warnings{"route-warning-1", "route-warning-2"}, expectedErr)
		})

		It("returns the error and prints warnings", func() {
			Expect(executeErr).To(Equal(ccerror.RequestError{}))

			Expect(testUI.Out).To(Say("Getting apps in org some-org / space some-space as steve\\.\\.\\."))

			Expect(testUI.Err).To(Say("warning-1"))
			Expect(testUI.Err).To(Say("warning-2"))
			Expect(testUI.Err).To(Say("route-warning-1"))
			Expect(testUI.Err).To(Say("route-warning-2"))
		})
	})

	When("the route actor does not return any errors", func() {
		BeforeEach(func() {
			fakeActor.CloudControllerAPIVersionReturns(ccversion.MinVersionApplicationFlowV3)
			fakeV2Actor.GetApplicationRoutesStub = func(appGUID string) (v2action.Routes, v2action.Warnings, error) {
				switch appGUID {
				case "app-guid-1":
					return []v2action.Route{
							{
								Host:   "some-app-1",
								Domain: v2action.Domain{Name: "some-other-domain"},
							},
							{
								Host:   "some-app-1",
								Domain: v2action.Domain{Name: "some-domain"},
							},
						},
						v2action.Warnings{"route-warning-1", "route-warning-2"},
						nil
				case "app-guid-2":
					return []v2action.Route{
							{
								Host:   "some-app-2",
								Domain: v2action.Domain{Name: "some-domain"},
							},
						},
						v2action.Warnings{"route-warning-3", "route-warning-4"},
						nil
				default:
					panic("unknown app guid")
				}
			}
		})

		Context("with existing apps", func() {
			BeforeEach(func() {
				appSummaries := []v7action.ApplicationWithProcessSummary{
					{
						Application: v7action.Application{
							GUID:  "app-guid-1",
							Name:  "some-app-1",
							State: constant.ApplicationStarted,
						},
						ProcessSummaries: []v7action.ProcessSummary{
							{
								Process: v7action.Process{
									Type: "console",
								},
								InstanceDetails: []v7action.ProcessInstance{},
							},
							{
								Process: v7action.Process{
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
								Process: v7action.Process{
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
					},
					{
						Application: v7action.Application{
							GUID:  "app-guid-2",
							Name:  "some-app-2",
							State: constant.ApplicationStopped,
						},
						ProcessSummaries: []v7action.ProcessSummary{
							{
								Process: v7action.Process{
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
					},
				}
				fakeActor.GetApplicationsWithProcessesBySpaceReturns(appSummaries, v7action.Warnings{"warning-1", "warning-2"}, nil)
			})

			It("prints the application summary and outputs warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(testUI.Out).To(Say("Getting apps in org some-org / space some-space as steve\\.\\.\\."))

				Expect(testUI.Out).To(Say("name\\s+requested state\\s+processes\\s+routes"))
				Expect(testUI.Out).To(Say("some-app-1\\s+started\\s+web:2/2, console:0/0, worker:0/1\\s+some-app-1.some-other-domain, some-app-1.some-domain"))
				Expect(testUI.Out).To(Say("some-app-2\\s+stopped\\s+web:0/2\\s+some-app-2.some-domain"))

				Expect(testUI.Err).To(Say("warning-1"))
				Expect(testUI.Err).To(Say("warning-2"))

				Expect(testUI.Err).To(Say("route-warning-1"))
				Expect(testUI.Err).To(Say("route-warning-2"))
				Expect(testUI.Err).To(Say("route-warning-3"))
				Expect(testUI.Err).To(Say("route-warning-4"))

				Expect(fakeActor.GetApplicationsWithProcessesBySpaceCallCount()).To(Equal(1))
				spaceGUID := fakeActor.GetApplicationsWithProcessesBySpaceArgsForCall(0)
				Expect(spaceGUID).To(Equal("some-space-guid"))

				Expect(fakeV2Actor.GetApplicationRoutesCallCount()).To(Equal(2))
				appGUID := fakeV2Actor.GetApplicationRoutesArgsForCall(0)
				Expect(appGUID).To(Equal("app-guid-1"))
				appGUID = fakeV2Actor.GetApplicationRoutesArgsForCall(1)
				Expect(appGUID).To(Equal("app-guid-2"))
			})
		})

		When("app does not have processes", func() {
			BeforeEach(func() {
				appSummaries := []v7action.ApplicationWithProcessSummary{
					{
						Application: v7action.Application{
							GUID:  "app-guid",
							Name:  "some-app",
							State: constant.ApplicationStarted,
						},
						ProcessSummaries: []v7action.ProcessSummary{},
					},
				}
				fakeActor.GetApplicationsWithProcessesBySpaceReturns(appSummaries, v7action.Warnings{"warning"}, nil)
			})

			It("it does not request or display routes information for app", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(testUI.Out).To(Say("Getting apps in org some-org / space some-space as steve\\.\\.\\."))

				Expect(testUI.Out).To(Say("name\\s+requested state\\s+processes\\s+routes"))
				Expect(testUI.Out).To(Say("some-app\\s+started\\s+$"))
				Expect(testUI.Err).To(Say("warning"))

				Expect(fakeActor.GetApplicationsWithProcessesBySpaceCallCount()).To(Equal(1))
				spaceGUID := fakeActor.GetApplicationsWithProcessesBySpaceArgsForCall(0)
				Expect(spaceGUID).To(Equal("some-space-guid"))

				Expect(fakeV2Actor.GetApplicationRoutesCallCount()).To(Equal(0))
			})
		})

		Context("with no apps", func() {
			BeforeEach(func() {
				fakeActor.GetApplicationsWithProcessesBySpaceReturns([]v7action.ApplicationWithProcessSummary{}, v7action.Warnings{"warning-1", "warning-2"}, nil)
			})

			It("displays there are no apps", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(testUI.Out).To(Say("Getting apps in org some-org / space some-space as steve\\.\\.\\."))
				Expect(testUI.Out).To(Say("No apps found"))
			})
		})
	})
})
