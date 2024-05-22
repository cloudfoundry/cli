package v7_test

import (
	"errors"
	"io/ioutil"
	"net/http"
	"os"

	"code.cloudfoundry.org/cli/command/flag"
	v7 "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"

	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("curl Command", func() {
	var (
		cmd             v7.CurlCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		binaryName      string

		CustomHeaders          []string
		HTTPMethod             string
		HTTPData               flag.PathWithAt
		FailOnHTTPError        bool
		IncludeResponseHeaders bool
		OutputFile             flag.Path
		executeErr             error
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
			RequiredArgs:           flag.APIPath{Path: "/"},
			CustomHeaders:          CustomHeaders,
			HTTPMethod:             HTTPMethod,
			HTTPData:               HTTPData,
			FailOnHTTPError:        FailOnHTTPError,
			IncludeResponseHeaders: IncludeResponseHeaders,
			OutputFile:             OutputFile,
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		fakeActor.MakeCurlRequestReturns([]byte("sarah, teal, and reid were here"), &http.Response{
			Header: http.Header{
				"X-Name": []string{"athleisure"},
			},
		}, nil)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	It("makes a request with the given flags", func() {
		Expect(fakeActor.MakeCurlRequestCallCount()).To(Equal(1))
		httpMethod, path, customHeaders, httpData, failOnHTTPError := fakeActor.MakeCurlRequestArgsForCall(0)
		Expect(httpMethod).To(Equal(HTTPMethod))
		Expect(path).To(Equal(cmd.RequiredArgs.Path))
		Expect(customHeaders).To(Equal(CustomHeaders))
		Expect(httpData).To(Equal(string(HTTPData)))
		Expect(failOnHTTPError).To(Equal(FailOnHTTPError))
		Expect(executeErr).ToNot(HaveOccurred())
	})

	When("the verbose flag is set", func() {
		BeforeEach(func() {
			fakeConfig.VerboseReturns(true, nil)
		})

		It("does not write any additional output", func() {
			Expect(executeErr).NotTo(HaveOccurred())
			Expect(testUI.Out).NotTo(Say(".+"))
		})
	})

	It("writes the response to stdout", func() {
		Expect(executeErr).NotTo(HaveOccurred())
		Expect(testUI.Out).To(Say("sarah, teal, and reid were here"))
	})

	When("the request errors", func() {
		BeforeEach(func() {
			fakeActor.MakeCurlRequestReturns([]byte{}, &http.Response{}, errors.New("this sucks"))
		})

		It("returns the error", func() {
			Expect(executeErr).To(MatchError("this sucks"))
		})
	})

	When("an output file is given", func() {
		BeforeEach(func() {
			outputFile, err := ioutil.TempFile("", "")
			Expect(err).NotTo(HaveOccurred())
			outputFileName := outputFile.Name()
			cmd.OutputFile = flag.Path(outputFileName)
		})

		AfterEach(func() {
			os.RemoveAll(string(cmd.OutputFile))
		})

		It("writes the output to the file", func() {
			fileContents, err := ioutil.ReadFile(string(cmd.OutputFile))
			Expect(string(fileContents)).To(Equal("sarah, teal, and reid were here"))
			Expect(err).ToNot(HaveOccurred())
		})

		When("the include-response-headers flag is set", func() {
			BeforeEach(func() {
				cmd.IncludeResponseHeaders = true
			})

			It("includes the headers in the output", func() {
				fileContents, err := ioutil.ReadFile(string(cmd.OutputFile))
				Expect(string(fileContents)).To(ContainSubstring("X-Name: athleisure"))
				Expect(string(fileContents)).To(ContainSubstring("sarah, teal, and reid were here"))
				Expect(err).ToNot(HaveOccurred())
			})
		})

		When("writing the file fails", func() {
			BeforeEach(func() {
				cmd.OutputFile = flag.Path("/🕵/sleuthin/for/a/file")
			})

			It("returns the error", func() {
				Expect(executeErr.Error()).To(ContainSubstring("Error creating file"))
			})
		})
	})

})
