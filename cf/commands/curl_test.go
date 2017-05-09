package commands_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/cli/cf/api/apifakes"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"
	testcmd "code.cloudfoundry.org/cli/util/testhelpers/commands"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"
	"code.cloudfoundry.org/gofileutils/fileutils"

	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/trace"
	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("curl command", func() {
	var (
		ui                  *testterm.FakeUI
		config              coreconfig.Repository
		requirementsFactory *requirementsfakes.FakeFactory
		curlRepo            *apifakes.OldFakeCurlRepository
		deps                commandregistry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.RepoLocator = deps.RepoLocator.SetCurlRepository(curlRepo)
		deps.Config = config
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("curl").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		config = testconfig.NewRepository()
		requirementsFactory = new(requirementsfakes.FakeFactory)
		requirementsFactory.NewAPIEndpointRequirementReturns(requirements.Passing{})
		curlRepo = new(apifakes.OldFakeCurlRepository)

		trace.LoggingToStdout = false
	})

	runCurlWithInputs := func(args []string) bool {
		return testcmd.RunCLICommand("curl", args, requirementsFactory, updateCommandDependency, false, ui)
	}

	runCurlAsPluginWithInputs := func(args []string) bool {
		return testcmd.RunCLICommand("curl", args, requirementsFactory, updateCommandDependency, true, ui)
	}

	It("fails with usage when not given enough input", func() {
		runCurlWithInputs([]string{})
		Expect(ui.Outputs()).To(ContainSubstrings(
			[]string{"Incorrect Usage", "An argument is missing or not correctly enclosed"},
		))
	})

	Context("requirements", func() {
		Context("when no api is set", func() {
			BeforeEach(func() {
				requirementsFactory.NewAPIEndpointRequirementReturns(requirements.Failing{Message: "no api set"})
			})

			It("fails", func() {
				Expect(runCurlWithInputs([]string{"/foo"})).To(BeFalse())
			})
		})

		Context("when api is set", func() {
			BeforeEach(func() {
				requirementsFactory.NewAPIEndpointRequirementReturns(requirements.Passing{})
			})

			It("passes", func() {
				Expect(runCurlWithInputs([]string{"/foo"})).To(BeTrue())
			})
		})
	})

	It("makes a get request given an endpoint", func() {
		curlRepo.ResponseHeader = "Content-Size:1024"
		curlRepo.ResponseBody = "response for get"
		runCurlWithInputs([]string{"/foo"})

		Expect(curlRepo.Method).To(Equal(""))
		Expect(curlRepo.Path).To(Equal("/foo"))
		Expect(ui.Outputs()).To(ContainSubstrings([]string{"response for get"}))
		Expect(ui.Outputs()).ToNot(ContainSubstrings(
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
		Expect(ui.Outputs()).ToNot(ContainSubstrings([]string{"FAILED"}))
	})

	It("sends headers given -H", func() {
		runCurlWithInputs([]string{"-H", "Content-Type:cat", "/foo"})

		Expect(curlRepo.Header).To(Equal("Content-Type:cat"))
		Expect(ui.Outputs()).ToNot(ContainSubstrings([]string{"FAILED"}))
	})

	It("sends multiple headers given multiple -H flags", func() {
		runCurlWithInputs([]string{"-H", "Content-Type:cat", "-H", "Content-Length:12", "/foo"})

		Expect(curlRepo.Header).To(Equal("Content-Type:cat\nContent-Length:12"))
		Expect(ui.Outputs()).ToNot(ContainSubstrings([]string{"FAILED"}))
	})

	It("prints out the response headers given -i", func() {
		curlRepo.ResponseHeader = "Content-Size:1024"
		curlRepo.ResponseBody = "response for get"
		runCurlWithInputs([]string{"-i", "/foo"})

		Expect(ui.Outputs()).To(ContainSubstrings(
			[]string{"Content-Size:1024"},
			[]string{"response for get"},
		))
		Expect(ui.Outputs()).ToNot(ContainSubstrings([]string{"FAILED"}))
	})

	Context("when -d is provided", func() {
		It("sets the request body", func() {
			runCurlWithInputs([]string{"-d", "body content to upload", "/foo"})

			Expect(curlRepo.Body).To(Equal("body content to upload"))
			Expect(ui.Outputs()).ToNot(ContainSubstrings([]string{"FAILED"}))
		})

		It("does not fail with empty string", func() {
			runCurlWithInputs([]string{"/foo", "-d", ""})

			Expect(curlRepo.Body).To(Equal(""))
			Expect(curlRepo.Method).To(Equal("POST"))
			Expect(ui.Outputs()).NotTo(ContainSubstrings([]string{"FAILED"}))
		})

		It("uses given http verb if -X is also provided", func() {
			runCurlWithInputs([]string{"/foo", "-d", "some body", "-X", "PUT"})

			Expect(curlRepo.Body).To(Equal("some body"))
			Expect(curlRepo.Method).To(Equal("PUT"))
			Expect(ui.Outputs()).NotTo(ContainSubstrings([]string{"FAILED"}))
		})

		It("sets the request body with an @-prefixed file", func() {
			tempfile, err := ioutil.TempFile("", "get-data-test")
			Expect(err).NotTo(HaveOccurred())
			defer os.RemoveAll(tempfile.Name())
			jsonData := `{"some":"json"}`
			ioutil.WriteFile(tempfile.Name(), []byte(jsonData), os.ModePerm)

			runCurlWithInputs([]string{"-d", "@" + tempfile.Name(), "/foo"})

			Expect(curlRepo.Body).To(Equal(`{"some":"json"}`))
			Expect(ui.Outputs()).ToNot(ContainSubstrings([]string{"FAILED"}))
		})
	})

	It("does not print the response when verbose output is enabled", func() {
		// This is to prevent the response from being printed twice

		trace.LoggingToStdout = true

		curlRepo.ResponseHeader = "Content-Size:1024"
		curlRepo.ResponseBody = "response for get"

		runCurlWithInputs([]string{"/foo"})

		Expect(ui.Outputs()).ToNot(ContainSubstrings([]string{"response for get"}))
	})

	It("prints the response even when verbose output is enabled if in a plugin call", func() {
		trace.LoggingToStdout = true

		curlRepo.ResponseHeader = "Content-Size:1024"
		curlRepo.ResponseBody = "response for get"

		runCurlAsPluginWithInputs([]string{"/foo"})

		Expect(ui.Outputs()).To(ContainSubstrings([]string{"response for get"}))
	})

	It("prints a failure message when the response is not success", func() {
		curlRepo.Error = errors.New("ooops")
		runCurlWithInputs([]string{"/foo"})

		Expect(ui.Outputs()).To(ContainSubstrings(
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

			Expect(ui.Outputs()).To(ContainSubstrings(
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

				Expect(ui.Outputs()).To(Equal([]string{"FAIL: crumpets need MOAR butterz"}))
			})
		})
	})
})
