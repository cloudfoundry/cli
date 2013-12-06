package requirements

import (
	"cf"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testterm "testhelpers/terminal"
	"testing"
)

func TestApplicationReqExecute(t *testing.T) {
	app := cf.Application{}
	app.Name = "my-app"
	app.Guid = "my-app-guid"
	appRepo := &testapi.FakeApplicationRepository{ReadApp: app}
	ui := new(testterm.FakeUI)

	appReq := newApplicationRequirement("foo", ui, appRepo)
	success := appReq.Execute()

	assert.True(t, success)
	assert.Equal(t, appRepo.ReadName, "foo")
	assert.Equal(t, appReq.GetApplication(), app)
}

func TestApplicationReqExecuteWhenApplicationNotFound(t *testing.T) {
	appRepo := &testapi.FakeApplicationRepository{ReadNotFound: true}
	ui := new(testterm.FakeUI)

	appReq := newApplicationRequirement("foo", ui, appRepo)
	success := appReq.Execute()

	assert.False(t, success)
}
