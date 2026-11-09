package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/v9/actor/v7action"
	"code.cloudfoundry.org/cli/v9/command/commandfakes"
	. "code.cloudfoundry.org/cli/v9/command/v7"
	"code.cloudfoundry.org/cli/v9/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/v9/resources"
	"code.cloudfoundry.org/cli/v9/util/configv3"
	"code.cloudfoundry.org/cli/v9/util/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("create-service-account Command", func() {
	var cmd CreateServiceAccountCommand
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
		actor.CreateServiceAccountInSpaceReturns(resources.ServiceAccount{
			Name: "shared-worker", ClientID: "cf:service-account:shared-worker", CertificateDNSSAN: "shared-worker.svc.identity", Status: "unprovisioned",
		}, v7action.Warnings{"account warning"}, nil)
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		cmd = CreateServiceAccountCommand{BaseCommand: BaseCommand{Actor: actor, SharedActor: shared, Config: config, UI: testUI}, Description: "Workers"}
		cmd.RequiredArgs.Name = "shared-worker"
	})
	It("creates in the targeted space, prints identity and accurately describes first-bind provisioning", func() {
		Expect(cmd.Execute(nil)).To(Succeed())
		Expect(shared.CheckTargetCallCount()).To(Equal(1))
		org, space := shared.CheckTargetArgsForCall(0)
		Expect(org).To(BeTrue())
		Expect(space).To(BeTrue())
		Expect(actor.CreateServiceAccountInSpaceCallCount()).To(Equal(1))
		name, description, guid := actor.CreateServiceAccountInSpaceArgsForCall(0)
		Expect(name).To(Equal("shared-worker"))
		Expect(description).To(Equal("Workers"))
		Expect(guid).To(Equal("space-guid"))
		Expect(testUI.Out).To(Say("Creating service account shared-worker in org poc / space workers as operator"))
		Expect(testUI.Out).To(Say("OK"))
		Expect(testUI.Out).To(Say("cf:service-account:shared-worker"))
		Expect(testUI.Out).To(Say("shared-worker.svc.identity"))
		Expect(testUI.Out).To(Say("unprovisioned"))
		Expect(testUI.Out).To(Say("Provisioning occurs on first bind"))
		Expect(testUI.Err).To(Say("account warning"))
	})
	It("fails before creation when login or the space target is missing", func() {
		missing := errors.New("no space targeted")
		shared.CheckTargetReturns(missing)
		Expect(cmd.Execute(nil)).To(MatchError(missing))
		Expect(actor.GetCurrentUserCallCount()).To(BeZero())
		Expect(actor.CreateServiceAccountInSpaceCallCount()).To(BeZero())
	})
	It("fails before creation when the current user cannot be resolved", func() {
		failed := errors.New("user unavailable")
		actor.GetCurrentUserReturns(configv3.User{}, failed)
		Expect(cmd.Execute(nil)).To(MatchError(failed))
		Expect(actor.CreateServiceAccountInSpaceCallCount()).To(BeZero())
	})
	It("displays warnings and returns a creation error without printing OK or a fabricated identity", func() {
		denied := errors.New("not authorized")
		actor.CreateServiceAccountInSpaceReturns(resources.ServiceAccount{}, v7action.Warnings{"denial warning"}, denied)
		Expect(cmd.Execute(nil)).To(MatchError(denied))
		Expect(testUI.Err).To(Say("denial warning"))
		Expect(testUI.Out).NotTo(Say("OK"))
		Expect(testUI.Out).NotTo(Say("cf:service-account:"))
	})
})
