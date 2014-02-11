package commands_test

import (
	"bytes"
	. "cf/commands"
	"cf/configuration"
	"cf/net"
	"cf/trace"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

func init() {
	Describe("Testing with ginkgo", func() {
		var deps curlDependencies

		BeforeEach(func() {
			deps = newCurlDependencies()
		})

		It("does not pass requirements when not logged in", func() {
			runCurlWithInputs(deps, []string{"/foo"})
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})

		Context("when logged in", func() {
			BeforeEach(func() {
				deps.reqFactory.LoginSuccess = true
			})

			It("fails with usage when not given enough input", func() {
				runCurlWithInputs(deps, []string{})
				Expect(deps.ui.FailedWithUsage).To(BeTrue())
			})

			It("passes requirements", func() {
				runCurlWithInputs(deps, []string{"/foo"})
				Expect(testcmd.CommandDidPassRequirements).To(BeTrue())
			})

			It("makes a get request given an endpoint", func() {
				deps.curlRepo.ResponseHeader = "Content-Size:1024"
				deps.curlRepo.ResponseBody = "response for get"
				runCurlWithInputs(deps, []string{"/foo"})

				Expect(deps.curlRepo.Method).To(Equal("GET"))
				Expect(deps.curlRepo.Path).To(Equal("/foo"))
				testassert.SliceContains(GinkgoT(), deps.ui.Outputs, testassert.Lines{
					{"response for get"},
				})
				testassert.SliceDoesNotContain(GinkgoT(), deps.ui.Outputs, testassert.Lines{
					{"FAILED"},
					{"Content-Size:1024"},
				})
			})

			It("makes a post request given -X", func() {
				runCurlWithInputs(deps, []string{"-X", "post", "/foo"})

				Expect(deps.curlRepo.Method).To(Equal("post"))
				testassert.SliceDoesNotContain(GinkgoT(), deps.ui.Outputs, testassert.Lines{
					{"FAILED"},
				})
			})

			It("sends headers given -H", func() {
				runCurlWithInputs(deps, []string{"-H", "Content-Type:cat", "/foo"})

				Expect(deps.curlRepo.Header).To(Equal("Content-Type:cat"))
				testassert.SliceDoesNotContain(GinkgoT(), deps.ui.Outputs, testassert.Lines{
					{"FAILED"},
				})
			})

			It("sends multiple headers given multiple -H flags", func() {
				runCurlWithInputs(deps, []string{"-H", "Content-Type:cat", "-H", "Content-Length:12", "/foo"})

				Expect(deps.curlRepo.Header).To(Equal("Content-Type:cat\nContent-Length:12"))
				testassert.SliceDoesNotContain(GinkgoT(), deps.ui.Outputs, testassert.Lines{
					{"FAILED"},
				})
			})

			It("prints out the response headers given -i", func() {
				deps.curlRepo.ResponseHeader = "Content-Size:1024"
				deps.curlRepo.ResponseBody = "response for get"
				runCurlWithInputs(deps, []string{"-i", "/foo"})

				testassert.SliceContains(GinkgoT(), deps.ui.Outputs, testassert.Lines{
					{"Content-Size:1024"},
					{"response for get"},
				})
				testassert.SliceDoesNotContain(GinkgoT(), deps.ui.Outputs, testassert.Lines{
					{"FAILED"},
				})
			})

			It("sets the request body given -d", func() {
				runCurlWithInputs(deps, []string{"-d", "body content to upload", "/foo"})

				Expect(deps.curlRepo.Body).To(Equal("body content to upload"))
				testassert.SliceDoesNotContain(GinkgoT(), deps.ui.Outputs, testassert.Lines{
					{"FAILED"},
				})
			})

			It("prints verbose output given the -v flag", func() {
				output := bytes.NewBuffer(make([]byte, 1024))
				trace.SetStdout(output)

				runCurlWithInputs(deps, []string{"-v", "/foo"})
				trace.Logger.Print("logging enabled")

				testassert.SliceContains(GinkgoT(), []string{output.String()}, testassert.Lines{
					{"logging enabled"},
				})
			})

			It("prints a failure message when the response is not success", func() {
				deps.curlRepo.ApiResponse = net.NewApiResponseWithMessage("ooops")
				runCurlWithInputs(deps, []string{"/foo"})

				testassert.SliceContains(GinkgoT(), deps.ui.Outputs, testassert.Lines{
					{"FAILED"},
					{"ooops"},
				})
			})
		})
	})
}

type curlDependencies struct {
	ui         *testterm.FakeUI
	config     configuration.Reader
	reqFactory *testreq.FakeReqFactory
	curlRepo   *testapi.FakeCurlRepository
}

func newCurlDependencies() (deps curlDependencies) {
	deps.ui = &testterm.FakeUI{}
	deps.config = testconfig.NewRepository()
	deps.reqFactory = &testreq.FakeReqFactory{}
	deps.curlRepo = &testapi.FakeCurlRepository{}
	return
}

func runCurlWithInputs(deps curlDependencies, inputs []string) {
	ctxt := testcmd.NewContext("curl", inputs)
	cmd := NewCurl(deps.ui, deps.config, deps.curlRepo)
	testcmd.RunCommand(cmd, ctxt, deps.reqFactory)
}
