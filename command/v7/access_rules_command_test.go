package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/v9/actor/actionerror"
	"code.cloudfoundry.org/cli/v9/actor/v7action"
	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/v9/command/commandfakes"
	v7 "code.cloudfoundry.org/cli/v9/command/v7"
	"code.cloudfoundry.org/cli/v9/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/v9/resources"
	"code.cloudfoundry.org/cli/v9/util/configv3"
	"code.cloudfoundry.org/cli/v9/util/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("access-rules Command", func() {
	var (
		cmd             v7.AccessRulesCommand
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

		cmd = v7.AccessRulesCommand{
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

		It("returns an error", func() {
			Expect(executeErr).To(Equal(expectedErr))
		})
	})

	When("getting access rules returns an error", func() {
		var expectedErr error

		BeforeEach(func() {
			expectedErr = ccerror.RequestError{}
			fakeActor.GetAccessRulesForSpaceReturns([]v7action.AccessRuleWithRoute{}, v7action.Warnings{"warning-1", "warning-2"}, expectedErr)
		})

		It("returns the error and prints warnings", func() {
			Expect(executeErr).To(Equal(ccerror.RequestError{}))

			Expect(testUI.Out).To(Say(`Getting access rules in org some-org / space some-space as steve\.\.\.`))

			Expect(testUI.Err).To(Say("warning-1"))
			Expect(testUI.Err).To(Say("warning-2"))
		})
	})

	When("getting access rules succeeds", func() {
		BeforeEach(func() {
			fakeActor.GetAccessRulesForSpaceReturns([]v7action.AccessRuleWithRoute{
				{
					AccessRule: resources.AccessRule{
						GUID:     "rule-guid-1",
						Selector: "cf:app:app-guid-1",
					},
					Route: resources.Route{
						GUID: "route-guid-1",
						Host: "myapp",
						Path: "/api",
					},
					DomainName: "example.com",
					ScopeType:  "app",
					SourceName: "my-app",
				},
				{
					AccessRule: resources.AccessRule{
						GUID:     "rule-guid-2",
						Selector: "cf:any",
					},
					Route: resources.Route{
						GUID: "route-guid-2",
						Host: "webapp",
						Path: "",
					},
					DomainName: "test.com",
					ScopeType:  "any",
					SourceName: "",
				},
			}, v7action.Warnings{"warning-1"}, nil)
		})

		It("displays the access rules in a table", func() {
			Expect(executeErr).ToNot(HaveOccurred())

			Expect(testUI.Out).To(Say(`Getting access rules in org some-org / space some-space as steve\.\.\.`))
			Expect(testUI.Out).To(Say(`host\s+domain\s+path\s+selector\s+scope\s+source`))
			Expect(testUI.Out).To(Say(`myapp\s+example\.com\s+/api\s+cf:app:app-guid-1\s+app\s+my-app`))
			Expect(testUI.Out).To(Say(`webapp\s+test\.com\s+cf:any\s+any`))

			Expect(testUI.Err).To(Say("warning-1"))

			Expect(fakeActor.GetAccessRulesForSpaceCallCount()).To(Equal(1))
			spaceGUID, domainName, hostname, path, labelSelector := fakeActor.GetAccessRulesForSpaceArgsForCall(0)
			Expect(spaceGUID).To(Equal("some-space-guid"))
			Expect(domainName).To(Equal(""))
			Expect(hostname).To(Equal(""))
			Expect(path).To(Equal(""))
			Expect(labelSelector).To(Equal(""))
		})
	})

	When("no access rules exist", func() {
		BeforeEach(func() {
			fakeActor.GetAccessRulesForSpaceReturns([]v7action.AccessRuleWithRoute{}, v7action.Warnings{}, nil)
		})

		It("displays a message indicating no access rules found", func() {
			Expect(executeErr).ToNot(HaveOccurred())

			Expect(testUI.Out).To(Say(`Getting access rules in org some-org / space some-space as steve\.\.\.`))
			Expect(testUI.Out).To(Say(`No access rules found\.`))
		})
	})

	When("filtering by domain", func() {
		BeforeEach(func() {
			cmd.Domain = "example.com"

			fakeActor.GetAccessRulesForSpaceReturns([]v7action.AccessRuleWithRoute{
				{
					AccessRule: resources.AccessRule{
						GUID:     "rule-guid-1",
						Selector: "cf:any",
					},
					Route: resources.Route{
						GUID: "route-guid-1",
						Host: "myapp",
					},
					DomainName: "example.com",
					ScopeType:  "any",
					SourceName: "",
				},
			}, v7action.Warnings{}, nil)
		})

		It("passes the domain filter to the actor", func() {
			Expect(executeErr).ToNot(HaveOccurred())

			Expect(fakeActor.GetAccessRulesForSpaceCallCount()).To(Equal(1))
			spaceGUID, domainName, hostname, path, labelSelector := fakeActor.GetAccessRulesForSpaceArgsForCall(0)
			Expect(spaceGUID).To(Equal("some-space-guid"))
			Expect(domainName).To(Equal("example.com"))
			Expect(hostname).To(Equal(""))
			Expect(path).To(Equal(""))
			Expect(labelSelector).To(Equal(""))
		})
	})

	When("filtering by hostname", func() {
		BeforeEach(func() {
			cmd.Hostname = "myapp"

			fakeActor.GetAccessRulesForSpaceReturns([]v7action.AccessRuleWithRoute{}, v7action.Warnings{}, nil)
		})

		It("passes the hostname filter to the actor", func() {
			Expect(executeErr).ToNot(HaveOccurred())

			Expect(fakeActor.GetAccessRulesForSpaceCallCount()).To(Equal(1))
			_, _, hostname, _, _ := fakeActor.GetAccessRulesForSpaceArgsForCall(0)
			Expect(hostname).To(Equal("myapp"))
		})
	})

	When("filtering by path", func() {
		BeforeEach(func() {
			cmd.Path = "/api"

			fakeActor.GetAccessRulesForSpaceReturns([]v7action.AccessRuleWithRoute{}, v7action.Warnings{}, nil)
		})

		It("passes the path filter to the actor", func() {
			Expect(executeErr).ToNot(HaveOccurred())

			Expect(fakeActor.GetAccessRulesForSpaceCallCount()).To(Equal(1))
			_, _, _, path, _ := fakeActor.GetAccessRulesForSpaceArgsForCall(0)
			Expect(path).To(Equal("/api"))
		})
	})

	When("filtering by labels", func() {
		BeforeEach(func() {
			cmd.Labels = "env=production,tier=frontend"

			fakeActor.GetAccessRulesForSpaceReturns([]v7action.AccessRuleWithRoute{}, v7action.Warnings{}, nil)
		})

		It("passes the label selector to the actor", func() {
			Expect(executeErr).ToNot(HaveOccurred())

			Expect(fakeActor.GetAccessRulesForSpaceCallCount()).To(Equal(1))
			_, _, _, _, labelSelector := fakeActor.GetAccessRulesForSpaceArgsForCall(0)
			Expect(labelSelector).To(Equal("env=production,tier=frontend"))
		})
	})

	When("using multiple filters", func() {
		BeforeEach(func() {
			cmd.Domain = "example.com"
			cmd.Hostname = "myapp"
			cmd.Path = "/api"
			cmd.Labels = "env=production"

			fakeActor.GetAccessRulesForSpaceReturns([]v7action.AccessRuleWithRoute{
				{
					AccessRule: resources.AccessRule{
						GUID:     "rule-guid-1",
						Selector: "cf:app:app-guid-1",
					},
					Route: resources.Route{
						GUID: "route-guid-1",
						Host: "myapp",
						Path: "/api",
					},
					DomainName: "example.com",
					ScopeType:  "app",
					SourceName: "my-app",
				},
			}, v7action.Warnings{}, nil)
		})

		It("passes all filters to the actor", func() {
			Expect(executeErr).ToNot(HaveOccurred())

			Expect(fakeActor.GetAccessRulesForSpaceCallCount()).To(Equal(1))
			spaceGUID, domainName, hostname, path, labelSelector := fakeActor.GetAccessRulesForSpaceArgsForCall(0)
			Expect(spaceGUID).To(Equal("some-space-guid"))
			Expect(domainName).To(Equal("example.com"))
			Expect(hostname).To(Equal("myapp"))
			Expect(path).To(Equal("/api"))
			Expect(labelSelector).To(Equal("env=production"))
		})

		It("displays the filtered results", func() {
			Expect(executeErr).ToNot(HaveOccurred())

			Expect(testUI.Out).To(Say(`Getting access rules in org some-org / space some-space as steve\.\.\.`))
			Expect(testUI.Out).To(Say(`host\s+domain\s+path\s+selector\s+scope\s+source`))
			Expect(testUI.Out).To(Say(`myapp\s+example\.com\s+/api\s+cf:app:app-guid-1\s+app\s+my-app`))
		})
	})

	When("route formatting handles edge cases", func() {
		BeforeEach(func() {
			fakeActor.GetAccessRulesForSpaceReturns([]v7action.AccessRuleWithRoute{
				{
					AccessRule: resources.AccessRule{
						GUID:     "rule-guid-1",
						Selector: "cf:any",
					},
					Route: resources.Route{
						GUID: "route-guid-1",
						Host: "",
						Path: "/api",
					},
					DomainName: "example.com",
					ScopeType:  "any",
					SourceName: "",
				},
				{
					AccessRule: resources.AccessRule{
						GUID:     "rule-guid-2",
						Selector: "cf:any",
					},
					Route: resources.Route{
						GUID: "route-guid-2",
						Host: "myapp",
						Path: "",
					},
					DomainName: "test.com",
					ScopeType:  "any",
					SourceName: "",
				},
			}, v7action.Warnings{}, nil)
		})

		It("formats routes correctly", func() {
			Expect(executeErr).ToNot(HaveOccurred())

			// No host, with path: empty host, example.com, /api
			Expect(testUI.Out).To(Say(`\s+example\.com\s+/api`))

			// With host, no path: myapp, test.com, empty path
			Expect(testUI.Out).To(Say(`myapp\s+test\.com\s+`))
		})
	})
})
