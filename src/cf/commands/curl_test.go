package commands_test

import (
	"bytes"
	. "cf/commands"
	"cf/configuration"
	"cf/net"
	"cf/trace"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

type curlDependencies struct {
	ui         *testterm.FakeUI
	config     *configuration.Configuration
	reqFactory *testreq.FakeReqFactory
	curlRepo   *testapi.FakeCurlRepository
}

func newCurlDependencies() (deps curlDependencies) {
	deps.ui = &testterm.FakeUI{}
	deps.config = &configuration.Configuration{}
	deps.reqFactory = &testreq.FakeReqFactory{
		LoginSuccess: true,
	}
	deps.curlRepo = &testapi.FakeCurlRepository{}
	return
}

func runCurlWithInputs(deps curlDependencies, inputs []string) {
	ctxt := testcmd.NewContext("curl", inputs)
	cmd := NewCurl(deps.ui, deps.config, deps.curlRepo)
	testcmd.RunCommand(cmd, ctxt, deps.reqFactory)
}
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestCurlFailsWithUsage", func() {
			deps := newCurlDependencies()
			runCurlWithInputs(deps, []string{})
			assert.True(mr.T(), deps.ui.FailedWithUsage)

			deps = newCurlDependencies()
			runCurlWithInputs(deps, []string{"/foo"})
			assert.False(mr.T(), deps.ui.FailedWithUsage)
		})
		It("TestCurlRequiresLogin", func() {

			deps := newCurlDependencies()
			deps.reqFactory.LoginSuccess = false
			runCurlWithInputs(deps, []string{"/foo"})
			assert.False(mr.T(), testcmd.CommandDidPassRequirements)

			deps = newCurlDependencies()
			runCurlWithInputs(deps, []string{"/foo"})
			assert.True(mr.T(), testcmd.CommandDidPassRequirements)
		})
		It("TestRequestWithNoFlags", func() {

			deps := newCurlDependencies()

			deps.curlRepo.ResponseHeader = "Content-Size:1024"
			deps.curlRepo.ResponseBody = "response for get"
			runCurlWithInputs(deps, []string{"/foo"})

			assert.Equal(mr.T(), deps.curlRepo.Method, "GET")
			assert.Equal(mr.T(), deps.curlRepo.Path, "/foo")
			testassert.SliceContains(mr.T(), deps.ui.Outputs, testassert.Lines{
				{"response for get"},
			})
			testassert.SliceDoesNotContain(mr.T(), deps.ui.Outputs, testassert.Lines{
				{"FAILED"},
				{"Content-Size:1024"},
			})
		})
		It("TestRequestWithMethodFlag", func() {

			deps := newCurlDependencies()

			runCurlWithInputs(deps, []string{"-X", "post", "/foo"})

			assert.Equal(mr.T(), deps.curlRepo.Method, "post")
			testassert.SliceDoesNotContain(mr.T(), deps.ui.Outputs, testassert.Lines{
				{"FAILED"},
			})
		})
		It("TestGetRequestWithHeaderFlag", func() {

			deps := newCurlDependencies()

			runCurlWithInputs(deps, []string{"-H", "Content-Type:cat", "/foo"})

			assert.Equal(mr.T(), deps.curlRepo.Header, "Content-Type:cat")
			testassert.SliceDoesNotContain(mr.T(), deps.ui.Outputs, testassert.Lines{
				{"FAILED"},
			})
		})
		It("TestGetRequestWithMultipleHeaderFlags", func() {

			deps := newCurlDependencies()

			runCurlWithInputs(deps, []string{"-H", "Content-Type:cat", "-H", "Content-Length:12", "/foo"})

			assert.Equal(mr.T(), deps.curlRepo.Header, "Content-Type:cat\nContent-Length:12")
			testassert.SliceDoesNotContain(mr.T(), deps.ui.Outputs, testassert.Lines{
				{"FAILED"},
			})
		})
		It("TestGetRequestWithIncludeFlag", func() {

			deps := newCurlDependencies()

			deps.curlRepo.ResponseHeader = "Content-Size:1024"
			deps.curlRepo.ResponseBody = "response for get"
			runCurlWithInputs(deps, []string{"-i", "/foo"})

			testassert.SliceContains(mr.T(), deps.ui.Outputs, testassert.Lines{
				{"Content-Size:1024"},
				{"response for get"},
			})
			testassert.SliceDoesNotContain(mr.T(), deps.ui.Outputs, testassert.Lines{
				{"FAILED"},
			})
		})
		It("TestGetRequestWithDataFlag", func() {

			deps := newCurlDependencies()

			runCurlWithInputs(deps, []string{"-d", "body content to upload", "/foo"})

			assert.Equal(mr.T(), deps.curlRepo.Body, "body content to upload")
			testassert.SliceDoesNotContain(mr.T(), deps.ui.Outputs, testassert.Lines{
				{"FAILED"},
			})
		})
		It("TestGetRequestWithVerboseFlagEnablesTrace", func() {

			deps := newCurlDependencies()
			output := bytes.NewBuffer(make([]byte, 1024))
			trace.SetStdout(output)

			runCurlWithInputs(deps, []string{"-v", "/foo"})
			trace.Logger.Print("logging enabled")

			assert.Contains(mr.T(), output.String(), "logging enabled")
		})
		It("TestGetRequestFailsWithError", func() {

			deps := newCurlDependencies()

			deps.curlRepo.ApiResponse = net.NewApiResponseWithMessage("ooops")
			runCurlWithInputs(deps, []string{"/foo"})

			testassert.SliceContains(mr.T(), deps.ui.Outputs, testassert.Lines{
				{"FAILED"},
				{"ooops"},
			})
		})
	})
}
