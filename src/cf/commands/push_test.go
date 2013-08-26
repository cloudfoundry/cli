package commands_test

import (
	"cf"
	"cf/api"
	. "cf/commands"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestPushingAppWhenItDoesNotExist(t *testing.T) {
	testhelpers.Login(t)
	config, err := configuration.Load()
	assert.NoError(t, err)
	config.Organization = cf.Organization{Name: "MyOrg"}

	config.Space = cf.Space{Name: "MySpace"}

	err = config.Save()
	assert.NoError(t, err)

	domains := []cf.Domain{
		cf.Domain{Name: "foo.cf-app.com", Guid: "foo-domain-guid"},
	}
	domainRepo := &testhelpers.FakeDomainRepository{Domains: domains}
	routeRepo := &testhelpers.FakeRouteRepository{}
	appRepo := &testhelpers.FakeApplicationRepository{}

	fakeUI := callPush([]string{"--name", "my-new-app"}, appRepo, domainRepo, routeRepo)

	assert.Contains(t, fakeUI.Outputs[0], "Creating my-new-app...")
	assert.Equal(t, appRepo.CreatedApp.Name, "my-new-app")
	assert.Contains(t, fakeUI.Outputs[1], "OK")

	assert.Contains(t, fakeUI.Outputs[2], "Creating route my-new-app.foo.cf-app.com...")
	assert.Equal(t, routeRepo.CreatedRoute.Host, "my-new-app")
	assert.Equal(t, routeRepo.CreatedRouteDomain.Guid, "foo-domain-guid")
	assert.Contains(t, fakeUI.Outputs[3], "OK")

	assert.Contains(t, fakeUI.Outputs[4], "Binding my-new-app.foo.cf-app.com to my-new-app...")
	assert.Equal(t, routeRepo.BoundApp.Name, "my-new-app")
	assert.Equal(t, routeRepo.BoundRoute.Host, "my-new-app")
	assert.Contains(t, fakeUI.Outputs[5], "OK")

	assert.Contains(t, fakeUI.Outputs[6], "Uploading my-new-app...")
	assert.Equal(t, appRepo.UploadedApp.Guid, "my-new-app-guid")
	assert.Contains(t, fakeUI.Outputs[7], "OK")
}

func callPush(args []string,
	appRepo api.ApplicationRepository,
	domainRepo api.DomainRepository,
	routeRepo api.RouteRepository) (fakeUI *testhelpers.FakeUI) {

	fakeUI = new(testhelpers.FakeUI)
	target := NewPush(fakeUI, appRepo, domainRepo, routeRepo)
	target.Run(testhelpers.NewContext(4, args))
	return
}
