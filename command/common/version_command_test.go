package common_test

import (
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/common"
	"code.cloudfoundry.org/cli/util/ui"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("Version Command", func() {
	var (
		cmd        VersionCommand
		testUI     *ui.UI
		fakeConfig *commandfakes.FakeConfig
		err        error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeConfig.BinaryNameReturns("faceman")
		fakeConfig.BinaryVersionReturns("0.0.0-invalid-version")

		cmd = VersionCommand{
			UI:     testUI,
			Config: fakeConfig,
		}
	})

	It("displays correct version", func() {
		err = cmd.Execute(nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(testUI.Out).To(Say("faceman version 0.0.0-invalid-version"))
	})
})
