package v7_test

import (
	"code.cloudfoundry.org/cli/command/flag"
	v7 "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"

	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("target Command", func() {
	var (
		cmd             v7.CurlCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		binaryName      string

		CustomHeaders         []string
		HTTPMethod            string
		HTTPData              flag.PathWithAt
		FailOnHTTPError       bool
		IncludeReponseHeaders bool
		OutputFile            flag.Path
		executeErr            error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		CustomHeaders = []string{"/v3/apps"}
		cmd = v7.CurlCommand{
			BaseCommand: v7.BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
			RequiredArgs:          flag.APIPath{Path: "/"},
			CustomHeaders:         CustomHeaders,
			HTTPMethod:            HTTPMethod,
			HTTPData:              HTTPData,
			FailOnHTTPError:       FailOnHTTPError,
			IncludeReponseHeaders: IncludeReponseHeaders,
			OutputFile:            OutputFile,
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	It("formats the URL", func() {
		//	expect makeRequestArgsForCall.URL = full URL with formatting

	})

	When("a user provides custom headers", func() {
		// one before each and it for conidtinoal path in mergeHeaders

	})

	When("a body is provided and no method is provided", func() {
		BeforeEach(func() {

		})

		It("defaults to post", func() {
			//	expect makeRequestArgsForCall.URL = full URL with formatting
		})
	})

	When("every flag is provided", func() {
		BeforeEach(func() {

		})

		It("formats the the input into a http request struct", func() {
			//	white box testing
		})
	})

	When("the request is not successful", func() {
		BeforeEach(func() {

		})

		It("returns the API error", func() {

		})
	})

	When("the request is successful", func() {
		BeforeEach(func() {

		})

		It("returns the API error", func() {
			// lack box testing
		})
	})
})
