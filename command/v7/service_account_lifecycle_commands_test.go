package v7_test

import (
	"errors"
	"strings"

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

var _ = Describe("Service Account Lifecycle Commands", func() {
	var base BaseCommand
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
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		base = BaseCommand{Actor: actor, SharedActor: shared, Config: config, UI: testUI}
	})
	run := func(operation string) error {
		switch operation {
		case "show":
			cmd := ServiceAccountCommand{BaseCommand: base}
			cmd.RequiredArgs.Name = "workers"
			return cmd.Execute(nil)
		case "delete":
			cmd := DeleteServiceAccountCommand{BaseCommand: base, Force: true}
			cmd.RequiredArgs.Name = "workers"
			return cmd.Execute(nil)
		case "enable":
			cmd := EnableServiceAccountCommand{BaseCommand: base}
			cmd.RequiredArgs.Name = "workers"
			return cmd.Execute(nil)
		case "disable":
			cmd := DisableServiceAccountCommand{BaseCommand: base}
			cmd.RequiredArgs.Name = "workers"
			return cmd.Execute(nil)
		case "bind":
			cmd := BindServiceAccountCommand{BaseCommand: base}
			cmd.RequiredArgs.AppName = "app"
			cmd.RequiredArgs.AccountName = "workers"
			return cmd.Execute(nil)
		default:
			cmd := UnbindServiceAccountCommand{BaseCommand: base}
			cmd.RequiredArgs.AppName = "app"
			return cmd.Execute(nil)
		}
	}
	DescribeTable("performs the operation in the targeted space and reports correct consequences", func(operation string) {
		actor.GetServiceAccountByNameAndSpaceReturns(resources.ServiceAccount{GUID: "account-guid", Name: "workers", Description: "Worker identity", ClientID: "cf:service-account:workers", CertificateDNSSAN: "workers.svc.identity", Enabled: true, Status: "ready"}, v7action.Warnings{"warning"}, nil)
		actor.DeleteServiceAccountByNameAndSpaceReturns(v7action.Warnings{"warning"}, nil)
		actor.SetServiceAccountEnabledByNameAndSpaceReturns(v7action.Warnings{"warning"}, nil)
		actor.BindServiceAccountByNameAndSpaceReturns(v7action.Warnings{"warning"}, nil)
		actor.UnbindServiceAccountByAppNameAndSpaceReturns(v7action.Warnings{"warning"}, nil)
		Expect(run(operation)).To(Succeed())
		org, space := shared.CheckTargetArgsForCall(0)
		Expect(org).To(BeTrue())
		Expect(space).To(BeTrue())
		Expect(testUI.Err).To(Say("warning"))
		switch operation {
		case "show":
			name, guid := actor.GetServiceAccountByNameAndSpaceArgsForCall(0)
			Expect(name).To(Equal("workers"))
			Expect(guid).To(Equal("space-guid"))
			Expect(string(testUI.Out.(*Buffer).Contents())).To(ContainSubstring("account-guid"))
			Expect(string(testUI.Out.(*Buffer).Contents())).To(ContainSubstring("Worker identity"))
			Expect(string(testUI.Out.(*Buffer).Contents())).To(ContainSubstring("cf:service-account:workers"))
			Expect(string(testUI.Out.(*Buffer).Contents())).To(ContainSubstring("workers.svc.identity"))
			Expect(string(testUI.Out.(*Buffer).Contents())).To(ContainSubstring("ready"))
		case "delete":
			name, guid := actor.DeleteServiceAccountByNameAndSpaceArgsForCall(0)
			Expect(name).To(Equal("workers"))
			Expect(guid).To(Equal("space-guid"))
			Expect(string(testUI.Out.(*Buffer).Contents())).To(ContainSubstring("permanently reserved"))
		case "enable", "disable":
			name, guid, enabled := actor.SetServiceAccountEnabledByNameAndSpaceArgsForCall(0)
			Expect(name).To(Equal("workers"))
			Expect(guid).To(Equal("space-guid"))
			Expect(enabled).To(Equal(operation == "enable"))
			if operation == "disable" {
				Expect(string(testUI.Out.(*Buffer).Contents())).To(ContainSubstring("Existing tokens and certificates are not revoked"))
			}
		case "bind":
			app, name, guid := actor.BindServiceAccountByNameAndSpaceArgsForCall(0)
			Expect(app).To(Equal("app"))
			Expect(name).To(Equal("workers"))
			Expect(guid).To(Equal("space-guid"))
			Expect(string(testUI.Out.(*Buffer).Contents())).To(ContainSubstring("Restart app"))
			Expect(string(testUI.Out.(*Buffer).Contents())).To(ContainSubstring("Roles are not granted by binding"))
		case "unbind":
			app, guid := actor.UnbindServiceAccountByAppNameAndSpaceArgsForCall(0)
			Expect(app).To(Equal("app"))
			Expect(guid).To(Equal("space-guid"))
			Expect(string(testUI.Out.(*Buffer).Contents())).To(ContainSubstring("Restart app"))
			Expect(string(testUI.Out.(*Buffer).Contents())).To(ContainSubstring("Existing tokens and certificates are not revoked"))
			Expect(actor.RestartApplicationCallCount()).To(BeZero())
		}
		if operation != "show" {
			Expect(testUI.Out).To(Say("OK"))
		}
	}, Entry("show", "show"), Entry("delete", "delete"), Entry("enable", "enable"), Entry("disable", "disable"), Entry("bind", "bind"), Entry("unbind", "unbind"))
	DescribeTable("returns operation failures and warnings without success guidance", func(operation string) {
		failed := errors.New("operation failed")
		actor.GetServiceAccountByNameAndSpaceReturns(resources.ServiceAccount{}, v7action.Warnings{"warning"}, failed)
		actor.DeleteServiceAccountByNameAndSpaceReturns(v7action.Warnings{"warning"}, failed)
		actor.SetServiceAccountEnabledByNameAndSpaceReturns(v7action.Warnings{"warning"}, failed)
		actor.BindServiceAccountByNameAndSpaceReturns(v7action.Warnings{"warning"}, failed)
		actor.UnbindServiceAccountByAppNameAndSpaceReturns(v7action.Warnings{"warning"}, failed)
		Expect(run(operation)).To(MatchError(failed))
		Expect(testUI.Err).To(Say("warning"))
		Expect(testUI.Out).NotTo(Say("OK"))
		Expect(testUI.Out).NotTo(Say("Restart app"))
	}, Entry("show", "show"), Entry("delete", "delete"), Entry("enable", "enable"), Entry("disable", "disable"), Entry("bind", "bind"), Entry("unbind", "unbind"))
	DescribeTable("stops before mutation on target failure", func(operation string) {
		failed := errors.New("no target")
		shared.CheckTargetReturns(failed)
		Expect(run(operation)).To(MatchError(failed))
		Expect(actor.GetCurrentUserCallCount()).To(BeZero())
		Expect(actor.GetServiceAccountByNameAndSpaceCallCount() + actor.DeleteServiceAccountByNameAndSpaceCallCount() + actor.SetServiceAccountEnabledByNameAndSpaceCallCount() + actor.BindServiceAccountByNameAndSpaceCallCount() + actor.UnbindServiceAccountByAppNameAndSpaceCallCount()).To(BeZero())
	}, Entry("show", "show"), Entry("delete", "delete"), Entry("enable", "enable"), Entry("disable", "disable"), Entry("bind", "bind"), Entry("unbind", "unbind"))
	It("cancels deletion by default without calling the actor", func() {
		base.UI = ui.NewTestUI(strings.NewReader("n\n"), NewBuffer(), NewBuffer())
		cmd := DeleteServiceAccountCommand{BaseCommand: base}
		cmd.RequiredArgs.Name = "workers"
		Expect(cmd.Execute(nil)).To(Succeed())
		Expect(actor.DeleteServiceAccountByNameAndSpaceCallCount()).To(BeZero())
		Expect(base.UI.(*ui.UI).Out).To(Say("Delete cancelled"))
	})
	It("confirms deletion interactively and explains retained name reservation", func() {
		base.UI = ui.NewTestUI(strings.NewReader("y\n"), NewBuffer(), NewBuffer())
		cmd := DeleteServiceAccountCommand{BaseCommand: base}
		cmd.RequiredArgs.Name = "workers"
		Expect(cmd.Execute(nil)).To(Succeed())
		Expect(actor.DeleteServiceAccountByNameAndSpaceCallCount()).To(Equal(1))
		Expect(base.UI.(*ui.UI).Out).To(Say("permanently reserved"))
	})
	It("returns current-user failure before unbinding", func() {
		actor.GetCurrentUserReturns(configv3.User{}, errors.New("user failed"))
		Expect(run("unbind")).To(MatchError("user failed"))
		Expect(actor.UnbindServiceAccountByAppNameAndSpaceCallCount()).To(BeZero())
	})
})
