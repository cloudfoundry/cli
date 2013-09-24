package commands_test

import (
	"cf"
	"cf/api"
	. "cf/commands"
	"github.com/stretchr/testify/assert"
	"os"
	"testhelpers"
	"testing"
)

func TestPushingRequirements(t *testing.T) {
	starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo := getPushDependencies()
	fakeUI := new(testhelpers.FakeUI)
	cmd := NewPush(fakeUI, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo)
	ctxt := testhelpers.NewContext("push", []string{})

	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}
	testhelpers.RunCommand(cmd, ctxt, reqFactory)
	assert.True(t, testhelpers.CommandDidPassRequirements)

	reqFactory = &testhelpers.FakeReqFactory{LoginSuccess: false, TargetedSpaceSuccess: true}
	testhelpers.RunCommand(cmd, ctxt, reqFactory)
	assert.False(t, testhelpers.CommandDidPassRequirements)

	testhelpers.CommandDidPassRequirements = true

	reqFactory = &testhelpers.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: false}
	testhelpers.RunCommand(cmd, ctxt, reqFactory)
	assert.False(t, testhelpers.CommandDidPassRequirements)
}

func TestPushingAppWhenItDoesNotExist(t *testing.T) {
	starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo := getPushDependencies()

	domains := []cf.Domain{
		cf.Domain{Name: "foo.cf-app.com", Guid: "foo-domain-guid"},
	}

	domainRepo.FindByNameDomain = domains[0]
	routeRepo.FindByHostErr = true
	appRepo.AppByNameErr = true
	stopper.StoppedApp = cf.Application{Name: "my-stopped-app"}

	fakeUI := callPush([]string{"--name", "my-new-app"}, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo)

	assert.Contains(t, fakeUI.Outputs[0], "my-new-app")
	assert.Equal(t, appRepo.CreatedApp.Name, "my-new-app")
	assert.Equal(t, appRepo.CreatedApp.Instances, 1)
	assert.Equal(t, appRepo.CreatedApp.Memory, 128)
	assert.Equal(t, appRepo.CreatedApp.BuildpackUrl, "")
	assert.Contains(t, fakeUI.Outputs[1], "OK")

	assert.Contains(t, fakeUI.Outputs[2], "my-new-app.foo.cf-app.com")
	assert.Equal(t, routeRepo.FindByHostHost, "my-new-app")
	assert.Equal(t, routeRepo.CreatedRoute.Host, "my-new-app")
	assert.Equal(t, routeRepo.CreatedRouteDomain.Guid, "foo-domain-guid")
	assert.Contains(t, fakeUI.Outputs[3], "OK")

	assert.Contains(t, fakeUI.Outputs[4], "my-new-app.foo.cf-app.com")
	assert.Equal(t, routeRepo.BoundApp.Name, "my-new-app")
	assert.Equal(t, routeRepo.BoundRoute.Host, "my-new-app")
	assert.Contains(t, fakeUI.Outputs[5], "OK")

	expectedAppDir, err := os.Getwd()
	assert.NoError(t, err)

	assert.Contains(t, fakeUI.Outputs[6], "my-new-app")
	assert.Equal(t, appBitsRepo.UploadedApp.Guid, "my-new-app-guid")
	assert.Equal(t, appBitsRepo.UploadedDir, expectedAppDir)
	assert.Contains(t, fakeUI.Outputs[7], "OK")

	assert.Equal(t, stopper.AppToStop.Name, "my-new-app")
	assert.Equal(t, starter.AppToStart.Name, "my-stopped-app")
}

func TestPushingAppWhenItDoesNotExistButRouteExists(t *testing.T) {
	starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo := getPushDependencies()

	domains := []cf.Domain{
		cf.Domain{Name: "foo.cf-app.com", Guid: "foo-domain-guid"},
	}
	route := cf.Route{Host: "my-new-app"}

	domainRepo.FindByNameDomain = domains[0]
	routeRepo.FindByHostRoute = route
	appRepo.AppByNameErr = true

	fakeUI := callPush([]string{"--name", "my-new-app"}, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo)

	assert.Empty(t, routeRepo.CreatedRoute.Host)
	assert.Empty(t, routeRepo.CreatedRouteDomain.Guid)
	assert.Contains(t, fakeUI.Outputs[2], "my-new-app.foo.cf-app.com")
	assert.Equal(t, routeRepo.FindByHostHost, "my-new-app")

	assert.Contains(t, fakeUI.Outputs[3], "my-new-app.foo.cf-app.com")
	assert.Equal(t, routeRepo.BoundApp.Name, "my-new-app")
	assert.Equal(t, routeRepo.BoundRoute.Host, "my-new-app")
	assert.Contains(t, fakeUI.Outputs[4], "OK")
}

func TestPushingAppWithCustomFlags(t *testing.T) {
	starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo := getPushDependencies()

	domain := cf.Domain{Name: "bar.cf-app.com", Guid: "bar-domain-guid"}
	stack := cf.Stack{Name: "customLinux", Guid: "custom-linux-guid"}

	domainRepo.FindByNameDomain = domain
	routeRepo.FindByHostErr = true
	appRepo.AppByNameErr = true
	stackRepo.FindByNameStack = stack

	fakeUI := callPush([]string{
		"--name", "my-new-app",
		"--domain", "bar.cf-app.com",
		"--host", "my-hostname",
		"--instances", "3",
		"--memory", "2G",
		"--buildpack", "https://github.com/heroku/heroku-buildpack-play.git",
		"--path", "/Users/pivotal/workspace/my-new-app",
		"--stack", "customLinux",
		"--no-start", "",
	}, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo)

	assert.Contains(t, fakeUI.Outputs[0], "customLinux")
	assert.Equal(t, stackRepo.FindByNameName, "customLinux")

	assert.Contains(t, fakeUI.Outputs[1], "my-new-app")
	assert.Equal(t, appRepo.CreatedApp.Name, "my-new-app")
	assert.Equal(t, appRepo.CreatedApp.Instances, 3)
	assert.Equal(t, appRepo.CreatedApp.Memory, 2048)
	assert.Equal(t, appRepo.CreatedApp.Stack.Guid, "custom-linux-guid")
	assert.Equal(t, appRepo.CreatedApp.BuildpackUrl, "https://github.com/heroku/heroku-buildpack-play.git")
	assert.Contains(t, fakeUI.Outputs[2], "OK")

	assert.Contains(t, fakeUI.Outputs[3], "my-hostname.bar.cf-app.com")
	assert.Equal(t, domainRepo.FindByNameName, "bar.cf-app.com")
	assert.Equal(t, routeRepo.CreatedRoute.Host, "my-hostname")
	assert.Equal(t, routeRepo.CreatedRouteDomain.Guid, "bar-domain-guid")
	assert.Contains(t, fakeUI.Outputs[4], "OK")

	assert.Contains(t, fakeUI.Outputs[5], "my-hostname.bar.cf-app.com")
	assert.Contains(t, fakeUI.Outputs[5], "my-new-app")
	assert.Equal(t, routeRepo.BoundApp.Name, "my-new-app")
	assert.Equal(t, routeRepo.BoundRoute.Host, "my-hostname")
	assert.Contains(t, fakeUI.Outputs[6], "OK")

	assert.Contains(t, fakeUI.Outputs[7], "my-new-app")
	assert.Equal(t, appBitsRepo.UploadedApp.Guid, "my-new-app-guid")
	assert.Equal(t, appBitsRepo.UploadedDir, "/Users/pivotal/workspace/my-new-app")
	assert.Contains(t, fakeUI.Outputs[8], "OK")

	assert.Equal(t, starter.AppToStart.Name, "")
}

func TestPushingAppWithMemoryInMegaBytes(t *testing.T) {
	starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo := getPushDependencies()

	domain := cf.Domain{Name: "bar.cf-app.com", Guid: "bar-domain-guid"}
	domainRepo.FindByNameDomain = domain
	appRepo.AppByNameErr = true

	callPush([]string{
		"--name", "my-new-app",
		"--memory", "256M",
	}, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo)

	assert.Equal(t, appRepo.CreatedApp.Memory, 256)
}

func TestPushingAppWithMemoryWithoutUnit(t *testing.T) {
	starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo := getPushDependencies()

	domain := cf.Domain{Name: "bar.cf-app.com", Guid: "bar-domain-guid"}
	domainRepo.FindByNameDomain = domain
	appRepo.AppByNameErr = true

	callPush([]string{
		"--name", "my-new-app",
		"--memory", "512",
	}, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo)

	assert.Equal(t, appRepo.CreatedApp.Memory, 512)
}

func TestPushingAppWithInvalidMemory(t *testing.T) {
	starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo := getPushDependencies()

	domain := cf.Domain{Name: "bar.cf-app.com", Guid: "bar-domain-guid"}
	domainRepo.FindByNameDomain = domain
	appRepo.AppByNameErr = true

	callPush([]string{
		"--name", "my-new-app",
		"--memory", "abcM",
	}, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo)

	assert.Equal(t, appRepo.CreatedApp.Memory, 128)
}

func TestPushingAppWhenItAlreadyExists(t *testing.T) {
	starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo := getPushDependencies()

	existingApp := cf.Application{Name: "existing-app", Guid: "existing-app-guid"}
	appRepo.AppByName = existingApp

	fakeUI := callPush([]string{"--name", "existing-app"}, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo)

	assert.Equal(t, stopper.AppToStop.Name, "existing-app")
	assert.Contains(t, fakeUI.Outputs[0], "existing-app")
	assert.Equal(t, appBitsRepo.UploadedApp.Guid, "existing-app-guid")
	assert.Contains(t, fakeUI.Outputs[1], "OK")
}

func getPushDependencies() (starter *testhelpers.FakeAppStarter,
	stopper *testhelpers.FakeAppStopper,
	appRepo *testhelpers.FakeApplicationRepository,
	domainRepo *testhelpers.FakeDomainRepository,
	routeRepo *testhelpers.FakeRouteRepository,
	stackRepo *testhelpers.FakeStackRepository,
	appBitsRepo *testhelpers.FakeApplicationBitsRepository) {

	starter = &testhelpers.FakeAppStarter{}
	stopper = &testhelpers.FakeAppStopper{}
	appRepo = &testhelpers.FakeApplicationRepository{}
	domainRepo = &testhelpers.FakeDomainRepository{}
	routeRepo = &testhelpers.FakeRouteRepository{}
	stackRepo = &testhelpers.FakeStackRepository{}
	appBitsRepo = &testhelpers.FakeApplicationBitsRepository{}

	return
}

func callPush(args []string,
	starter ApplicationStarter,
	stopper ApplicationStopper,
	appRepo api.ApplicationRepository,
	domainRepo api.DomainRepository,
	routeRepo api.RouteRepository,
	stackRepo api.StackRepository,
	appBitsRepo *testhelpers.FakeApplicationBitsRepository) (fakeUI *testhelpers.FakeUI) {

	fakeUI = new(testhelpers.FakeUI)
	ctxt := testhelpers.NewContext("push", args)
	cmd := NewPush(fakeUI, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo)
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}
	testhelpers.RunCommand(cmd, ctxt, reqFactory)

	return
}
