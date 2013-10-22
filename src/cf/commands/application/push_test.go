package application_test

import (
	"cf"
	"cf/api"
	. "cf/commands/application"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	"os"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestPushingRequirements(t *testing.T) {
	starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo := getPushDependencies()
	fakeUI := new(testterm.FakeUI)
	config := &configuration.Configuration{}
	cmd := NewPush(fakeUI, config, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo)
	ctxt := testcmd.NewContext("push", []string{})

	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}
	assert.True(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: false, TargetedSpaceSuccess: true}
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	assert.False(t, testcmd.CommandDidPassRequirements)

	testcmd.CommandDidPassRequirements = true

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: false}
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	assert.False(t, testcmd.CommandDidPassRequirements)
}

func TestPushingAppWhenItDoesNotExist(t *testing.T) {
	starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo := getPushDependencies()

	domains := []cf.Domain{
		cf.Domain{Name: "foo.cf-app.com", Guid: "foo-domain-guid"},
	}

	domainRepo.FindByNameDomain = domains[0]
	routeRepo.FindByHostErr = true
	appRepo.FindByNameNotFound = true
	stopper.StoppedApp = cf.Application{Name: "my-stopped-app"}

	fakeUI := callPush(t, []string{"my-new-app"}, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo)

	assert.Contains(t, fakeUI.Outputs[0], "Creating app")
	assert.Contains(t, fakeUI.Outputs[0], "my-new-app")
	assert.Contains(t, fakeUI.Outputs[0], "my-org")
	assert.Contains(t, fakeUI.Outputs[0], "my-space")
	assert.Contains(t, fakeUI.Outputs[0], "my-user")
	assert.Equal(t, appRepo.CreatedApp.Name, "my-new-app")
	assert.Equal(t, appRepo.CreatedApp.Instances, 1)
	assert.Equal(t, appRepo.CreatedApp.Memory, uint64(128))
	assert.Equal(t, appRepo.CreatedApp.BuildpackUrl, "")
	assert.Contains(t, fakeUI.Outputs[1], "OK")

	assert.Contains(t, fakeUI.Outputs[3], "my-new-app.foo.cf-app.com")
	assert.Equal(t, routeRepo.FindByHostHost, "my-new-app")
	assert.Equal(t, routeRepo.CreatedRoute.Host, "my-new-app")
	assert.Equal(t, routeRepo.CreatedRouteDomain.Guid, "foo-domain-guid")
	assert.Contains(t, fakeUI.Outputs[4], "OK")

	assert.Contains(t, fakeUI.Outputs[6], "my-new-app.foo.cf-app.com")
	assert.Equal(t, routeRepo.BoundApp.Name, "my-new-app")
	assert.Equal(t, routeRepo.BoundRoute.Host, "my-new-app")
	assert.Contains(t, fakeUI.Outputs[7], "OK")

	expectedAppDir, err := os.Getwd()
	assert.NoError(t, err)

	assert.Contains(t, fakeUI.Outputs[9], "my-new-app")
	assert.Equal(t, appBitsRepo.UploadedApp.Guid, "my-new-app-guid")
	assert.Equal(t, appBitsRepo.UploadedDir, expectedAppDir)
	assert.Contains(t, fakeUI.Outputs[10], "OK")

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
	appRepo.FindByNameNotFound = true

	fakeUI := callPush(t, []string{"my-new-app"}, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo)

	assert.Empty(t, routeRepo.CreatedRoute.Host)
	assert.Empty(t, routeRepo.CreatedRouteDomain.Guid)
	assert.Contains(t, fakeUI.Outputs[3], "my-new-app.foo.cf-app.com")
	assert.Equal(t, routeRepo.FindByHostHost, "my-new-app")

	assert.Contains(t, fakeUI.Outputs[4], "my-new-app.foo.cf-app.com")
	assert.Equal(t, routeRepo.BoundApp.Name, "my-new-app")
	assert.Equal(t, routeRepo.BoundRoute.Host, "my-new-app")
	assert.Contains(t, fakeUI.Outputs[5], "OK")
}

func TestPushingAppWithCustomFlags(t *testing.T) {
	starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo := getPushDependencies()

	domain := cf.Domain{Name: "bar.cf-app.com", Guid: "bar-domain-guid"}
	stack := cf.Stack{Name: "customLinux", Guid: "custom-linux-guid"}

	domainRepo.FindByNameDomain = domain
	routeRepo.FindByHostErr = true
	stackRepo.FindByNameStack = stack
	appRepo.FindByNameNotFound = true

	fakeUI := callPush(t, []string{
		"-c", "unicorn -c config/unicorn.rb -D",
		"-d", "bar.cf-app.com",
		"-n", "my-hostname",
		"-i", "3",
		"-m", "2G",
		"-b", "https://github.com/heroku/heroku-buildpack-play.git",
		"-p", "/Users/pivotal/workspace/my-new-app",
		"-s", "customLinux",
		"--no-start",
		"my-new-app",
	}, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo)

	assert.Contains(t, fakeUI.Outputs[0], "customLinux")
	assert.Equal(t, stackRepo.FindByNameName, "customLinux")

	assert.Contains(t, fakeUI.Outputs[1], "my-new-app")
	assert.Equal(t, appRepo.CreatedApp.Name, "my-new-app")
	assert.Equal(t, appRepo.CreatedApp.Command, "unicorn -c config/unicorn.rb -D")
	assert.Equal(t, appRepo.CreatedApp.Instances, 3)
	assert.Equal(t, appRepo.CreatedApp.Memory, uint64(2048))
	assert.Equal(t, appRepo.CreatedApp.Stack.Guid, "custom-linux-guid")
	assert.Equal(t, appRepo.CreatedApp.BuildpackUrl, "https://github.com/heroku/heroku-buildpack-play.git")
	assert.Contains(t, fakeUI.Outputs[2], "OK")

	assert.Contains(t, fakeUI.Outputs[4], "my-hostname.bar.cf-app.com")
	assert.Equal(t, domainRepo.FindByNameName, "bar.cf-app.com")
	assert.Equal(t, routeRepo.CreatedRoute.Host, "my-hostname")
	assert.Equal(t, routeRepo.CreatedRouteDomain.Guid, "bar-domain-guid")
	assert.Contains(t, fakeUI.Outputs[5], "OK")

	assert.Contains(t, fakeUI.Outputs[7], "my-hostname.bar.cf-app.com")
	assert.Contains(t, fakeUI.Outputs[7], "my-new-app")
	assert.Equal(t, routeRepo.BoundApp.Name, "my-new-app")
	assert.Equal(t, routeRepo.BoundRoute.Host, "my-hostname")
	assert.Contains(t, fakeUI.Outputs[8], "OK")

	assert.Contains(t, fakeUI.Outputs[10], "my-new-app")
	assert.Equal(t, appBitsRepo.UploadedApp.Guid, "my-new-app-guid")
	assert.Equal(t, appBitsRepo.UploadedDir, "/Users/pivotal/workspace/my-new-app")
	assert.Contains(t, fakeUI.Outputs[11], "OK")

	assert.Equal(t, starter.AppToStart.Name, "")
}

func TestPushingAppWithNoRoute(t *testing.T) {
	starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo := getPushDependencies()

	domain := cf.Domain{Name: "bar.cf-app.com", Guid: "bar-domain-guid"}
	stack := cf.Stack{Name: "customLinux", Guid: "custom-linux-guid"}

	domainRepo.FindByNameDomain = domain
	routeRepo.FindByHostErr = true
	stackRepo.FindByNameStack = stack
	appRepo.FindByNameNotFound = true

	fakeUI := callPush(t, []string{
		"--no-route",
		"my-new-app",
	}, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo)

	assert.Contains(t, fakeUI.Outputs[0], "my-new-app")
	assert.Equal(t, appRepo.CreatedApp.Name, "my-new-app")
	assert.Equal(t, routeRepo.CreatedRoute.Host, "")
	assert.Equal(t, routeRepo.CreatedRouteDomain.Guid, "")
}

func TestPushingAppWithMemoryInMegaBytes(t *testing.T) {
	starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo := getPushDependencies()

	domain := cf.Domain{Name: "bar.cf-app.com", Guid: "bar-domain-guid"}
	domainRepo.FindByNameDomain = domain
	appRepo.FindByNameNotFound = true

	callPush(t, []string{
		"-m", "256M",
		"my-new-app",
	}, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo)

	assert.Equal(t, appRepo.CreatedApp.Memory, uint64(256))
}

func TestPushingAppWithMemoryWithoutUnit(t *testing.T) {
	starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo := getPushDependencies()

	domain := cf.Domain{Name: "bar.cf-app.com", Guid: "bar-domain-guid"}
	domainRepo.FindByNameDomain = domain
	appRepo.FindByNameNotFound = true

	callPush(t, []string{
		"-m", "512",
		"my-new-app",
	}, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo)

	assert.Equal(t, appRepo.CreatedApp.Memory, uint64(512))
}

func TestPushingAppWithInvalidMemory(t *testing.T) {
	starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo := getPushDependencies()

	domain := cf.Domain{Name: "bar.cf-app.com", Guid: "bar-domain-guid"}
	domainRepo.FindByNameDomain = domain
	appRepo.FindByNameNotFound = true

	callPush(t, []string{
		"-m", "abcM",
		"my-new-app",
	}, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo)

	assert.Equal(t, appRepo.CreatedApp.Memory, uint64(128))
}

func TestPushingAppWhenItAlreadyExists(t *testing.T) {
	starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo := getPushDependencies()

	existingApp := cf.Application{Name: "existing-app", Guid: "existing-app-guid"}
	appRepo.FindByNameApp = existingApp

	fakeUI := callPush(t, []string{"existing-app"}, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo)

	assert.Equal(t, stopper.AppToStop.Name, "existing-app")
	assert.Contains(t, fakeUI.Outputs[0], "existing-app")
	assert.Equal(t, appBitsRepo.UploadedApp.Guid, "existing-app-guid")
	assert.Contains(t, fakeUI.Outputs[1], "OK")
}

func TestPushingAppWithInvalidPath(t *testing.T) {
	starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo := getPushDependencies()
	appBitsRepo.UploadAppErr = true
	appRepo.FindByNameApp = cf.Application{Name: "app", Guid: "app-guid"}

	fakeUI := callPush(t, []string{"app"}, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo)

	assert.Equal(t, 3, len(fakeUI.Outputs))
	assert.Contains(t, fakeUI.Outputs[0], "Uploading")
	assert.Contains(t, fakeUI.Outputs[1], "FAILED")
}

func getPushDependencies() (starter *testcmd.FakeAppStarter,
	stopper *testcmd.FakeAppStopper,
	appRepo *testapi.FakeApplicationRepository,
	domainRepo *testapi.FakeDomainRepository,
	routeRepo *testapi.FakeRouteRepository,
	stackRepo *testapi.FakeStackRepository,
	appBitsRepo *testapi.FakeApplicationBitsRepository) {

	starter = &testcmd.FakeAppStarter{}
	stopper = &testcmd.FakeAppStopper{}
	appRepo = &testapi.FakeApplicationRepository{}
	domainRepo = &testapi.FakeDomainRepository{}
	routeRepo = &testapi.FakeRouteRepository{}
	stackRepo = &testapi.FakeStackRepository{}
	appBitsRepo = &testapi.FakeApplicationBitsRepository{}

	return
}

func callPush(t *testing.T,
	args []string,
	starter ApplicationStarter,
	stopper ApplicationStopper,
	appRepo api.ApplicationRepository,
	domainRepo api.DomainRepository,
	routeRepo api.RouteRepository,
	stackRepo api.StackRepository,
	appBitsRepo *testapi.FakeApplicationBitsRepository) (fakeUI *testterm.FakeUI) {

	fakeUI = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("push", args)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)

	config := &configuration.Configuration{
		Space:        cf.Space{Name: "my-space"},
		Organization: cf.Organization{Name: "my-org"},
		AccessToken:  token,
	}

	cmd := NewPush(fakeUI, config, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo)
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}
	testcmd.RunCommand(cmd, ctxt, reqFactory)

	return
}
