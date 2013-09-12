package requirements_test

import (
	"cf"
	. "cf/requirements"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestApplicationReqExecute(t *testing.T) {
	app := cf.Application{Name: "my-app", Guid: "my-app-guid"}
	appRepo := &testhelpers.FakeApplicationRepository{AppByName: app}
	ui := new(testhelpers.FakeUI)

	appReq := NewApplicationRequirement("foo", ui, appRepo)
	success := appReq.Execute()

	assert.True(t, success)
	assert.Equal(t, appRepo.AppName, "foo")
	assert.Equal(t, appReq.GetApplication(), app)
}
