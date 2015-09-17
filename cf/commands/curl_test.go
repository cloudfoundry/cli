package commands_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"

	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/trace"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	"github.com/cloudfoundry/gofileutils/fileutils"

	"github.com/cloudfoundry/cli/cf/command_registry"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("curl command", func() {
	var (
		ui                  *testterm.FakeUI
		config              core_config.Repository
		requirementsFactory *testreq.FakeReqFactory
		curlRepo            *testapi.FakeCurlRepository
		deps                command_registry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.RepoLocator = deps.RepoLocator.SetCurlRepository(curlRepo)
		deps.Config = config
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("curl").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		config = testconfig.NewRepository()
		requirementsFactory = &testreq.FakeReqFactory{}
		curlRepo = &testapi.FakeCurlRepository{}
	})

	runCurlWithInputs := func(args []string) bool {
		return testcmd.RunCliCommand("curl", args, requirementsFactory, updateCommandDependency, false)
	}

	It("does not pass requirements when not logged in", func() {
		Expect(runCurlWithInputs([]string{"/foo"})).To(BeFalse())
	})

	Context("when logged in", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
		})

		It("fails with usage when not given enough input", func() {
			runCurlWithInputs([]string{})
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Incorrect Usage", "An argument is missing or not correctly enclosed"},
			))
		})

		It("passes requirements", func() {
			Expect(runCurlWithInputs([]string{"/foo"})).To(BeTrue())
		})

		It("makes a get request given an endpoint", func() {
			curlRepo.ResponseHeader = "Content-Size:1024"
			curlRepo.ResponseBody = "response for get"
			runCurlWithInputs([]string{"/foo"})

			Expect(curlRepo.Method).To(Equal("GET"))
			Expect(curlRepo.Path).To(Equal("/foo"))
			Expect(ui.Outputs).To(ContainSubstrings([]string{"response for get"}))
			Expect(ui.Outputs).ToNot(ContainSubstrings(
				[]string{"FAILED"},
				[]string{"Content-Size:1024"},
			))
		})

		Context("when the --output flag is provided", func() {
			It("saves the body of the response to the given filepath if it exists", func() {
				fileutils.TempFile("poor-mans-pipe", func(tempFile *os.File, err error) {
					Expect(err).ToNot(HaveOccurred())
					curlRepo.ResponseBody = "hai"

					runCurlWithInputs([]string{"--output", tempFile.Name(), "/foo"})
					contents, err := ioutil.ReadAll(tempFile)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(contents)).To(Equal("hai"))
				})
			})

			It("saves the body of the response to the given filepath if it doesn't exists", func() {
				fileutils.TempDir("poor-mans-dir", func(tmpDir string, err error) {
					Expect(err).ToNot(HaveOccurred())
					curlRepo.ResponseBody = "hai"

					filePath := filepath.Join(tmpDir, "subdir1", "banana.txt")
					runCurlWithInputs([]string{"--output", filePath, "/foo"})

					file, err := os.Open(filePath)
					Expect(err).ToNot(HaveOccurred())

					contents, err := ioutil.ReadAll(file)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(contents)).To(Equal("hai"))
				})
			})
		})

		It("makes a post request given -X", func() {
			runCurlWithInputs([]string{"-X", "post", "/foo"})

			Expect(curlRepo.Method).To(Equal("post"))
			Expect(ui.Outputs).ToNot(ContainSubstrings([]string{"FAILED"}))
		})

		It("sends headers given -H", func() {
			runCurlWithInputs([]string{"-H", "Content-Type:cat", "/foo"})

			Expect(curlRepo.Header).To(Equal("Content-Type:cat"))
			Expect(ui.Outputs).ToNot(ContainSubstrings([]string{"FAILED"}))
		})

		It("sends multiple headers given multiple -H flags", func() {
			runCurlWithInputs([]string{"-H", "Content-Type:cat", "-H", "Content-Length:12", "/foo"})

			Expect(curlRepo.Header).To(Equal("Content-Type:cat\nContent-Length:12"))
			Expect(ui.Outputs).ToNot(ContainSubstrings([]string{"FAILED"}))
		})

		It("prints out the response headers given -i", func() {
			curlRepo.ResponseHeader = "Content-Size:1024"
			curlRepo.ResponseBody = "response for get"
			runCurlWithInputs([]string{"-i", "/foo"})

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Content-Size:1024"},
				[]string{"response for get"},
			))
			Expect(ui.Outputs).ToNot(ContainSubstrings([]string{"FAILED"}))
		})

		It("sets the request body given -d", func() {
			runCurlWithInputs([]string{"-d", "body content to upload", "/foo"})

			Expect(curlRepo.Body).To(Equal("body content to upload"))
			Expect(ui.Outputs).ToNot(ContainSubstrings([]string{"FAILED"}))
		})

		It("prints verbose output given the -v flag", func() {
			output := bytes.NewBuffer(make([]byte, 1024))
			trace.SetStdout(output)

			runCurlWithInputs([]string{"-v", "/foo"})
			trace.Logger.Print("logging enabled")

			Expect([]string{output.String()}).To(ContainSubstrings([]string{"logging enabled"}))
		})

		It("prints a failure message when the response is not success", func() {
			curlRepo.Error = errors.New("ooops")
			runCurlWithInputs([]string{"/foo"})

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"FAILED"},
				[]string{"ooops"},
			))
		})

		Context("Whent the content type is JSON", func() {
			BeforeEach(func() {
				curlRepo.ResponseHeader = "Content-Type: application/json;charset=utf-8"
				curlRepo.ResponseBody = `{"total_results":0,"total_pages":1,"prev_url":null,"next_url":null,"resources":[]}`
			})

			It("pretty-prints the response body", func() {
				runCurlWithInputs([]string{"/ugly-printed-json-endpoint"})

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"{"},
					[]string{"  \"total_results", "0"},
					[]string{"  \"total_pages", "1"},
					[]string{"  \"prev_url", "null"},
					[]string{"  \"next_url", "null"},
					[]string{"  \"resources", "[]"},
					[]string{"}"},
				))
			})

			Context("But the body is not JSON", func() {
				BeforeEach(func() {
					curlRepo.ResponseBody = "FAIL: crumpets need MOAR butterz"
				})

				It("regular-prints the response body", func() {
					runCurlWithInputs([]string{"/whateverz"})

					Expect(ui.Outputs).To(Equal([]string{"FAIL: crumpets need MOAR butterz"}))
				})
			})
		})
	})
})
