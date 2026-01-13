package v7_test

import (
	"errors"
	"net/http"
	"os"

	"code.cloudfoundry.org/cli/v9/command/flag"
	v7 "code.cloudfoundry.org/cli/v9/command/v7"
	"code.cloudfoundry.org/cli/v9/command/v7/v7fakes"

	"code.cloudfoundry.org/cli/v9/command/commandfakes"
	"code.cloudfoundry.org/cli/v9/util/ui"
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

	When("the request contains double curly braces", func() {
		BeforeEach(func() {
			fakeActor.MakeCurlRequestReturns([]byte("{{ }}"), &http.Response{}, nil)
		})

		It("returns the literal text", func() {
			Expect(executeErr).NotTo(HaveOccurred())
			Expect(testUI.Out).To(Say("{{ }}"))
		})
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
			outputFile, err := os.CreateTemp("", "")
			Expect(err).NotTo(HaveOccurred())
			outputFileName := outputFile.Name()
			cmd.OutputFile = flag.Path(outputFileName)
		})

		AfterEach(func() {
			os.RemoveAll(string(cmd.OutputFile))
		})

		It("writes the output to the file", func() {
			fileContents, err := os.ReadFile(string(cmd.OutputFile))
			Expect(string(fileContents)).To(Equal("sarah, teal, and reid were here"))
			Expect(err).ToNot(HaveOccurred())
		})

		When("the include-response-headers flag is set", func() {
			BeforeEach(func() {
				cmd.IncludeResponseHeaders = true
			})

			It("includes the headers in the output", func() {
				fileContents, err := os.ReadFile(string(cmd.OutputFile))
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

	When("the response contains binary data", func() {
		var binaryData []byte

		BeforeEach(func() {
			// Create binary data with null bytes (like a droplet file)
			binaryData = []byte{0x50, 0x4B, 0x03, 0x04, 0x00, 0x00, 0x08, 0x00, 0x00, 0x00}
			fakeActor.MakeCurlRequestReturns(binaryData, &http.Response{
				Header: http.Header{
					"Content-Type": []string{"application/octet-stream"},
				},
			}, nil)
		})

		It("writes binary data directly to stdout without string conversion", func() {
			Expect(executeErr).NotTo(HaveOccurred())
			Expect(testUI.Out).To(Say(string(binaryData)))
		})

		When("Content-Type is not a known binary MIME type", func() {
			BeforeEach(func() {
				// Create binary data with null bytes (like a droplet file)
				binaryData = []byte{0x50, 0x4B, 0x03, 0x04, 0x00, 0x00, 0x08, 0x00, 0x00, 0x00}
				fakeActor.MakeCurlRequestReturns(binaryData, &http.Response{
					Header: http.Header{
						"Content-Type": []string{"text/plain"},
					},
				}, nil)
			})
			It("inspects the response data and writes binary data directly to stdout without string conversion", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(testUI.Out).To(Say(string(binaryData)))
			})
		})

		When("include-response-headers flag is set", func() {
			BeforeEach(func() {
				cmd.IncludeResponseHeaders = true
			})

			It("writes headers as text and binary data separately", func() {
				Expect(executeErr).NotTo(HaveOccurred())

				// Check that headers are written as text (using DisplayTextLiteral)
				Expect(testUI.Out).To(Say("Content-Type: application/octet-stream"))

				// Check that binary data is preserved
				Expect(testUI.Out).To(Say(string(binaryData)))
			})
		})

		When("output file is specified", func() {
			BeforeEach(func() {
				outputFile, err := os.CreateTemp("", "binary-output")
				Expect(err).NotTo(HaveOccurred())
				cmd.OutputFile = flag.Path(outputFile.Name())
			})

			AfterEach(func() {
				os.RemoveAll(string(cmd.OutputFile))
			})

			It("writes binary data to the file correctly", func() {
				Expect(executeErr).NotTo(HaveOccurred())

				fileContents, err := os.ReadFile(string(cmd.OutputFile))
				Expect(err).ToNot(HaveOccurred())
				Expect(fileContents).To(Equal(binaryData))
			})
		})
	})

	When("the response is empty", func() {
		BeforeEach(func() {
			fakeActor.MakeCurlRequestReturns([]byte{}, &http.Response{
				Header: http.Header{},
			}, nil)
		})

		It("handles empty response correctly", func() {
			Expect(executeErr).NotTo(HaveOccurred())
			Expect(testUI.Out).To(Say(""))
		})
	})

})
