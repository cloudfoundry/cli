package common_test

import (
	"code.cloudfoundry.org/cli/v9/actor/sharedaction"
	"code.cloudfoundry.org/cli/v9/command/commandfakes"
	"code.cloudfoundry.org/cli/v9/command/common"
	"code.cloudfoundry.org/cli/v9/util/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("service account help discovery", func() {
	DescribeTable("shows both service-account commands in a dedicated category", func(all bool, heading string) {
		config := new(commandfakes.FakeConfig)
		config.BinaryNameReturns("cf")
		testUI := ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		cmd := common.HelpCommand{UI: testUI, Config: config, Actor: sharedaction.NewActor(config), AllCommands: all}
		Expect(cmd.Execute(nil)).To(Succeed())
		Expect(testUI.Out).To(Say(heading))
		Expect(testUI.Out).To(Say("service-accounts"))
		Expect(testUI.Out).To(Say("create-service-account"))
		for _, name := range []string{"service-account", "delete-service-account", "enable-service-account", "disable-service-account", "bind-service-account", "unbind-service-account"} {
			Expect(common.Commands.HasCommand(name)).To(BeTrue(), name)
			Expect(testUI.Out).To(Say(name))
		}
	},
		Entry("default help", false, "Service accounts:"),
		Entry("full help", true, "SERVICE ACCOUNTS:"),
	)
})
