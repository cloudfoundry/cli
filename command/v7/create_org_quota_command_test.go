package v7_test

import (
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	v7 "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("create-org-quota Command", func() {
	var (
		cmd             v7.CreateOrgQuotaCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeCreateOrgQuotaActor
		orgQuotaName    string
		executeErr      error
		//input           *Buffer
	)

	BeforeEach(func() {
		// input = NewBuffer()
		// testUI = ui.NewTestUIinput, NewBuffer(), NewBuffer()
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeCreateOrgQuotaActor)

		cmd = v7.CreateOrgQuotaCommand{}

		// cmd.Args.Username = "some-user"

	})

	JustBeforeEach(func() {
		cmd = v7.CreateOrgQuotaCommand{
			UI:           testUI,
			Config:       fakeConfig,
			SharedActor:  fakeSharedActor,
			Actor:        fakeActor,
			RequiredArgs: flag.OrganizationQuota{OrganizationQuota: orgQuotaName},
		}
		executeErr = cmd.Execute(nil)
	})

	When("the org quota is created successfully", func() {
		// before each to make sure the env is set up
		BeforeEach(func() {
			fakeActor.CreateOrganizationQuotaReturns(
				v7action.OrganizationQuota{GUID: "new-org-quota-guid", Name: "new-org-quota"},
				v7action.Warnings{"warning"},
				nil)
			orgQuotaName = "new-org-quota"
		})
		It("creates a quota with a given name", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(fakeActor.CreateOrganizationQuotaCallCount()).To(Equal(1))

			//quota := fakeActor.CreateOrgQuotaArgsForCall(0)
			//Expect(quota).To(Equal("my-quota")) // shuold this be the guid?
			Expect(testUI.Out).To(Say("Creating quota <quota name> as <user> "))
			Expect(testUI.Out).To(Say("OK"))
		})
	})
})
