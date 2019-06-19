package v7_test

import (
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/ui"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("unmap-route Command", func() {
	var (
		cmd             UnmapRouteCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeMapRouteActor
		input           *Buffer
		binaryName      string
		executeErr      error
		domain          string
		appName         string
		hostname        string
		path            string
	)

	BeforeEach(func() {
		input = NewBuffer()
		testUI = ui.NewTestUI(input, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeMapRouteActor)

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		appName = "my-app"
		domain = "some-domain.com"

		cmd = UnmapRouteCommand{
			RequiredArgs: flag.AppDomain{App: appName, Domain: domain},
			Hostname:     hostname,
			Path:         path,
			UI:           testUI,
			Config:       fakeConfig,
			SharedActor:  fakeSharedActor,
			Actor:        fakeActor,
		}
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	It("does not return an error", func() {
		Expect(executeErr).NotTo(HaveOccurred())
	})
})
