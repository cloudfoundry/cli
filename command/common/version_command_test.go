package common_test

import (
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/common"
	"code.cloudfoundry.org/cli/util/ui"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("Version Command", func() {
	var (
		cmd        VersionCommand
		fakeUI     *ui.UI
		fakeConfig *commandfakes.FakeConfig
		err        error
	)

	BeforeEach(func() {
		fakeUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeConfig.BinaryNameReturns("faceman")
		fakeConfig.BinaryVersionReturns("face2.0")
		fakeConfig.BinaryBuildDateReturns("yesterday")

		cmd = VersionCommand{
			UI:     fakeUI,
			Config: fakeConfig,
		}
	})

	It("displays correct version", func() {
		err = cmd.Execute(nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(fakeUI.Out).To(Say("faceman version face2.0-yesterday"))
	})
})
