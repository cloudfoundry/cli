package v7_test

import (
	"code.cloudfoundry.org/cli/v9/actor/v7action"
	"code.cloudfoundry.org/cli/v9/command/commandfakes"
	. "code.cloudfoundry.org/cli/v9/command/v7"
	"code.cloudfoundry.org/cli/v9/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/v9/resources"
	"code.cloudfoundry.org/cli/v9/util/configv3"
	"code.cloudfoundry.org/cli/v9/util/ui"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("service-accounts Command", func() {
	var cmd ServiceAccountsCommand
	var actor *v7fakes.FakeActor
	var shared *commandfakes.FakeSharedActor
	var testUI *ui.UI
	BeforeEach(func() {
		actor = new(v7fakes.FakeActor)
		shared = new(commandfakes.FakeSharedActor)
		config := new(commandfakes.FakeConfig)
		config.TargetedSpaceReturns(configv3.Space{GUID: "space-guid", Name: "workers"})
		config.TargetedOrganizationReturns(configv3.Organization{Name: "poc"})
		actor.GetCurrentUserReturns(configv3.User{Name: "operator"}, nil)
		actor.GetServiceAccountsInSpaceReturns([]resources.ServiceAccount{
			{Name: "first-account", ClientID: "cf:service-account:first-account", CertificateDNSSAN: "first-account.svc.identity", Enabled: true, Status: "ready"},
			{Name: "second-account", Enabled: false, Status: "unprovisioned"},
		}, v7action.Warnings{"list warning"}, nil)
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		cmd = ServiceAccountsCommand{BaseCommand: BaseCommand{Actor: actor, SharedActor: shared, Config: config, UI: testUI}}
	})
	It("lists the targeted space and displays identity, enabled state and provisioning status", func() {
		Expect(cmd.Execute(nil)).To(Succeed())
		Expect(shared.CheckTargetCallCount()).To(Equal(1))
		org, space := shared.CheckTargetArgsForCall(0)
		Expect(org).To(BeTrue())
		Expect(space).To(BeTrue())
		Expect(actor.GetServiceAccountsInSpaceCallCount()).To(Equal(1))
		Expect(actor.GetServiceAccountsInSpaceArgsForCall(0)).To(Equal("space-guid"))
		Expect(testUI.Out).To(Say("Getting service accounts in org poc / space workers as operator"))
		Expect(testUI.Out).To(Say("name\\s+enabled\\s+status\\s+client id\\s+certificate DNS SAN"))
		Expect(testUI.Out).To(Say("first-account\\s+true\\s+ready\\s+cf:service-account:first-account\\s+first-account.svc.identity"))
		Expect(testUI.Out).To(Say("second-account\\s+false\\s+unprovisioned"))
		Expect(testUI.Err).To(Say("list warning"))
	})
	It("reports an empty space", func() {
		actor.GetServiceAccountsInSpaceReturns(nil, nil, nil)
		Expect(cmd.Execute(nil)).To(Succeed())
		Expect(testUI.Out).To(Say("No service accounts found"))
	})
	It("does not query when the target is missing", func() {
		missing := errors.New("no space targeted")
		shared.CheckTargetReturns(missing)
		Expect(cmd.Execute(nil)).To(MatchError(missing))
		Expect(actor.GetCurrentUserCallCount()).To(BeZero())
		Expect(actor.GetServiceAccountsInSpaceCallCount()).To(BeZero())
	})
	It("does not query when user resolution fails", func() {
		failed := errors.New("user unavailable")
		actor.GetCurrentUserReturns(configv3.User{}, failed)
		Expect(cmd.Execute(nil)).To(MatchError(failed))
		Expect(actor.GetServiceAccountsInSpaceCallCount()).To(BeZero())
	})
	It("returns listing errors with warnings rather than reporting an empty space", func() {
		denied := errors.New("denied")
		actor.GetServiceAccountsInSpaceReturns(nil, v7action.Warnings{"denial warning"}, denied)
		Expect(cmd.Execute(nil)).To(MatchError(denied))
		Expect(testUI.Err).To(Say("denial warning"))
		Expect(testUI.Out).NotTo(Say("No service accounts found"))
	})
})
