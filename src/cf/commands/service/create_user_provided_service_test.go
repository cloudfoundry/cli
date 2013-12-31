package service_test

import (
	"cf"
	"cf/api"
	. "cf/commands/service"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestCreateUserProvidedServiceWithParameterList(t *testing.T) {
	repo := &testapi.FakeUserProvidedServiceInstanceRepo{}
	ui := callCreateUserProvidedService(t,
		[]string{"-p", `"foo, bar, baz"`, "my-custom-service"},
		[]string{"foo value", "bar value", "baz value"},
		repo,
	)

	testassert.SliceContains(t, ui.Prompts, testassert.Lines{
		{"foo"},
		{"bar"},
		{"baz"},
	})

	assert.Equal(t, repo.CreateName, "my-custom-service")
	assert.Equal(t, repo.CreateParams, map[string]string{
		"foo": "foo value",
		"bar": "bar value",
		"baz": "baz value",
	})

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Creating user provided service", "my-custom-service", "my-org", "my-space", "my-user"},
		{"OK"},
	})
}

func TestCreateUserProvidedServiceWithJson(t *testing.T) {
	repo := &testapi.FakeUserProvidedServiceInstanceRepo{}
	ui := callCreateUserProvidedService(t,
		[]string{"-p", `{"foo": "foo value", "bar": "bar value", "baz": "baz value"}`, "my-custom-service"},
		[]string{},
		repo,
	)

	assert.Empty(t, ui.Prompts)

	assert.Equal(t, repo.CreateName, "my-custom-service")
	assert.Equal(t, repo.CreateParams, map[string]string{
		"foo": "foo value",
		"bar": "bar value",
		"baz": "baz value",
	})

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Creating user provided service"},
		{"OK"},
	})
}

func TestCreateUserProvidedServiceWithNoSecondArgument(t *testing.T) {
	userProvidedServiceInstanceRepo := &testapi.FakeUserProvidedServiceInstanceRepo{}
	ui := callCreateUserProvidedService(t,
		[]string{"my-custom-service"},
		[]string{},
		userProvidedServiceInstanceRepo,
	)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Creating user provided service"},
		{"OK"},
	})
}

func TestCreateUserProvidedServiceWithSyslogDrain(t *testing.T) {
	repo := &testapi.FakeUserProvidedServiceInstanceRepo{}

	ui := callCreateUserProvidedService(t,
		[]string{"-l", "syslog://example.com", "-p", `{"foo": "foo value", "bar": "bar value", "baz": "baz value"}`, "my-custom-service"},
		[]string{},
		repo,
	)
	assert.Equal(t, repo.CreateDrainUrl, "syslog://example.com")
	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Creating user provided service"},
		{"OK"},
	})
}

func callCreateUserProvidedService(t *testing.T, args []string, inputs []string, userProvidedServiceInstanceRepo api.UserProvidedServiceInstanceRepository) (fakeUI *testterm.FakeUI) {
	fakeUI = &testterm.FakeUI{Inputs: inputs}
	ctxt := testcmd.NewContext("create-user-provided-service", args)
	reqFactory := &testreq.FakeReqFactory{}

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)
	org := cf.OrganizationFields{}
	org.Name = "my-org"
	space := cf.SpaceFields{}
	space.Name = "my-space"
	config := &configuration.Configuration{
		SpaceFields:        space,
		OrganizationFields: org,
		AccessToken:        token,
	}

	cmd := NewCreateUserProvidedService(fakeUI, config, userProvidedServiceInstanceRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
