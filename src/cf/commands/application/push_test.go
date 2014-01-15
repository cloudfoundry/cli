package application_test

import (
	"cf"
	. "cf/commands/application"
	"cf/configuration"
	"cf/manifest"
	"errors"
	"generic"
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"strings"
	"syscall"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	"testhelpers/maker"
	testmanifest "testhelpers/manifest"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestPushingRequirements(t *testing.T) {
	ui := new(testterm.FakeUI)
	config := &configuration.Configuration{}
	deps := getPushDependencies()
	manifestRepo := deps.manifestRepo
	starter := deps.starter
	stopper := deps.stopper
	binder := deps.binder
	appRepo := deps.appRepo
	domainRepo := deps.domainRepo
	routeRepo := deps.routeRepo
	stackRepo := deps.stackRepo
	appBitsRepo := deps.appBitsRepo
	serviceRepo := deps.serviceRepo

	cmd := NewPush(ui, config, manifestRepo, starter, stopper, binder, appRepo, domainRepo, routeRepo, stackRepo, serviceRepo, appBitsRepo)
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
	deps := getPushDependencies()

	domain := cf.Domain{}
	domain.Guid = "not-the-right-guid"
	domain.Name = "not shared domain"
	domain.OwningOrganizationGuid = "my-org-guid"

	sharedDomain := cf.Domain{}
	sharedDomain.Name = "foo.cf-app.com"
	sharedDomain.Shared = true
	sharedDomain.Guid = "foo-domain-guid"
	appRepo := deps.appRepo
	domainRepo := deps.domainRepo
	routeRepo := deps.routeRepo
	appBitsRepo := deps.appBitsRepo
	stopper := deps.stopper
	starter := deps.starter

	domainRepo.ListDomainsForOrgDomains = []cf.Domain{domain, sharedDomain}
	routeRepo.FindByHostAndDomainErr = true

	appRepo.ReadNotFound = true

	ui := callPush(t, []string{"-t", "111", "my-new-app"}, deps)
	assert.Equal(t, appRepo.CreatedAppParams().Get("name").(string), "my-new-app")
	assert.Equal(t, appRepo.CreatedAppParams().Get("space_guid").(string), "my-space-guid")

	assert.Equal(t, routeRepo.FindByHostAndDomainHost, "my-new-app")
	assert.Equal(t, routeRepo.CreatedHost, "my-new-app")
	assert.Equal(t, routeRepo.CreatedDomainGuid, "foo-domain-guid")
	assert.Equal(t, routeRepo.BoundAppGuid, "my-new-app-guid")
	assert.Equal(t, routeRepo.BoundRouteGuid, "my-new-app-route-guid")

	assert.Equal(t, appBitsRepo.UploadedAppGuid, "my-new-app-guid")

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Creating app", "my-new-app", "my-org", "my-space"},
		{"OK"},
		{"Creating", "my-new-app.foo.cf-app.com"},
		{"OK"},
		{"Binding", "my-new-app.foo.cf-app.com"},
		{"OK"},
		{"Uploading my-new-app"},
		{"OK"},
	})

	assert.Equal(t, stopper.AppToStop.Guid, "")
	assert.Equal(t, starter.AppToStart.Guid, "my-new-app-guid")
	assert.Equal(t, starter.AppToStart.Name, "my-new-app")
	assert.Equal(t, starter.Timeout, 111)
}

func TestPushingAppWhenItDoesNotExistButRouteExists(t *testing.T) {
	deps := getPushDependencies()

	domain := cf.Domain{}
	domain.Name = "foo.cf-app.com"
	domain.Guid = "foo-domain-guid"
	domain.Shared = true

	route := cf.Route{}
	route.Guid = "my-route-guid"
	route.Host = "my-new-app"
	route.Domain = domain.DomainFields

	deps.domainRepo.ListDomainsForOrgDomains = []cf.Domain{domain}

	deps.routeRepo.FindByHostAndDomainRoute = route
	deps.appRepo.ReadNotFound = true

	ui := callPush(t, []string{"my-new-app"}, deps)

	assert.Empty(t, deps.routeRepo.CreatedHost)
	assert.Empty(t, deps.routeRepo.CreatedDomainGuid)
	assert.Equal(t, deps.routeRepo.FindByHostAndDomainHost, "my-new-app")
	assert.Equal(t, deps.routeRepo.BoundAppGuid, "my-new-app-guid")
	assert.Equal(t, deps.routeRepo.BoundRouteGuid, "my-route-guid")

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Using", "my-new-app.foo.cf-app.com"},
		{"Binding", "my-new-app.foo.cf-app.com"},
		{"OK"},
	})
}

func TestPushingAppWithCustomFlags(t *testing.T) {
	deps := getPushDependencies()
	domain := cf.Domain{}
	domain.Name = "bar.cf-app.com"
	domain.Guid = "bar-domain-guid"
	stack := cf.Stack{}
	stack.Name = "customLinux"
	stack.Guid = "custom-linux-guid"

	deps.domainRepo.FindByNameDomain = domain
	deps.routeRepo.FindByHostAndDomainErr = true
	deps.stackRepo.FindByNameStack = stack
	deps.appRepo.ReadNotFound = true
	deps.manifestRepo.ManifestDir = "/path/to/my-new-app"

	ui := callPush(t, []string{
		"-c", "unicorn -c config/unicorn.rb -D",
		"-d", "bar.cf-app.com",
		"-n", "my-hostname",
		"-i", "3",
		"-m", "2G",
		"-b", "https://github.com/heroku/heroku-buildpack-play.git",
		"-p", "/path/to/my-new-app",
		"-s", "customLinux",
		"-t", "1",
		"--no-start",
		"my-new-app",
	}, deps)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Using", "customLinux"},
		{"OK"},
		{"Creating app", "my-new-app"},
		{"OK"},
		{"Creating Route", "my-hostname.bar.cf-app.com"},
		{"OK"},
		{"Binding", "my-hostname.bar.cf-app.com", "my-new-app"},
		{"Uploading", "my-new-app"},
		{"OK"},
	})

	assert.Equal(t, deps.stackRepo.FindByNameName, "customLinux")

	assert.Equal(t, deps.appRepo.CreatedAppParams().Get("name").(string), "my-new-app")
	assert.Equal(t, deps.appRepo.CreatedAppParams().Get("command").(string), "unicorn -c config/unicorn.rb -D")
	assert.Equal(t, deps.appRepo.CreatedAppParams().Get("instances").(int), 3)
	assert.Equal(t, deps.appRepo.CreatedAppParams().Get("memory").(uint64), uint64(2048))
	assert.Equal(t, deps.appRepo.CreatedAppParams().Get("stack_guid"), "custom-linux-guid")
	assert.Equal(t, deps.appRepo.CreatedAppParams().Get("health_check_timeout").(int), 1)
	assert.Equal(t, deps.appRepo.CreatedAppParams().Get("buildpack"), "https://github.com/heroku/heroku-buildpack-play.git")

	assert.Equal(t, deps.domainRepo.FindByNameInCurrentSpaceName, "bar.cf-app.com")

	assert.Equal(t, deps.routeRepo.CreatedHost, "my-hostname")
	assert.Equal(t, deps.routeRepo.CreatedDomainGuid, "bar-domain-guid")
	assert.Equal(t, deps.routeRepo.BoundAppGuid, "my-new-app-guid")
	assert.Equal(t, deps.routeRepo.BoundRouteGuid, "my-hostname-route-guid")

	assert.Equal(t, deps.appBitsRepo.UploadedAppGuid, "my-new-app-guid")
	assert.Equal(t, deps.appBitsRepo.UploadedDir, "/path/to/my-new-app")

	assert.Equal(t, deps.starter.AppToStart.Name, "")
}

func TestPushingAppWithInvalidTimeout(t *testing.T) {
	deps := getPushDependencies()
	deps.appRepo.ReadNotFound = true

	ui := callPush(t, []string{
		"-t", "FooeyTimeout",
		"my-new-app",
	}, deps)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"FAILED"},
		{"invalid", "timeout"},
	})
}

func TestPushingAppToResetStartCommand(t *testing.T) {
	deps := getPushDependencies()

	existingApp := cf.Application{}
	existingApp.Name = "existing-app"
	existingApp.Guid = "existing-app-guid"
	existingApp.Command = "unicorn -c config/unicorn.rb -D"

	deps.appRepo.ReadApp = existingApp

	args := []string{
		"-c", "null",
		"existing-app",
	}
	_ = callPush(t, args, deps)

	assert.Equal(t, deps.appRepo.UpdateParams.Get("command"), "null")
}

func TestPushingAppWithSingleAppManifest(t *testing.T) {
	deps := getPushDependencies()
	domain := cf.Domain{}
	domain.Name = "manifest-example.com"
	domain.Guid = "bar-domain-guid"
	deps.domainRepo.FindByNameDomain = domain
	deps.routeRepo.FindByHostAndDomainErr = true
	deps.appRepo.ReadNotFound = true

	m, errs := manifest.ParseToManifest(strings.NewReader(maker.ManifestWithName("single app")))
	testassert.AssertNoErrors(t, errs)
	deps.manifestRepo.ReadManifestManifest = m

	ui := callPush(t, []string{}, deps)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Creating route", "manifest-host.manifest-example.com"},
		{"OK"},
		{"Binding", "manifest-host.manifest-example.com"},
		{"manifest-app-name"},
	})

	assert.Equal(t, deps.appRepo.CreatedAppParams().Get("name").(string), "manifest-app-name")
	assert.Equal(t, deps.appRepo.CreatedAppParams().Get("memory").(uint64), uint64(128))
	assert.Equal(t, deps.appRepo.CreatedAppParams().Get("instances").(int), 1)
	assert.Equal(t, deps.appRepo.CreatedAppParams().Get("stack").(string), "custom-stack")
	assert.Equal(t, deps.appRepo.CreatedAppParams().Get("buildpack").(string), "some-buildpack")
	assert.Equal(t, deps.appRepo.CreatedAppParams().Get("command").(string), "JAVA_HOME=$PWD/.openjdk JAVA_OPTS=\"-Xss995K\" ./bin/start.sh run")
	assert.Equal(t, deps.appRepo.CreatedAppParams().Get("path").(string), filepath.Clean("../../fixtures/example-app"))

	assert.True(t, deps.appRepo.CreatedAppParams().Has("env"))
	envVars := deps.appRepo.CreatedAppParams().Get("env").(generic.Map)

	assert.Equal(t, 2, envVars.Count())
	assert.True(t, envVars.Has("PATH"))
	assert.True(t, envVars.Has("FOO"))

	assert.Equal(t, envVars.Get("PATH").(string), "/u/apps/my-app/bin")
	assert.Equal(t, envVars.Get("FOO").(string), "baz")
}

func TestPushingAppManifestWithNulls(t *testing.T) {
	deps := getPushDependencies()
	deps.appRepo.ReadNotFound = true

	m, errs := manifest.ParseToManifest(strings.NewReader(maker.ManifestWithName("nulls")))
	deps.manifestRepo.ReadManifestManifest = m
	deps.manifestRepo.ReadManifestErrors = errs

	ui := callPush(t, []string{}, deps)
	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"FAILED"},
		{"Error", "reading", "manifest"},
		{"command", "should not be null"},
		{"space_guid", "should not be null"},
		{"buildpack", "should not be null"},
		{"disk_quota", "should not be null"},
		{"instances", "should not be null"},
		{"memory", "should not be null"},
		{"env", "should not be null"},
	})
}

func TestPushingManyAppsFromManifest(t *testing.T) {
	deps := getPushDependencies()
	deps.appRepo.ReadNotFound = true
	m, errs := manifest.ParseToManifest(strings.NewReader(maker.ManifestWithName("many apps")))
	testassert.AssertNoErrors(t, errs)
	deps.manifestRepo.ReadManifestManifest = m

	ui := callPush(t, []string{}, deps)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Creating", "app1"},
		{"Creating", "app2"},
	})
	assert.Equal(t, len(deps.appRepo.CreateAppParams), 2)

	firstApp := deps.appRepo.CreateAppParams[0]
	secondApp := deps.appRepo.CreateAppParams[1]
	assert.Equal(t, firstApp.Get("name").(string), "app1")
	assert.Equal(t, secondApp.Get("name").(string), "app2")

	assert.True(t, firstApp.Has("env"))
	assert.True(t, secondApp.Has("env"))

	envVars := firstApp.Get("env").(generic.Map)
	assert.Equal(t, 2, envVars.Count())
	assert.Equal(t, envVars.Get("PATH").(string), "/u/apps/something/bin")
	assert.Equal(t, envVars.Get("SOMETHING").(string), "definitely-something")

	envVars = secondApp.Get("env").(generic.Map)
	assert.Equal(t, 2, envVars.Count())
	assert.Equal(t, envVars.Get("PATH").(string), "/u/apps/something/bin")
	assert.Equal(t, envVars.Get("SOMETHING").(string), "nothing")
}

func TestPushingWithBindingGlobalServices(t *testing.T) {
	deps := getPushDependencies()
	deps.appRepo.ReadNotFound = true

	expectedServiceInstance := cf.ServiceInstance{}
	expectedServiceInstance.Name = "work-queue"
	deps.serviceRepo.FindInstanceByNameServiceInstance = expectedServiceInstance

	m, errs := manifest.ParseToManifest(strings.NewReader(maker.ManifestWithName("global services")))
	testassert.AssertNoErrors(t, errs)
	deps.manifestRepo.ReadManifestManifest = m

	ui := callPush(t, []string{}, deps)

	assert.Equal(t, len(deps.binder.AppsToBind), 1)
	assert.Equal(t, deps.binder.AppsToBind[0].Name, "app-with-redis-backend")
	assert.Equal(t, deps.binder.InstancesToBindTo[0].Name, "work-queue")
	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Creating", "app-with-redis-backend"},
		{"OK"},
		{"Binding service", "work-queue", "app-with-redis-backend", "my-org", "my-space", "my-user"},
		{"OK"},
	})
}

func TestPushingWithBindingLocalServices(t *testing.T) {
	deps := getPushDependencies()
	deps.appRepo.ReadNotFound = true

	expectedServiceInstance := cf.ServiceInstance{}
	expectedServiceInstance.Name = "work-queue"
	deps.serviceRepo.FindInstanceByNameServiceInstance = expectedServiceInstance

	m, errs := manifest.ParseToManifest(strings.NewReader(maker.ManifestWithName("local services")))
	testassert.AssertNoErrors(t, errs)
	deps.manifestRepo.ReadManifestManifest = m

	ui := callPush(t, []string{}, deps)

	assert.Equal(t, len(deps.binder.AppsToBind), 1)
	assert.Equal(t, deps.binder.AppsToBind[0].Name, "app-with-redis-backend")
	assert.Equal(t, deps.binder.InstancesToBindTo[0].Name, "work-queue")
	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Creating", "app-with-redis-backend"},
		{"OK"},
		{"Binding service", "work-queue", "app-with-redis-backend", "my-org", "my-space", "my-user"},
		{"OK"},
	})
}

func TestPushingWithBindingMergedServices(t *testing.T) {
	deps := getPushDependencies()
	deps.appRepo.ReadNotFound = true

	service1 := cf.ServiceInstance{}
	service1.Name = "work-queue"
	service2 := cf.ServiceInstance{}
	service2.Name = "nested-service"
	service3 := cf.ServiceInstance{}
	service3.Name = "app2-service"

	mapOfServices := generic.NewMap()
	mapOfServices.Set("global-service", service1)
	mapOfServices.Set("nested-service", service2)
	mapOfServices.Set("app2-service", service3)
	deps.serviceRepo.FindInstanceByNameMap = mapOfServices

	m, errs := manifest.ParseToManifest(strings.NewReader(maker.ManifestWithName("merged services")))
	testassert.AssertNoErrors(t, errs)
	deps.manifestRepo.ReadManifestManifest = m

	ui := callPush(t, []string{}, deps)

	assert.Equal(t, len(deps.binder.AppsToBind), 4)
	assert.Equal(t, deps.binder.AppsToBind[0].Name, "app-with-redis-backend")
	assert.Equal(t, deps.binder.InstancesToBindTo[0].Name, "work-queue")

	assert.Equal(t, deps.binder.AppsToBind[1].Name, "app-with-redis-backend")
	assert.Equal(t, deps.binder.InstancesToBindTo[1].Name, "nested-service")

	assert.Equal(t, deps.binder.AppsToBind[2].Name, "app2")
	assert.Equal(t, deps.binder.InstancesToBindTo[2].Name, "work-queue")

	assert.Equal(t, deps.binder.AppsToBind[3].Name, "app2")
	assert.Equal(t, deps.binder.InstancesToBindTo[3].Name, "app2-service")

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Creating", "app-with-redis-backend"},
		{"OK"},
		{"Binding service", "global-service", "app-with-redis-backend", "my-org", "my-space", "my-user"},
		{"OK"},
		{"Binding service", "nested-service", "app-with-redis-backend", "my-org", "my-space", "my-user"},
		{"OK"},
		{"Creating", "app2"},
		{"OK"},
		{"Binding service", "global-service", "app2", "my-org", "my-space", "my-user"},
		{"OK"},
		{"Binding service", "app2-service", "app2", "my-org", "my-space", "my-user"},
		{"OK"},
	})
}

func TestPushWithInvalidManifestProperties(t *testing.T) {
	deps := getPushDependencies()
	deps.appRepo.ReadNotFound = true

	m, errs := manifest.ParseToManifest(strings.NewReader(maker.ManifestWithName("invalid")))
	deps.manifestRepo.ReadManifestManifest = m
	deps.manifestRepo.ReadManifestErrors = errs

	ui := callPush(t, []string{}, deps)
	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"FAILED"},
		{"Error", "reading", "manifest"},
		{"Expected", "local services"},
		{"Expected", "local env"},
		{"Expected", "global env"},
		{"Expected", "global services"},
	})
}

func TestPushWithServicesThatAreNotFound(t *testing.T) {
	deps := getPushDependencies()
	deps.routeRepo.FindByHostAndDomainErr = true
	deps.serviceRepo.FindInstanceByNameErr = true

	m, errs := manifest.ParseToManifest(strings.NewReader(maker.ManifestWithName("global services")))
	testassert.AssertNoErrors(t, errs)
	deps.manifestRepo.ReadManifestManifest = m

	ui := callPush(t, []string{}, deps)
	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"FAILED"},
		{"Could not find service", "work-queue", "app-with-redis-backend"},
	})
}

func TestPushingAppWithPath(t *testing.T) {
	deps := getPushDependencies()
	deps.appRepo.ReadNotFound = true

	m, errs := manifest.ParseToManifest(strings.NewReader(maker.ManifestWithName("single app")))
	testassert.AssertNoErrors(t, errs)
	deps.manifestRepo.ReadManifestManifest = m
	deps.manifestRepo.ManifestDir = "/foo/bar/baz"

	callPush(t, []string{
		"-p", "/foo/bar/baz",
		"my-new-app",
	}, deps)

	assert.Equal(t, deps.manifestRepo.UserSpecifiedPath, "/foo/bar/baz")
	assert.Equal(t, deps.appRepo.CreatedAppParams().Get("path").(string), filepath.Join("/foo/bar/baz", "../../fixtures/example-app"))
}

func TestPushingWithRelativeManifestPath(t *testing.T) {
	deps := getPushDependencies()
	deps.appRepo.ReadNotFound = true

	m, errs := manifest.ParseToManifest(strings.NewReader(maker.ManifestWithName("single app")))
	testassert.AssertNoErrors(t, errs)
	deps.manifestRepo.ReadManifestManifest = m
	deps.manifestRepo.ManifestDir = "returned/path/"
	deps.manifestRepo.ManifestFilename = "different-manifest.yml"

	ui := callPush(t, []string{
		"--manifest", "user/supplied/path/different-manifest.yml",
		"-p", "foo/bar/baz",
	}, deps)

	assert.Equal(t, deps.manifestRepo.UserSpecifiedPath, "user/supplied/path/different-manifest.yml")
	assert.Equal(t, deps.manifestRepo.ReadManifestPath, filepath.Clean("returned/path/different-manifest.yml"))
	assert.Equal(t, deps.appRepo.CreatedAppParams().Get("path").(string), filepath.Join("returned/path", "../../fixtures/example-app"))

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"-p is ignored when using a manifest"},
	})
}

func TestPushingWithBadManifestPath(t *testing.T) {
	deps := getPushDependencies()
	deps.appRepo.ReadNotFound = true

	m, errs := manifest.ParseToManifest(strings.NewReader(maker.ManifestWithName("single app")))
	testassert.AssertNoErrors(t, errs)
	deps.manifestRepo.ReadManifestManifest = m
	deps.manifestRepo.ManifestPathErr = errors.New("read manifest error")

	ui := callPush(t, []string{
		"--manifest", "bad/manifest/path",
	}, deps)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"FAILED"},
		{"read manifest error"},
	})
}

func TestPushingWithDefaultManifestNotFound(t *testing.T) {
	deps := getPushDependencies()
	deps.appRepo.ReadNotFound = true

	m, errs := manifest.ParseToManifest(strings.NewReader(maker.ManifestWithName("single app")))
	testassert.AssertNoErrors(t, errs)
	deps.manifestRepo.ReadManifestManifest = m
	deps.manifestRepo.ReadManifestErrors = manifest.ManifestErrors{syscall.ENOENT}

	ui := callPush(t, []string{"--no-route", "app-name"}, deps)

	testassert.SliceDoesNotContain(t, ui.Outputs, testassert.Lines{
		{"FAILED"},
	})
}

func TestPushingWithSpecifiedManifestNotFound(t *testing.T) {
	deps := getPushDependencies()
	deps.appRepo.ReadNotFound = true

	m, errs := manifest.ParseToManifest(strings.NewReader(maker.ManifestWithName("single app")))
	testassert.AssertNoErrors(t, errs)
	deps.manifestRepo.ReadManifestManifest = m
	deps.manifestRepo.ManifestPathErr = syscall.ENOENT

	ui := callPush(t, []string{
		"--manifest", "bad/manifest/path",
	}, deps)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"FAILED"},
	})
}

func TestPushingWithAppPathFromManifestFile(t *testing.T) {
	deps := getPushDependencies()
	deps.appRepo.ReadNotFound = true

	m, errs := manifest.ParseToManifest(strings.NewReader(maker.ManifestWithName("single app")))
	testassert.AssertNoErrors(t, errs)
	deps.manifestRepo.ReadManifestManifest = m
	deps.manifestRepo.ManifestDir = "some/relative/path/"
	deps.manifestRepo.ManifestFilename = "different-manifest.yml"
	ui := callPush(t, []string{
		"--manifest", "some/relative/path/different-manifest.yml",
	}, deps)

	expectedManifestPath := filepath.Clean("some/relative/path/different-manifest.yml")
	assert.Equal(t, deps.manifestRepo.UserSpecifiedPath, "some/relative/path/different-manifest.yml")
	assert.Equal(t, deps.manifestRepo.ReadManifestPath, expectedManifestPath)
	assert.Equal(t, deps.appRepo.CreatedAppParams().Get("path").(string), filepath.Join("some/relative/path/", "../../fixtures/example-app"))

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Using manifest file", expectedManifestPath},
	})
}

func TestPushingWithManifestInAppDirectory(t *testing.T) {
	deps := getPushDependencies()
	deps.appRepo.ReadNotFound = true

	m, errs := manifest.ParseToManifest(strings.NewReader(maker.ManifestWithName("single app")))
	testassert.AssertNoErrors(t, errs)
	deps.manifestRepo.ReadManifestManifest = m

	ui := callPush(t, []string{"-p", "some/relative/path"}, deps)

	expectedManifestPath := filepath.Clean("some/relative/path/manifest.yml")
	assert.Equal(t, deps.manifestRepo.UserSpecifiedPath, "some/relative/path")
	assert.Equal(t, deps.manifestRepo.ReadManifestPath, expectedManifestPath)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Using manifest file", expectedManifestPath},
	})
}

func TestPushingAppWhenManifestIncludesRelativePathForApp(t *testing.T) {
	deps := getPushDependencies()
	deps.appRepo.ReadNotFound = true

	m, errs := manifest.ParseToManifest(strings.NewReader(maker.ManifestWithName("single app")))
	testassert.AssertNoErrors(t, errs)
	deps.manifestRepo.ReadManifestManifest = m
	deps.manifestRepo.ManifestDir = "some/relative/path"

	_ = callPush(t, []string{"--manifest", "some/relative/path"}, deps)

	assert.Equal(t, deps.manifestRepo.UserSpecifiedPath, "some/relative/path")
	assert.Equal(t, deps.appRepo.CreatedAppParams().Get("path").(string), filepath.Join("some/relative/path/", "../../fixtures/example-app"))
}

func TestPushingWithNoManifestFlag(t *testing.T) {
	deps := getPushDependencies()
	deps.appRepo.ReadNotFound = true

	m, errs := manifest.ParseToManifest(strings.NewReader(maker.ManifestWithName("nulls")))
	deps.manifestRepo.ReadManifestManifest = m
	deps.manifestRepo.ReadManifestErrors = errs

	ui := callPush(t, []string{"--no-route", "--no-manifest", "app-name"}, deps)

	testassert.SliceDoesNotContain(t, ui.Outputs, testassert.Lines{
		{"FAILED"},
		{"hacker-manifesto"},
	})
	assert.Equal(t, deps.appRepo.CreatedAppParams().Get("name").(string), "app-name")
}

func TestPushingWithNoManifestFlagAndMissingAppName(t *testing.T) {
	deps := getPushDependencies()
	deps.appRepo.ReadNotFound = true

	m, errs := manifest.ParseToManifest(strings.NewReader(maker.ManifestWithName("nulls")))
	deps.manifestRepo.ReadManifestManifest = m
	deps.manifestRepo.ReadManifestErrors = errs

	ui := callPush(t, []string{"--no-route", "--no-manifest"}, deps)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"FAILED"},
	})

	testassert.SliceDoesNotContain(t, ui.Outputs, testassert.Lines{
		{"hacker-manifesto"},
	})
}

func TestPushingAppWithNoRoute(t *testing.T) {
	deps := getPushDependencies()
	domain := cf.Domain{}
	domain.Name = "bar.cf-app.com"
	domain.Guid = "bar-domain-guid"
	stack := cf.Stack{}
	stack.Name = "customLinux"
	stack.Guid = "custom-linux-guid"

	deps.domainRepo.FindByNameDomain = domain
	deps.routeRepo.FindByHostErr = true
	deps.stackRepo.FindByNameStack = stack
	deps.appRepo.ReadNotFound = true

	callPush(t, []string{
		"--no-route",
		"my-new-app",
	}, deps)

	assert.Equal(t, deps.appRepo.CreatedAppParams().Get("name").(string), "my-new-app")
	assert.Equal(t, deps.routeRepo.CreatedHost, "")
	assert.Equal(t, deps.routeRepo.CreatedDomainGuid, "")
}

func TestPushingAppWithNoHostname(t *testing.T) {
	deps := getPushDependencies()
	domain := cf.Domain{}
	domain.Name = "bar.cf-app.com"
	domain.Guid = "bar-domain-guid"
	domain.Shared = true

	stack := cf.Stack{}
	stack.Name = "customLinux"
	stack.Guid = "custom-linux-guid"

	deps.domainRepo.ListDomainsForOrgDomains = []cf.Domain{domain}
	deps.routeRepo.FindByHostAndDomainErr = true
	deps.stackRepo.FindByNameStack = stack
	deps.appRepo.ReadNotFound = true

	callPush(t, []string{
		"--no-hostname",
		"my-new-app",
	}, deps)

	assert.Equal(t, deps.appRepo.CreatedAppParams().Get("name").(string), "my-new-app")
	assert.Equal(t, deps.routeRepo.CreatedHost, "")
	assert.Equal(t, deps.routeRepo.CreatedDomainGuid, "bar-domain-guid")
}

func TestPushingAppWithMemoryInMegaBytes(t *testing.T) {
	deps := getPushDependencies()
	deps.appRepo.ReadNotFound = true

	callPush(t, []string{
		"-m", "256M",
		"my-new-app",
	}, deps)

	assert.Equal(t, deps.appRepo.CreatedAppParams().Get("memory").(uint64), uint64(256))
}

func TestPushingAppWithMemoryWithoutUnit(t *testing.T) {
	deps := getPushDependencies()
	deps.appRepo.ReadNotFound = true

	callPush(t, []string{
		"-m", "512",
		"my-new-app",
	}, deps)

	assert.Equal(t, deps.appRepo.CreatedAppParams().Get("memory").(uint64), uint64(512))
}

func TestPushingAppWithInvalidMemory(t *testing.T) {
	deps := getPushDependencies()
	deps.appRepo.ReadNotFound = true

	ui := callPush(t, []string{
		"-m", "abcM",
		"my-new-app",
	}, deps)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"FAILED"},
		{"invalid", "memory"},
	})
}

func TestPushingAppWhenItAlreadyExistsAndNothingIsSpecified(t *testing.T) {
	deps := getPushDependencies()
	existingApp := maker.NewApp(maker.Overrides{"name": "existing-app"})
	deps.appRepo.ReadApp = existingApp
	deps.appRepo.UpdateAppResult = existingApp

	_ = callPush(t, []string{"existing-app"}, deps)

	assert.Equal(t, deps.stopper.AppToStop.Guid, existingApp.Guid)
	assert.Equal(t, deps.appBitsRepo.UploadedAppGuid, existingApp.Guid)
}

func TestPushingAppWhenItIsStopped(t *testing.T) {
	deps := getPushDependencies()
	stoppedApp := maker.NewApp(maker.Overrides{"state": "stopped", "name": "stopped-app"})

	deps.appRepo.ReadApp = stoppedApp
	deps.appRepo.UpdateAppResult = stoppedApp

	_ = callPush(t, []string{"stopped-app"}, deps)

	assert.Equal(t, deps.stopper.AppToStop.Guid, "")
}

func TestPushingAppWhenItAlreadyExistsAndChangingOptions(t *testing.T) {
	deps := getPushDependencies()

	existingRoute := cf.RouteSummary{}
	existingRoute.Host = "existing-app"

	existingApp := cf.Application{}
	existingApp.Name = "existing-app"
	existingApp.Guid = "existing-app-guid"
	existingApp.Routes = []cf.RouteSummary{existingRoute}

	deps.appRepo.ReadApp = existingApp

	stack := cf.Stack{}
	stack.Name = "differentStack"
	stack.Guid = "differentStack-guid"
	deps.stackRepo.FindByNameStack = stack

	args := []string{
		"-c", "different start command",
		"-i", "10",
		"-m", "1G",
		"-b", "https://github.com/heroku/heroku-buildpack-different.git",
		"-s", "differentStack",
		"existing-app",
	}
	_ = callPush(t, args, deps)

	assert.Equal(t, deps.appRepo.UpdateParams.Get("command"), "different start command")
	assert.Equal(t, deps.appRepo.UpdateParams.Get("instances"), 10)
	assert.Equal(t, deps.appRepo.UpdateParams.Get("memory"), uint64(1024))
	assert.Equal(t, deps.appRepo.UpdateParams.Get("buildpack"), "https://github.com/heroku/heroku-buildpack-different.git")
	assert.Equal(t, deps.appRepo.UpdateParams.Get("stack_guid"), "differentStack-guid")
}

func TestPushingAppWhenItAlreadyExistsAndDomainIsSpecifiedIsAlreadyBound(t *testing.T) {
	deps := getPushDependencies()

	domain := cf.DomainFields{}
	domain.Name = "example.com"
	domain.Guid = "domain-guid"

	existingRoute := cf.RouteSummary{}
	existingRoute.Host = "existing-app"
	existingRoute.Domain = domain

	existingApp := cf.Application{}
	existingApp.Name = "existing-app"
	existingApp.Guid = "existing-app-guid"
	existingApp.Routes = []cf.RouteSummary{existingRoute}

	foundRoute := cf.Route{}
	foundRoute.RouteFields = existingRoute.RouteFields
	foundRoute.Domain = existingRoute.Domain

	deps.appRepo.ReadApp = existingApp
	deps.appRepo.UpdateAppResult = existingApp
	deps.routeRepo.FindByHostAndDomainRoute = foundRoute

	ui := callPush(t, []string{"-d", "example.com", "existing-app"}, deps)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Using route", "existing-app", "example.com"},
	})
	assert.Equal(t, deps.appBitsRepo.UploadedAppGuid, "existing-app-guid")
}

func TestPushingAppWhenItAlreadyExistsAndDomainSpecifiedIsNotBound(t *testing.T) {
	deps := getPushDependencies()

	domain := cf.DomainFields{}
	domain.Name = "example.com"

	existingRoute := cf.RouteSummary{}
	existingRoute.Host = "existing-app"
	existingRoute.Domain = domain

	existingApp := cf.Application{}
	existingApp.Name = "existing-app"
	existingApp.Guid = "existing-app-guid"
	existingApp.Routes = []cf.RouteSummary{existingRoute}

	foundDomain := cf.Domain{}
	foundDomain.Guid = "domain-guid"
	foundDomain.Name = "newdomain.com"

	deps.appRepo.ReadApp = existingApp
	deps.appRepo.UpdateAppResult = existingApp
	deps.routeRepo.FindByHostAndDomainNotFound = true
	deps.domainRepo.FindByNameDomain = foundDomain

	ui := callPush(t, []string{"-d", "newdomain.com", "existing-app"}, deps)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Creating route", "existing-app.newdomain.com"},
		{"OK"},
		{"Binding", "existing-app.newdomain.com"},
	})

	assert.Equal(t, deps.appBitsRepo.UploadedAppGuid, "existing-app-guid")
	assert.Equal(t, deps.domainRepo.FindByNameInCurrentSpaceName, "newdomain.com")
	assert.Equal(t, deps.routeRepo.FindByHostAndDomainDomain, "newdomain.com")
	assert.Equal(t, deps.routeRepo.FindByHostAndDomainHost, "existing-app")
	assert.Equal(t, deps.routeRepo.CreatedHost, "existing-app")
	assert.Equal(t, deps.routeRepo.CreatedDomainGuid, "domain-guid")
}

func TestPushingAppWithNoFlagsWhenAppIsAlreadyBoundToDomain(t *testing.T) {
	deps := getPushDependencies()

	domain := cf.DomainFields{}
	domain.Name = "example.com"

	existingRoute := cf.RouteSummary{}
	existingRoute.Host = "foo"
	existingRoute.Domain = domain

	existingApp := cf.Application{}
	existingApp.Name = "existing-app"
	existingApp.Guid = "existing-app-guid"
	existingApp.Routes = []cf.RouteSummary{existingRoute}

	deps.appRepo.ReadApp = existingApp
	deps.appRepo.UpdateAppResult = existingApp

	_ = callPush(t, []string{"existing-app"}, deps)

	assert.Equal(t, deps.appBitsRepo.UploadedAppGuid, "existing-app-guid")
	assert.Equal(t, deps.domainRepo.FindByNameInCurrentSpaceName, "")
	assert.Equal(t, deps.routeRepo.FindByHostAndDomainDomain, "")
	assert.Equal(t, deps.routeRepo.FindByHostAndDomainHost, "")
	assert.Equal(t, deps.routeRepo.CreatedHost, "")
	assert.Equal(t, deps.routeRepo.CreatedDomainGuid, "")
}

func TestPushingAppWhenItAlreadyExistsAndHostIsSpecified(t *testing.T) {
	deps := getPushDependencies()

	domain := cf.Domain{}
	domain.Name = "example.com"
	domain.Guid = "domain-guid"
	domain.Shared = true

	existingRoute := cf.RouteSummary{}
	existingRoute.Host = "existing-app"
	existingRoute.Domain = domain.DomainFields

	existingApp := cf.Application{}
	existingApp.Name = "existing-app"
	existingApp.Guid = "existing-app-guid"
	existingApp.Routes = []cf.RouteSummary{existingRoute}

	deps.appRepo.ReadApp = existingApp
	deps.appRepo.UpdateAppResult = existingApp
	deps.routeRepo.FindByHostAndDomainNotFound = true
	deps.domainRepo.ListDomainsForOrgDomains = []cf.Domain{domain}

	ui := callPush(t, []string{"-n", "new-host", "existing-app"}, deps)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Creating route", "new-host.example.com"},
		{"OK"},
		{"Binding", "new-host.example.com"},
	})

	assert.Equal(t, deps.routeRepo.FindByHostAndDomainDomain, "example.com")
	assert.Equal(t, deps.routeRepo.FindByHostAndDomainHost, "new-host")
	assert.Equal(t, deps.routeRepo.CreatedHost, "new-host")
	assert.Equal(t, deps.routeRepo.CreatedDomainGuid, "domain-guid")
}

func TestPushingAppWhenItAlreadyExistsAndNoRouteFlagIsPresent(t *testing.T) {
	deps := getPushDependencies()
	existingApp := cf.Application{}
	existingApp.Name = "existing-app"
	existingApp.Guid = "existing-app-guid"

	deps.appRepo.ReadApp = existingApp
	deps.appRepo.UpdateAppResult = existingApp

	ui := callPush(t, []string{"--no-route", "existing-app"}, deps)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Uploading", "existing-app"},
		{"OK"},
	})

	assert.Equal(t, deps.appBitsRepo.UploadedAppGuid, "existing-app-guid")
	assert.Equal(t, deps.domainRepo.FindByNameInCurrentSpaceName, "")
	assert.Equal(t, deps.routeRepo.FindByHostAndDomainDomain, "")
	assert.Equal(t, deps.routeRepo.FindByHostAndDomainHost, "")
	assert.Equal(t, deps.routeRepo.CreatedHost, "")
	assert.Equal(t, deps.routeRepo.CreatedDomainGuid, "")
}

func TestPushingAppWhenItAlreadyExistsAndNoHostFlagIsPresent(t *testing.T) {
	deps := getPushDependencies()

	domain := cf.Domain{}
	domain.Name = "example.com"
	domain.Guid = "domain-guid"
	domain.Shared = true

	existingRoute := cf.RouteSummary{}
	existingRoute.Host = "existing-app"
	existingRoute.Domain = domain.DomainFields

	existingApp := cf.Application{}
	existingApp.Name = "existing-app"
	existingApp.Guid = "existing-app-guid"
	existingApp.Routes = []cf.RouteSummary{existingRoute}

	deps.appRepo.ReadApp = existingApp
	deps.appRepo.UpdateAppResult = existingApp
	deps.routeRepo.FindByHostAndDomainNotFound = true
	deps.domainRepo.ListDomainsForOrgDomains = []cf.Domain{domain}

	ui := callPush(t, []string{"--no-hostname", "existing-app"}, deps)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Creating route", "example.com"},
		{"OK"},
		{"Binding", "example.com"},
	})
	testassert.SliceDoesNotContain(t, ui.Outputs, testassert.Lines{
		{"existing-app.example.com"},
	})

	assert.Equal(t, deps.routeRepo.FindByHostAndDomainDomain, "example.com")
	assert.Equal(t, deps.routeRepo.FindByHostAndDomainHost, "")
	assert.Equal(t, deps.routeRepo.CreatedHost, "")
	assert.Equal(t, deps.routeRepo.CreatedDomainGuid, "domain-guid")
}

func TestPushingAppWhenItAlreadyExistsWithoutARouteAndARouteIsNotProvided(t *testing.T) {
	deps := getPushDependencies()
	existingApp := cf.Application{}
	existingApp.Name = "existing-app"
	existingApp.Guid = "existing-app-guid"

	deps.appRepo.ReadApp = existingApp
	deps.appRepo.UpdateAppResult = existingApp

	ui := callPush(t, []string{"existing-app"}, deps)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		testassert.Line{"skipping route creation"},
		testassert.Line{"Uploading"},
		testassert.Line{"OK"},
	})

	assert.Equal(t, deps.routeRepo.FindByHostAndDomainDomain, "")
	assert.Equal(t, deps.routeRepo.FindByHostAndDomainHost, "")
	assert.Equal(t, deps.routeRepo.CreatedHost, "")
	assert.Equal(t, deps.routeRepo.CreatedDomainGuid, "")
}

func TestPushingAppWithInvalidPath(t *testing.T) {
	deps := getPushDependencies()
	deps.appBitsRepo.UploadAppErr = true

	ui := callPush(t, []string{"app"}, deps)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Uploading"},
		{"FAILED"},
	})
}

type pushDependencies struct {
	manifestRepo *testmanifest.FakeManifestRepository
	starter      *testcmd.FakeAppStarter
	stopper      *testcmd.FakeAppStopper
	binder       *testcmd.FakeAppBinder
	appRepo      *testapi.FakeApplicationRepository
	domainRepo   *testapi.FakeDomainRepository
	routeRepo    *testapi.FakeRouteRepository
	stackRepo    *testapi.FakeStackRepository
	appBitsRepo  *testapi.FakeApplicationBitsRepository
	serviceRepo  *testapi.FakeServiceRepo
}

func getPushDependencies() (deps pushDependencies) {
	deps.manifestRepo = &testmanifest.FakeManifestRepository{}
	deps.starter = &testcmd.FakeAppStarter{}
	deps.stopper = &testcmd.FakeAppStopper{}
	deps.binder = &testcmd.FakeAppBinder{}
	deps.appRepo = &testapi.FakeApplicationRepository{}
	deps.domainRepo = &testapi.FakeDomainRepository{}
	deps.routeRepo = &testapi.FakeRouteRepository{}
	deps.stackRepo = &testapi.FakeStackRepository{}
	deps.appBitsRepo = &testapi.FakeApplicationBitsRepository{}
	deps.serviceRepo = &testapi.FakeServiceRepo{}

	return
}

func callPush(t *testing.T, args []string, deps pushDependencies) (ui *testterm.FakeUI) {

	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("push", args)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)
	org := cf.OrganizationFields{}
	org.Name = "my-org"
	org.Guid = "my-org-guid"
	space := cf.SpaceFields{}
	space.Name = "my-space"
	space.Guid = "my-space-guid"

	config := &configuration.Configuration{
		SpaceFields:        space,
		OrganizationFields: org,
		AccessToken:        token,
	}

	manifestRepo := deps.manifestRepo
	starter := deps.starter
	stopper := deps.stopper
	binder := deps.binder
	appRepo := deps.appRepo
	domainRepo := deps.domainRepo
	routeRepo := deps.routeRepo
	stackRepo := deps.stackRepo
	appBitsRepo := deps.appBitsRepo
	serviceRepo := deps.serviceRepo

	cmd := NewPush(ui, config, manifestRepo, starter, stopper, binder, appRepo, domainRepo, routeRepo, stackRepo, serviceRepo, appBitsRepo)
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}
	testcmd.RunCommand(cmd, ctxt, reqFactory)

	return
}
