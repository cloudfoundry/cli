package requirements

import (
	"cf"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testterm "testhelpers/terminal"
	"testing"
)

func TestApplicationReqExecute(t *testing.T) {
	app := cf.Application{Name: "my-app", Guid: "my-app-guid"}
	appRepo := &testapi.FakeApplicationRepository{FindByNameApp: app}
	ui := new(testterm.FakeUI)

	appReq := NewApplicationRequirement("foo", ui, appRepo)
	success := appReq.Execute()

	assert.True(t, success)
	assert.Equal(t, appRepo.FindByNameName, "foo")
	assert.Equal(t, appReq.GetApplication(), app)
}

func TestApplicationReqExecuteWhenApplicationNotFound(t *testing.T) {
	appRepo := &testapi.FakeApplicationRepository{FindByNameNotFound: true}
	ui := new(testterm.FakeUI)

	appReq := NewApplicationRequirement("foo", ui, appRepo)
	success := appReq.Execute()

	assert.False(t, success)
}
