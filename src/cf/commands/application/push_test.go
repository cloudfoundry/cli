package application_test

import (
	"cf"
	"cf/api"
	. "cf/commands/application"
	"cf/configuration"
	"cf/manifest"
	"generic"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"strings"
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

const singleAppManifest = `
---
env:
  PATH: /u/apps/my-app/bin
  FOO: bar
applications:
- name: manifest-app-name
  memory: 128M
  instances: 1
  host: manifest-host
  domain: manifest-example.com
  stack: custom-stack
  buildpack: some-buildpack
  command: JAVA_HOME=$PWD/.openjdk JAVA_OPTS="-Xss995K" ./bin/start.sh run
  path: ../../fixtures/example-app
  env:
    FOO: baz
`

const appManifestWithNulls = `
applications:
- name: hacker-manifesto
  command: null
  space_guid: null
  buildpack: null
  disk_quota: null
  instances: null
  memory: null
  env: null
`

const manifestWithManyApps = `
---
env:
  PATH: /u/apps/something/bin
  SOMETHING: nothing
applications:
- name: app1
  env:
    SOMETHING: definitely-something
- name: app2
`

func TestPushingRequirements(t *testing.T) {
	manifestRepo, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo := getPushDependencies()
	ui := new(testterm.FakeUI)
	config := &configuration.Configuration{}

	cmd := NewPush(ui, config, manifestRepo, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo)
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
	manifestRepo, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo := getPushDependencies()

	domain := cf.Domain{}
	domain.Guid = "not-the-right-guid"
	domain.Name = "not shared domain"
	domain.OwningOrganizationGuid = "my-org-guid"

	sharedDomain := cf.Domain{}
	sharedDomain.Name = "foo.cf-app.com"
	sharedDomain.Shared = true
	sharedDomain.Guid = "foo-domain-guid"

	domainRepo.ListDomainsForOrgDomains = []cf.Domain{domain, sharedDomain}
	routeRepo.FindByHostAndDomainErr = true

	appRepo.ReadNotFound = true

	ui := callPush(t, []string{"my-new-app"}, manifestRepo, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo)

	assert.Equal(t, appRepo.CreatedAppParams().Get("name").(string), "my-new-app")
	assert.Equal(t, appRepo.CreatedAppParams().Get("space_guid").(string), "my-space-guid")

	assert.Equal(t, routeRepo.FindByHostAndDomainHost, "my-new-app")
	assert.Equal(t, routeRepo.CreatedHost, "my-new-app")
	assert.Equal(t, routeRepo.CreatedDomainGuid, "foo-domain-guid")
	assert.Equal(t, routeRepo.BoundAppGuid, "my-new-app-guid")
	assert.Equal(t, routeRepo.BoundRouteGuid, "my-new-app-route-guid")

	assert.Equal(t, appBitsRepo.UploadedAppGuid, "my-new-app-guid")

	expectedAppDir, err := os.Getwd()
	assert.NoError(t, err)
	assert.Equal(t, appBitsRepo.UploadedDir, expectedAppDir)

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
}

func TestPushingAppWhenItDoesNotExistButRouteExists(t *testing.T) {
	manifestRepo, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo := getPushDependencies()

	domain := cf.Domain{}
	domain.Name = "foo.cf-app.com"
	domain.Guid = "foo-domain-guid"
	domain.Shared = true

	route := cf.Route{}
	route.Guid = "my-route-guid"
	route.Host = "my-new-app"
	route.Domain = domain.DomainFields

	domainRepo.ListDomainsForOrgDomains = []cf.Domain{domain}

	routeRepo.FindByHostAndDomainRoute = route
	appRepo.ReadNotFound = true

	ui := callPush(t, []string{"my-new-app"}, manifestRepo, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo)

	assert.Empty(t, routeRepo.CreatedHost)
	assert.Empty(t, routeRepo.CreatedDomainGuid)
	assert.Equal(t, routeRepo.FindByHostAndDomainHost, "my-new-app")
	assert.Equal(t, routeRepo.BoundAppGuid, "my-new-app-guid")
	assert.Equal(t, routeRepo.BoundRouteGuid, "my-route-guid")

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Using", "my-new-app.foo.cf-app.com"},
		{"Binding", "my-new-app.foo.cf-app.com"},
		{"OK"},
	})
}

func TestPushingAppWithCustomFlags(t *testing.T) {
	manifestRepo, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo := getPushDependencies()
	domain := cf.Domain{}
	domain.Name = "bar.cf-app.com"
	domain.Guid = "bar-domain-guid"
	stack := cf.Stack{}
	stack.Name = "customLinux"
	stack.Guid = "custom-linux-guid"

	domainRepo.FindByNameDomain = domain
	routeRepo.FindByHostAndDomainErr = true
	stackRepo.FindByNameStack = stack
	appRepo.ReadNotFound = true

	ui := callPush(t, []string{
		"-c", "unicorn -c config/unicorn.rb -D",
		"-d", "bar.cf-app.com",
		"-n", "my-hostname",
		"-i", "3",
		"-m", "2G",
		"-b", "https://github.com/heroku/heroku-buildpack-play.git",
		"-p", "/Users/pivotal/workspace/my-new-app",
		"-s", "customLinux",
		"-t", "1",
		"--no-start",
		"my-new-app",
	}, manifestRepo, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo)

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

	assert.Equal(t, stackRepo.FindByNameName, "customLinux")

	assert.Equal(t, appRepo.CreatedAppParams().Get("name").(string), "my-new-app")
	assert.Equal(t, appRepo.CreatedAppParams().Get("command").(string), "unicorn -c config/unicorn.rb -D")
	assert.Equal(t, appRepo.CreatedAppParams().Get("instances").(int), 3)
	assert.Equal(t, appRepo.CreatedAppParams().Get("memory").(uint64), uint64(2048))
	assert.Equal(t, appRepo.CreatedAppParams().Get("stack_guid"), "custom-linux-guid")
	assert.Equal(t, appRepo.CreatedAppParams().Get("health_check_timeout").(int), 1)
	assert.Equal(t, appRepo.CreatedAppParams().Get("buildpack"), "https://github.com/heroku/heroku-buildpack-play.git")

	assert.Equal(t, domainRepo.FindByNameInCurrentSpaceName, "bar.cf-app.com")

	assert.Equal(t, routeRepo.CreatedHost, "my-hostname")
	assert.Equal(t, routeRepo.CreatedDomainGuid, "bar-domain-guid")
	assert.Equal(t, routeRepo.BoundAppGuid, "my-new-app-guid")
	assert.Equal(t, routeRepo.BoundRouteGuid, "my-hostname-route-guid")

	assert.Equal(t, appBitsRepo.UploadedAppGuid, "my-new-app-guid")
	assert.Equal(t, appBitsRepo.UploadedDir, "/Users/pivotal/workspace/my-new-app")

	assert.Equal(t, starter.AppToStart.Name, "")
}

func TestPushingAppWithInvalidTimeout(t *testing.T) {
	manifestRepo, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo := getPushDependencies()
	domain := cf.Domain{}
	domain.Name = "bar.cf-app.com"
	domain.Guid = "bar-domain-guid"
	domainRepo.FindByNameDomain = domain
	appRepo.ReadNotFound = true

	ui := callPush(t, []string{
		"-t", "FooeyTimeout",
		"my-new-app",
	}, manifestRepo, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"FAILED"},
		{"invalid", "timeout"},
	})
}

func TestPushingAppToResetStartCommand(t *testing.T) {
	manifestRepo, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo := getPushDependencies()

	existingApp := cf.Application{}
	existingApp.Name = "existing-app"
	existingApp.Guid = "existing-app-guid"
	existingApp.Command = "unicorn -c config/unicorn.rb -D"

	appRepo.ReadApp = existingApp

	args := []string{
		"-c", "null",
		"existing-app",
	}
	_ = callPush(t, args, manifestRepo, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo)

	assert.Equal(t, appRepo.UpdateParams.Get("command"), "null")
}

func TestPushingAppWithSingleAppManifest(t *testing.T) {
	manifestRepo, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo := getPushDependencies()
	domain := cf.Domain{}
	domain.Name = "manifest-example.com"
	domain.Guid = "bar-domain-guid"

	domainRepo.FindByNameDomain = domain
	routeRepo.FindByHostAndDomainErr = true
	appRepo.ReadNotFound = true

	m, err := manifest.Parse(strings.NewReader(singleAppManifest))
	assert.NoError(t, err)
	manifestRepo.ReadManifestManifest = m

	ui := callPush(t, []string{}, manifestRepo, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Creating route", "manifest-host.manifest-example.com"},
		{"OK"},
		{"Binding", "manifest-host.manifest-example.com"},
		{"manifest-app-name"},
	})

	assert.Equal(t, appRepo.CreatedAppParams().Get("name").(string), "manifest-app-name")
	assert.Equal(t, appRepo.CreatedAppParams().Get("memory").(uint64), uint64(128))
	assert.Equal(t, appRepo.CreatedAppParams().Get("instances").(int), 1)
	assert.Equal(t, appRepo.CreatedAppParams().Get("stack").(string), "custom-stack")
	assert.Equal(t, appRepo.CreatedAppParams().Get("buildpack").(string), "some-buildpack")
	assert.Equal(t, appRepo.CreatedAppParams().Get("command").(string), "JAVA_HOME=$PWD/.openjdk JAVA_OPTS=\"-Xss995K\" ./bin/start.sh run")

	dir, err := os.Getwd()
	assert.NoError(t, err)
	assert.Equal(t, appRepo.CreatedAppParams().Get("path").(string), filepath.Join(dir, "../../fixtures/example-app"))

	assert.True(t, appRepo.CreatedAppParams().Has("env"))
	envVars := appRepo.CreatedAppParams().Get("env").(generic.Map)

	assert.Equal(t, 2, envVars.Count())
	assert.True(t, envVars.Has("PATH"))
	assert.True(t, envVars.Has("FOO"))

	assert.Equal(t, envVars.Get("PATH").(string), "/u/apps/my-app/bin")
	assert.Equal(t, envVars.Get("FOO").(string), "baz")
}

func TestPushingAppManifestWithNulls(t *testing.T) {
	manifestRepo, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo := getPushDependencies()
	domain := cf.Domain{}
	domain.Name = "bar.cf-app.com"
	domain.Guid = "bar-domain-guid"
	stack := cf.Stack{}
	stack.Name = "customLinux"
	stack.Guid = "custom-linux-guid"

	domainRepo.FindByNameDomain = domain
	routeRepo.FindByHostAndDomainErr = true
	stackRepo.FindByNameStack = stack
	appRepo.ReadNotFound = true

	m, err := manifest.Parse(strings.NewReader(appManifestWithNulls))
	assert.NoError(t, err)
	manifestRepo.ReadManifestManifest = m

	ui := callPush(t, []string{}, manifestRepo, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"hacker-manifesto"},
	})

	assert.Equal(t, appRepo.CreatedAppParams().Get("name").(string), "hacker-manifesto")

	shouldNotHaveKeys := []string{"command", "memory", "instances", "domain", "host", "env", "stack_guid", "buildpack"}
	for _, key := range shouldNotHaveKeys {
		assert.False(t, appRepo.CreatedAppParams().Has(key))
	}
}

func TestPushingManyAppsFromManifest(t *testing.T) {
	manifestRepo, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo := getPushDependencies()
	domain := cf.Domain{}
	domain.Name = "bar.cf-app.com"
	domain.Guid = "bar-domain-guid"
	stack := cf.Stack{}
	stack.Name = "customLinux"
	stack.Guid = "custom-linux-guid"

	domainRepo.FindByNameDomain = domain
	routeRepo.FindByHostAndDomainErr = true
	stackRepo.FindByNameStack = stack
	appRepo.ReadNotFound = true

	m, err := manifest.Parse(strings.NewReader(manifestWithManyApps))
	assert.NoError(t, err)
	manifestRepo.ReadManifestManifest = m

	ui := callPush(t, []string{}, manifestRepo, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Creating", "app1"},
		{"Creating", "app2"},
	})
	assert.Equal(t, len(appRepo.CreateAppParams), 2)

	firstApp := appRepo.CreateAppParams[0]
	secondApp := appRepo.CreateAppParams[1]
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

func TestPushingAppWithPath(t *testing.T) {
	manifestRepo, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo := getPushDependencies()
	domain := cf.Domain{}
	domain.Name = "bar.cf-app.com"
	domain.Guid = "bar-domain-guid"
	domainRepo.FindByNameDomain = domain
	appRepo.ReadNotFound = true

	m, err := manifest.Parse(strings.NewReader(singleAppManifest))
	assert.NoError(t, err)
	manifestRepo.ReadManifestManifest = m

	callPush(t, []string{
		"-p", "/foo/bar/baz",
		"my-new-app",
	}, manifestRepo, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo)

	assert.Equal(t, appRepo.CreatedAppParams().Get("path").(string), filepath.Join("/foo/bar/baz", "../../fixtures/example-app"))
}

func TestPushingAppWithNoRoute(t *testing.T) {
	manifestRepo, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo := getPushDependencies()
	domain := cf.Domain{}
	domain.Name = "bar.cf-app.com"
	domain.Guid = "bar-domain-guid"
	stack := cf.Stack{}
	stack.Name = "customLinux"
	stack.Guid = "custom-linux-guid"

	domainRepo.FindByNameDomain = domain
	routeRepo.FindByHostErr = true
	stackRepo.FindByNameStack = stack
	appRepo.ReadNotFound = true

	callPush(t, []string{
		"--no-route",
		"my-new-app",
	}, manifestRepo, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo)

	assert.Equal(t, appRepo.CreatedAppParams().Get("name").(string), "my-new-app")
	assert.Equal(t, routeRepo.CreatedHost, "")
	assert.Equal(t, routeRepo.CreatedDomainGuid, "")
}

func TestPushingAppWithNoHostname(t *testing.T) {
	manifestRepo, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo := getPushDependencies()
	domain := cf.Domain{}
	domain.Name = "bar.cf-app.com"
	domain.Guid = "bar-domain-guid"
	domain.Shared = true

	stack := cf.Stack{}
	stack.Name = "customLinux"
	stack.Guid = "custom-linux-guid"

	domainRepo.ListDomainsForOrgDomains = []cf.Domain{domain}
	routeRepo.FindByHostAndDomainErr = true
	stackRepo.FindByNameStack = stack
	appRepo.ReadNotFound = true

	callPush(t, []string{
		"--no-hostname",
		"my-new-app",
	}, manifestRepo, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo)

	assert.Equal(t, appRepo.CreatedAppParams().Get("name").(string), "my-new-app")
	assert.Equal(t, routeRepo.CreatedHost, "")
	assert.Equal(t, routeRepo.CreatedDomainGuid, "bar-domain-guid")
}

func TestPushingAppWithMemoryInMegaBytes(t *testing.T) {
	manifestRepo, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo := getPushDependencies()
	domain := cf.Domain{}
	domain.Name = "bar.cf-app.com"
	domain.Guid = "bar-domain-guid"
	domainRepo.FindByNameDomain = domain
	appRepo.ReadNotFound = true

	callPush(t, []string{
		"-m", "256M",
		"my-new-app",
	}, manifestRepo, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo)

	assert.Equal(t, appRepo.CreatedAppParams().Get("memory").(uint64), uint64(256))
}

func TestPushingAppWithMemoryWithoutUnit(t *testing.T) {
	manifestRepo, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo := getPushDependencies()
	domain := cf.Domain{}
	domain.Name = "bar.cf-app.com"
	domain.Guid = "bar-domain-guid"
	domainRepo.FindByNameDomain = domain
	appRepo.ReadNotFound = true

	callPush(t, []string{
		"-m", "512",
		"my-new-app",
	}, manifestRepo, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo)

	assert.Equal(t, appRepo.CreatedAppParams().Get("memory").(uint64), uint64(512))
}

func TestPushingAppWithInvalidMemory(t *testing.T) {
	manifestRepo, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo := getPushDependencies()
	domain := cf.Domain{}
	domain.Name = "bar.cf-app.com"
	domain.Guid = "bar-domain-guid"
	domainRepo.FindByNameDomain = domain
	appRepo.ReadNotFound = true

	ui := callPush(t, []string{
		"-m", "abcM",
		"my-new-app",
	}, manifestRepo, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"FAILED"},
		{"invalid", "memory"},
	})
}

func TestPushingAppWhenItAlreadyExistsAndNothingIsSpecified(t *testing.T) {
	manifestRepo, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo := getPushDependencies()

	existingApp := maker.NewApp(maker.Overrides{"name": "existing-app"})
	appRepo.ReadApp = existingApp
	appRepo.UpdateAppResult = existingApp

	_ = callPush(t, []string{"existing-app"}, manifestRepo, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo)

	assert.Equal(t, stopper.AppToStop.Guid, existingApp.Guid)
	assert.Equal(t, appBitsRepo.UploadedAppGuid, existingApp.Guid)
}

func TestPushingAppWhenItIsStopped(t *testing.T) {
	manifestRepo, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo := getPushDependencies()
	stoppedApp := maker.NewApp(maker.Overrides{"state": "stopped", "name": "stopped-app"})

	appRepo.ReadApp = stoppedApp
	appRepo.UpdateAppResult = stoppedApp

	_ = callPush(t, []string{"stopped-app"}, manifestRepo, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo)

	assert.Equal(t, stopper.AppToStop.Guid, "")
}

func TestPushingAppWhenItAlreadyExistsAndChangingOptions(t *testing.T) {
	manifestRepo, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo := getPushDependencies()

	existingRoute := cf.RouteSummary{}
	existingRoute.Host = "existing-app"

	existingApp := cf.Application{}
	existingApp.Name = "existing-app"
	existingApp.Guid = "existing-app-guid"
	existingApp.Routes = []cf.RouteSummary{existingRoute}

	appRepo.ReadApp = existingApp

	domain := cf.DomainFields{}
	domain.Name = "example.com"

	foundRoute := cf.Route{}
	foundRoute.RouteFields = existingRoute.RouteFields
	foundRoute.Domain = domain

	routeRepo.FindByHostAndDomainRoute = foundRoute

	stack := cf.Stack{}
	stack.Name = "differentStack"
	stack.Guid = "differentStack-guid"
	stackRepo.FindByNameStack = stack

	args := []string{
		"-c", "different start command",
		"-i", "10",
		"-m", "1G",
		"-b", "https://github.com/heroku/heroku-buildpack-different.git",
		"-s", "differentStack",
		"existing-app",
	}
	_ = callPush(t, args, manifestRepo, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo)

	assert.Equal(t, appRepo.UpdateParams.Get("command"), "different start command")
	assert.Equal(t, appRepo.UpdateParams.Get("instances"), 10)
	assert.Equal(t, appRepo.UpdateParams.Get("memory"), uint64(1024))
	assert.Equal(t, appRepo.UpdateParams.Get("buildpack"), "https://github.com/heroku/heroku-buildpack-different.git")
	assert.Equal(t, appRepo.UpdateParams.Get("stack_guid"), "differentStack-guid")
}

func TestPushingAppWhenItAlreadyExistsAndDomainIsSpecifiedIsAlreadyBound(t *testing.T) {
	manifestRepo, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo := getPushDependencies()

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

	appRepo.ReadApp = existingApp
	appRepo.UpdateAppResult = existingApp
	routeRepo.FindByHostAndDomainRoute = foundRoute

	ui := callPush(t, []string{"-d", "example.com", "existing-app"}, manifestRepo, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Using route", "existing-app", "example.com"},
	})
	assert.Equal(t, appBitsRepo.UploadedAppGuid, "existing-app-guid")
}

func TestPushingAppWhenItAlreadyExistsAndDomainSpecifiedIsNotBound(t *testing.T) {
	manifestRepo, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo := getPushDependencies()

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

	appRepo.ReadApp = existingApp
	appRepo.UpdateAppResult = existingApp
	routeRepo.FindByHostAndDomainNotFound = true
	domainRepo.FindByNameDomain = foundDomain

	ui := callPush(t, []string{"-d", "newdomain.com", "existing-app"}, manifestRepo, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Creating route", "existing-app.newdomain.com"},
		{"OK"},
		{"Binding", "existing-app.newdomain.com"},
	})

	assert.Equal(t, appBitsRepo.UploadedAppGuid, "existing-app-guid")
	assert.Equal(t, domainRepo.FindByNameInCurrentSpaceName, "newdomain.com")
	assert.Equal(t, routeRepo.FindByHostAndDomainDomain, "newdomain.com")
	assert.Equal(t, routeRepo.FindByHostAndDomainHost, "existing-app")
	assert.Equal(t, routeRepo.CreatedHost, "existing-app")
	assert.Equal(t, routeRepo.CreatedDomainGuid, "domain-guid")
}

func TestPushingAppWithNoFlagsWhenAppIsAlreadyBoundToDomain(t *testing.T) {
	manifestRepo, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo := getPushDependencies()

	domain := cf.DomainFields{}
	domain.Name = "example.com"

	existingRoute := cf.RouteSummary{}
	existingRoute.Host = "foo"
	existingRoute.Domain = domain

	existingApp := cf.Application{}
	existingApp.Name = "existing-app"
	existingApp.Guid = "existing-app-guid"
	existingApp.Routes = []cf.RouteSummary{existingRoute}

	appRepo.ReadApp = existingApp
	appRepo.UpdateAppResult = existingApp

	_ = callPush(t, []string{"existing-app"}, manifestRepo, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo)

	assert.Equal(t, appBitsRepo.UploadedAppGuid, "existing-app-guid")
	assert.Equal(t, domainRepo.FindByNameInCurrentSpaceName, "")
	assert.Equal(t, routeRepo.FindByHostAndDomainDomain, "")
	assert.Equal(t, routeRepo.FindByHostAndDomainHost, "")
	assert.Equal(t, routeRepo.CreatedHost, "")
	assert.Equal(t, routeRepo.CreatedDomainGuid, "")
}

func TestPushingAppWhenItAlreadyExistsAndHostIsSpecified(t *testing.T) {
	manifestRepo, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo := getPushDependencies()

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

	appRepo.ReadApp = existingApp
	appRepo.UpdateAppResult = existingApp
	routeRepo.FindByHostAndDomainNotFound = true
	domainRepo.ListDomainsForOrgDomains = []cf.Domain{domain}

	ui := callPush(t, []string{"-n", "new-host", "existing-app"}, manifestRepo, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Creating route", "new-host.example.com"},
		{"OK"},
		{"Binding", "new-host.example.com"},
	})

	assert.Equal(t, routeRepo.FindByHostAndDomainDomain, "example.com")
	assert.Equal(t, routeRepo.FindByHostAndDomainHost, "new-host")
	assert.Equal(t, routeRepo.CreatedHost, "new-host")
	assert.Equal(t, routeRepo.CreatedDomainGuid, "domain-guid")
}

func TestPushingAppWhenItAlreadyExistsAndNoRouteFlagIsPresent(t *testing.T) {
	manifestRepo, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo := getPushDependencies()
	existingApp := cf.Application{}
	existingApp.Name = "existing-app"
	existingApp.Guid = "existing-app-guid"

	appRepo.ReadApp = existingApp
	appRepo.UpdateAppResult = existingApp

	ui := callPush(t, []string{"--no-route", "existing-app"}, manifestRepo, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Uploading", "existing-app"},
		{"OK"},
	})

	assert.Equal(t, appBitsRepo.UploadedAppGuid, "existing-app-guid")
	assert.Equal(t, domainRepo.FindByNameInCurrentSpaceName, "")
	assert.Equal(t, routeRepo.FindByHostAndDomainDomain, "")
	assert.Equal(t, routeRepo.FindByHostAndDomainHost, "")
	assert.Equal(t, routeRepo.CreatedHost, "")
	assert.Equal(t, routeRepo.CreatedDomainGuid, "")
}

func TestPushingAppWhenItAlreadyExistsAndNoHostFlagIsPresent(t *testing.T) {
	manifestRepo, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo := getPushDependencies()

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

	appRepo.ReadApp = existingApp
	appRepo.UpdateAppResult = existingApp
	routeRepo.FindByHostAndDomainNotFound = true
	domainRepo.ListDomainsForOrgDomains = []cf.Domain{domain}

	ui := callPush(t, []string{"--no-hostname", "existing-app"}, manifestRepo, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Creating route", "example.com"},
		{"OK"},
		{"Binding", "example.com"},
	})
	testassert.SliceDoesNotContain(t, ui.Outputs, testassert.Lines{
		{"existing-app.example.com"},
	})

	assert.Equal(t, routeRepo.FindByHostAndDomainDomain, "example.com")
	assert.Equal(t, routeRepo.FindByHostAndDomainHost, "")
	assert.Equal(t, routeRepo.CreatedHost, "")
	assert.Equal(t, routeRepo.CreatedDomainGuid, "domain-guid")
}

func TestPushingAppWhenItAlreadyExistsWithoutARouteAndARouteIsNotProvided(t *testing.T) {
	manifestRepo, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo := getPushDependencies()
	existingApp := cf.Application{}
	existingApp.Name = "existing-app"
	existingApp.Guid = "existing-app-guid"

	appRepo.ReadApp = existingApp
	appRepo.UpdateAppResult = existingApp

	ui := callPush(t, []string{"existing-app"}, manifestRepo, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		testassert.Line{"skipping route creation"},
		testassert.Line{"Uploading"},
		testassert.Line{"OK"},
	})

	assert.Equal(t, routeRepo.FindByHostAndDomainDomain, "")
	assert.Equal(t, routeRepo.FindByHostAndDomainHost, "")
	assert.Equal(t, routeRepo.CreatedHost, "")
	assert.Equal(t, routeRepo.CreatedDomainGuid, "")
}

func TestPushingAppWithInvalidPath(t *testing.T) {
	manifestRepo, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo := getPushDependencies()
	appBitsRepo.UploadAppErr = true

	ui := callPush(t, []string{"app"}, manifestRepo, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Uploading"},
		{"FAILED"},
	})
}

func getPushDependencies() (
	manifestRepo *testmanifest.FakeManifestRepository,
	starter *testcmd.FakeAppStarter,
	stopper *testcmd.FakeAppStopper,
	appRepo *testapi.FakeApplicationRepository,
	domainRepo *testapi.FakeDomainRepository,
	routeRepo *testapi.FakeRouteRepository,
	stackRepo *testapi.FakeStackRepository,
	appBitsRepo *testapi.FakeApplicationBitsRepository) {

	manifestRepo = &testmanifest.FakeManifestRepository{}
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
	manifestRepo *testmanifest.FakeManifestRepository,
	starter ApplicationStarter,
	stopper ApplicationStopper,
	appRepo api.ApplicationRepository,
	domainRepo api.DomainRepository,
	routeRepo api.RouteRepository,
	stackRepo api.StackRepository,
	appBitsRepo *testapi.FakeApplicationBitsRepository) (ui *testterm.FakeUI) {

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

	cmd := NewPush(ui, config, manifestRepo, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo)
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}
	testcmd.RunCommand(cmd, ctxt, reqFactory)

	return
}
