package commands_test

import (
	"bytes"
	. "cf/commands"
	"cf/configuration"
	"cf/net"
	"cf/trace"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestCurlFailsWithUsage(t *testing.T) {
	deps := newCurlDependencies()
	runCurlWithInputs(deps, []string{})
	assert.True(t, deps.ui.FailedWithUsage)

	deps = newCurlDependencies()
	runCurlWithInputs(deps, []string{"/foo"})
	assert.False(t, deps.ui.FailedWithUsage)
}

func TestCurlRequiresLogin(t *testing.T) {
	deps := newCurlDependencies()
	deps.reqFactory.LoginSuccess = false
	runCurlWithInputs(deps, []string{"/foo"})
	assert.False(t, testcmd.CommandDidPassRequirements)

	deps = newCurlDependencies()
	runCurlWithInputs(deps, []string{"/foo"})
	assert.True(t, testcmd.CommandDidPassRequirements)
}

func TestRequestWithNoFlags(t *testing.T) {
	deps := newCurlDependencies()

	deps.curlRepo.ResponseHeader = "Content-Size:1024"
	deps.curlRepo.ResponseBody = "response for get"
	runCurlWithInputs(deps, []string{"/foo"})

	assert.Equal(t, deps.curlRepo.Method, "GET")
	assert.Equal(t, deps.curlRepo.Path, "/foo")
	testassert.SliceContains(t, deps.ui.Outputs, testassert.Lines{
		{"response for get"},
	})
	testassert.SliceDoesNotContain(t, deps.ui.Outputs, testassert.Lines{
		{"FAILED"},
		{"Content-Size:1024"},
	})
}

func TestRequestWithMethodFlag(t *testing.T) {
	deps := newCurlDependencies()

	runCurlWithInputs(deps, []string{"-X", "post", "/foo"})

	assert.Equal(t, deps.curlRepo.Method, "post")
	testassert.SliceDoesNotContain(t, deps.ui.Outputs, testassert.Lines{
		{"FAILED"},
	})
}

func TestGetRequestWithHeaderFlag(t *testing.T) {
	deps := newCurlDependencies()

	runCurlWithInputs(deps, []string{"-H", "Content-Type:cat", "/foo"})

	assert.Equal(t, deps.curlRepo.Header, "Content-Type:cat")
	testassert.SliceDoesNotContain(t, deps.ui.Outputs, testassert.Lines{
		{"FAILED"},
	})
}

func TestGetRequestWithMultipleHeaderFlags(t *testing.T) {
	deps := newCurlDependencies()

	runCurlWithInputs(deps, []string{"-H", "Content-Type:cat", "-H", "Content-Length:12", "/foo"})

	assert.Equal(t, deps.curlRepo.Header, "Content-Type:cat\nContent-Length:12")
	testassert.SliceDoesNotContain(t, deps.ui.Outputs, testassert.Lines{
		{"FAILED"},
	})
}

func TestGetRequestWithIncludeFlag(t *testing.T) {
	deps := newCurlDependencies()

	deps.curlRepo.ResponseHeader = "Content-Size:1024"
	deps.curlRepo.ResponseBody = "response for get"
	runCurlWithInputs(deps, []string{"-i", "/foo"})

	testassert.SliceContains(t, deps.ui.Outputs, testassert.Lines{
		{"Content-Size:1024"},
		{"response for get"},
	})
	testassert.SliceDoesNotContain(t, deps.ui.Outputs, testassert.Lines{
		{"FAILED"},
	})
}

func TestGetRequestWithDataFlag(t *testing.T) {
	deps := newCurlDependencies()

	runCurlWithInputs(deps, []string{"-d", "body content to upload", "/foo"})

	assert.Equal(t, deps.curlRepo.Body, "body content to upload")
	testassert.SliceDoesNotContain(t, deps.ui.Outputs, testassert.Lines{
		{"FAILED"},
	})
}

func TestGetRequestWithVerboseFlagEnablesTrace(t *testing.T) {
	deps := newCurlDependencies()
	output := bytes.NewBuffer(make([]byte, 1024))
	trace.SetStdout(output)

	runCurlWithInputs(deps, []string{"-v", "/foo"})
	trace.Logger.Print("logging enabled")

	assert.Contains(t, output.String(), "logging enabled")
}

func TestGetRequestFailsWithError(t *testing.T) {
	deps := newCurlDependencies()

	deps.curlRepo.ApiResponse = net.NewApiResponseWithMessage("ooops")
	runCurlWithInputs(deps, []string{"/foo"})

	testassert.SliceContains(t, deps.ui.Outputs, testassert.Lines{
		{"FAILED"},
		{"ooops"},
	})
}

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
