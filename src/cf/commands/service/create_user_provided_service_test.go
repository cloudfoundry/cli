package service_test

import (
	"cf/api"
	. "cf/commands/service"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

func callCreateUserProvidedService(t mr.TestingT, args []string, inputs []string, userProvidedServiceInstanceRepo api.UserProvidedServiceInstanceRepository) (fakeUI *testterm.FakeUI) {
	fakeUI = &testterm.FakeUI{Inputs: inputs}
	ctxt := testcmd.NewContext("create-user-provided-service", args)
	reqFactory := &testreq.FakeReqFactory{}

	config := testconfig.NewRepositoryWithDefaults()

	cmd := NewCreateUserProvidedService(fakeUI, config, userProvidedServiceInstanceRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestCreateUserProvidedServiceWithParameterList", func() {
			repo := &testapi.FakeUserProvidedServiceInstanceRepo{}
			ui := callCreateUserProvidedService(mr.T(), []string{"-p", `"foo, bar, baz"`, "my-custom-service"},
				[]string{"foo value", "bar value", "baz value"},
				repo,
			)

			testassert.SliceContains(mr.T(), ui.Prompts, testassert.Lines{
				{"foo"},
				{"bar"},
				{"baz"},
			})

			assert.Equal(mr.T(), repo.CreateName, "my-custom-service")
			assert.Equal(mr.T(), repo.CreateParams, map[string]string{
				"foo": "foo value",
				"bar": "bar value",
				"baz": "baz value",
			})

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Creating user provided service", "my-custom-service", "my-org", "my-space", "my-user"},
				{"OK"},
			})
		})
		It("TestCreateUserProvidedServiceWithJson", func() {

			repo := &testapi.FakeUserProvidedServiceInstanceRepo{}
			ui := callCreateUserProvidedService(mr.T(), []string{"-p", `{"foo": "foo value", "bar": "bar value", "baz": "baz value"}`, "my-custom-service"},
				[]string{},
				repo,
			)

			assert.Empty(mr.T(), ui.Prompts)

			assert.Equal(mr.T(), repo.CreateName, "my-custom-service")
			assert.Equal(mr.T(), repo.CreateParams, map[string]string{
				"foo": "foo value",
				"bar": "bar value",
				"baz": "baz value",
			})

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Creating user provided service"},
				{"OK"},
			})
		})
		It("TestCreateUserProvidedServiceWithNoSecondArgument", func() {

			userProvidedServiceInstanceRepo := &testapi.FakeUserProvidedServiceInstanceRepo{}
			ui := callCreateUserProvidedService(mr.T(), []string{"my-custom-service"},
				[]string{},
				userProvidedServiceInstanceRepo,
			)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Creating user provided service"},
				{"OK"},
			})
		})
		It("TestCreateUserProvidedServiceWithSyslogDrain", func() {

			repo := &testapi.FakeUserProvidedServiceInstanceRepo{}

			ui := callCreateUserProvidedService(mr.T(), []string{"-l", "syslog://example.com", "-p", `{"foo": "foo value", "bar": "bar value", "baz": "baz value"}`, "my-custom-service"},
				[]string{},
				repo,
			)
			assert.Equal(mr.T(), repo.CreateDrainUrl, "syslog://example.com")
			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Creating user provided service"},
				{"OK"},
			})
		})
	})
}
