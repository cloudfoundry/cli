package requirements_test

import (
	"cf"
	"cf/configuration"
	. "cf/requirements"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestApplicationReqExecute(t *testing.T) {
	app := cf.Application{Name: "my-app", Guid: "my-app-guid"}
	appRepo := &testhelpers.FakeApplicationRepository{AppByName: app}
	config := &configuration.Configuration{}
	ui := new(testhelpers.FakeUI)

	appReq := NewApplicationRequirement("foo", ui, config, appRepo)
	err := appReq.Execute()

	assert.NoError(t, err)
	assert.Equal(t, appRepo.AppName, "foo")
	assert.Equal(t, appReq.GetApplication(), app)
}
