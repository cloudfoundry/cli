package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("domains Command", func() {
	var (
		cmd             DomainsCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeDomainsActor
		executeErr      error
		args            []string
		binaryName      string
	)

	const tableHeaders = `name\s+availability\s+internal`

	JustBeforeEach(func() {
		executeErr = cmd.Execute(args)
	})

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeDomainsActor)
		args = nil

		cmd = DomainsCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
	})

	When("the environment is not setup correctly", func() {
		When("checking target fails", func() {
			BeforeEach(func() {
				fakeSharedActor.CheckTargetReturns(actionerror.NotLoggedInError{BinaryName: binaryName})
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError(actionerror.NotLoggedInError{BinaryName: binaryName}))

				Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
				checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
				Expect(checkTargetedOrg).To(BeTrue())
				Expect(checkTargetedSpace).To(BeFalse())
			})
		})

		When("when there is no org targeted", func() {
			BeforeEach(func() {
				fakeSharedActor.CheckTargetReturns(actionerror.NoOrganizationTargetedError{BinaryName: binaryName})
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError(actionerror.NoOrganizationTargetedError{BinaryName: binaryName}))
				checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
				Expect(checkTargetedOrg).To(BeTrue())
				Expect(checkTargetedSpace).To(BeFalse())
			})
		})
	})

	Context("When the environment is setup correctly", func() {
		BeforeEach(func() {
			fakeConfig.CurrentUserReturns(configv3.User{Name: "banana"}, nil)
		})

		When("DomainsActor returns an error", func() {
			var expectedErr error

			BeforeEach(func() {
				warnings := v7action.Warnings{"warning-1", "warning-2"}
				expectedErr = errors.New("some-error")
				fakeActor.GetOrganizationDomainsReturns(nil, warnings, expectedErr)
			})

			It("prints that error with warnings", func() {
				Expect(executeErr).To(Equal(expectedErr))

				Expect(testUI.Err).To(Say("warning-1"))
				Expect(testUI.Err).To(Say("warning-2"))
				Expect(testUI.Out).ToNot(Say(tableHeaders))
			})
		})

		When("GetDomains returns some domains", func() {
			var domains []resources.Domain

			BeforeEach(func() {
				domains = []resources.Domain{
					{Name: "domain1", GUID: "domain-guid-1", Internal: types.NullBool{IsSet: true, Value: true}},
					{Name: "domain3", GUID: "domain-guid-3", Internal: types.NullBool{IsSet: false, Value: false}, OrganizationGUID: "owning-org-guid"},
					{Name: "domain2", GUID: "domain-guid-2", Internal: types.NullBool{IsSet: true, Value: false}},
				}

				fakeActor.GetOrganizationDomainsReturns(
					domains,
					v7action.Warnings{"actor-warning-1", "actor-warning-2", "actor-warning-3"},
					nil,
				)

				fakeConfig.TargetedOrganizationReturns(configv3.Organization{
					GUID: "some-org-guid",
					Name: "some-org",
				})
			})

			It("asks the DomainsActor for a list of domains", func() {
				Expect(fakeActor.GetOrganizationDomainsCallCount()).To(Equal(1))
			})

			It("prints warnings", func() {
				Expect(testUI.Err).To(Say("actor-warning-1"))
				Expect(testUI.Err).To(Say("actor-warning-2"))
				Expect(testUI.Err).To(Say("actor-warning-3"))
			})

			It("prints the list of domains in alphabetical order", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(testUI.Out).To(Say(tableHeaders))
				Expect(testUI.Out).To(Say(`domain1\s+shared\s+true`))
				Expect(testUI.Out).To(Say(`domain2\s+shared`))
				Expect(testUI.Out).To(Say(`domain3\s+private`))
			})

			It("prints the flavor text", func() {
				Expect(testUI.Out).To(Say("Getting domains in org some-org as banana...\n\n"))
			})
		})

		When("GetDomains returns no domains", func() {
			var domains []resources.Domain

			BeforeEach(func() {
				domains = []resources.Domain{}

				fakeActor.GetOrganizationDomainsReturns(
					domains,
					v7action.Warnings{"actor-warning-1", "actor-warning-2", "actor-warning-3"},
					nil,
				)

				fakeConfig.TargetedOrganizationReturns(configv3.Organization{
					GUID: "some-org-guid",
					Name: "some-org",
				})
			})

			It("asks the DomainsActor for a list of domains", func() {
				Expect(fakeActor.GetOrganizationDomainsCallCount()).To(Equal(1))
			})

			It("prints warnings", func() {
				Expect(testUI.Err).To(Say("actor-warning-1"))
				Expect(testUI.Err).To(Say("actor-warning-2"))
				Expect(testUI.Err).To(Say("actor-warning-3"))
			})

			It("does not print table headers", func() {
				Expect(testUI.Out).NotTo(Say(tableHeaders))
			})

			It("prints a message indicating that no domains were found", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(testUI.Out).To(Say("No domains found."))
			})

			It("prints the flavor text", func() {
				Expect(testUI.Out).To(Say("Getting domains in org some-org as banana...\n\n"))
			})
		})
		Context("when a labels flag is set", func() {
			BeforeEach(func() {
				cmd.Labels = "fish=moose"
			})

			It("passes the flag to the API", func() {
				Expect(fakeActor.GetOrganizationDomainsCallCount()).To(Equal(1))
				_, labelSelector := fakeActor.GetOrganizationDomainsArgsForCall(0)
				Expect(labelSelector).To(Equal("fish=moose"))
			})
		})
	})
})
