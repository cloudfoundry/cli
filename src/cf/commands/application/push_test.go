package application_test

import (
	. "cf/commands/application"
	"cf/configuration"
	"cf/manifest"
	"cf/models"
	"cf/net"
	"errors"
	"generic"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	"os"
	"path/filepath"
	"syscall"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	"testhelpers/maker"
	testmanifest "testhelpers/manifest"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

func singleAppManifest() *manifest.Manifest {
	name := "manifest-app-name"
	memory := uint64(128)
	instances := 1
	host := "manifest-host"
	domain := "manifest-example.com"
	stack := "custom-stack"
	timeout := 360
	buildpackUrl := "some-buildpack"
	command := `JAVA_HOME=$PWD/.openjdk JAVA_OPTS="-Xss995K" ./bin/start.sh run`
	path := "/some/path/from/manifest"

	return &manifest.Manifest{
		Applications: []models.AppParams{
			models.AppParams{
				Name:               &name,
				Memory:             &memory,
				InstanceCount:      &instances,
				Host:               &host,
				Domain:             &domain,
				StackName:          &stack,
				HealthCheckTimeout: &timeout,
				BuildpackUrl:       &buildpackUrl,
				Command:            &command,
				Path:               &path,
				EnvironmentVars: &map[string]string{
					"FOO":  "baz",
					"PATH": "/u/apps/my-app/bin",
				},
			},
		},
	}
}

func manifestWithServicesAndEnv() *manifest.Manifest {
	name1 := "app1"
	name2 := "app2"
	return &manifest.Manifest{
		Applications: []models.AppParams{
			models.AppParams{
				Name:     &name1,
				Services: &[]string{"app1-service", "global-service"},
				EnvironmentVars: &map[string]string{
					"SOMETHING": "definitely-something",
				},
			},
			models.AppParams{
				Name:     &name2,
				Services: &[]string{"app2-service", "global-service"},
				EnvironmentVars: &map[string]string{
					"SOMETHING": "nothing",
				},
			},
		},
	}
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
	sharedDomain := maker.NewSharedDomain(maker.Overrides{"name": "foo.cf-app.com", "guid": "foo-domain-guid"})
	deps.domainRepo.ListSharedDomainsDomains = []models.Domain{sharedDomain}

	deps.routeRepo = &testapi.FakeRouteRepository{}
	deps.stackRepo = &testapi.FakeStackRepository{}
	deps.appBitsRepo = &testapi.FakeApplicationBitsRepository{}
	deps.serviceRepo = &testapi.FakeServiceRepo{}

	return
}

func callPush(t mr.TestingT, args []string, deps pushDependencies) (ui *testterm.FakeUI) {

	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("push", args)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)
	org := maker.NewOrgFields(maker.Overrides{"name": "my-org"})
	space := maker.NewSpaceFields(maker.Overrides{"name": "my-space", "guid": "my-space-guid"})

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
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestPushingRequirements", func() {
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
			testcmd.RunCommand(cmd, ctxt, reqFactory)
			assert.True(mr.T(), testcmd.CommandDidPassRequirements)

			reqFactory = &testreq.FakeReqFactory{LoginSuccess: false, TargetedSpaceSuccess: true}
			testcmd.RunCommand(cmd, ctxt, reqFactory)
			assert.False(mr.T(), testcmd.CommandDidPassRequirements)

			testcmd.CommandDidPassRequirements = true

			reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: false}
			testcmd.RunCommand(cmd, ctxt, reqFactory)
			assert.False(mr.T(), testcmd.CommandDidPassRequirements)
		})

		It("TestPushingAppWithOldV2DomainsEndpoint", func() {
			deps := getPushDependencies()

			privateDomain := models.Domain{}
			privateDomain.Shared = false
			privateDomain.Name = "private.cf-app.com"
			privateDomain.Guid = "private-domain-guid"

			sharedDomain := models.Domain{}
			sharedDomain.Name = "shared.cf-app.com"
			sharedDomain.Shared = true
			sharedDomain.Guid = "shared-domain-guid"

			deps.domainRepo.ListSharedDomainsApiResponse = net.NewNotFoundApiResponse("whoopsie")
			deps.domainRepo.ListDomainsDomains = []models.Domain{privateDomain, sharedDomain}
			deps.routeRepo.FindByHostAndDomainErr = true
			deps.appRepo.ReadNotFound = true

			ui := callPush(mr.T(), []string{"-t", "111", "my-new-app"}, deps)

			assert.Equal(mr.T(), deps.routeRepo.FindByHostAndDomainHost, "my-new-app")
			assert.Equal(mr.T(), deps.routeRepo.CreatedHost, "my-new-app")
			assert.Equal(mr.T(), deps.routeRepo.CreatedDomainGuid, "shared-domain-guid")
			assert.Equal(mr.T(), deps.routeRepo.BoundAppGuid, "my-new-app-guid")
			assert.Equal(mr.T(), deps.routeRepo.BoundRouteGuid, "my-new-app-route-guid")

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Creating app", "my-new-app", "my-org", "my-space"},
				{"OK"},
				{"Creating", "my-new-app.shared.cf-app.com"},
				{"OK"},
				{"Binding", "my-new-app.shared.cf-app.com"},
				{"OK"},
				{"Uploading my-new-app"},
				{"OK"},
			})
		})

		It("TestPushingAppWhenItDoesNotExist", func() {
			deps := getPushDependencies()

			deps.routeRepo.FindByHostAndDomainErr = true
			deps.appRepo.ReadNotFound = true

			ui := callPush(mr.T(), []string{"-t", "111", "my-new-app"}, deps)
			assert.Equal(mr.T(), *deps.appRepo.CreatedAppParams().Name, "my-new-app")
			assert.Equal(mr.T(), *deps.appRepo.CreatedAppParams().SpaceGuid, "my-space-guid")

			assert.Equal(mr.T(), deps.routeRepo.FindByHostAndDomainHost, "my-new-app")
			assert.Equal(mr.T(), deps.routeRepo.CreatedHost, "my-new-app")
			assert.Equal(mr.T(), deps.routeRepo.CreatedDomainGuid, "foo-domain-guid")
			assert.Equal(mr.T(), deps.routeRepo.BoundAppGuid, "my-new-app-guid")
			assert.Equal(mr.T(), deps.routeRepo.BoundRouteGuid, "my-new-app-route-guid")

			assert.Equal(mr.T(), deps.appBitsRepo.UploadedAppGuid, "my-new-app-guid")

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Creating app", "my-new-app", "my-org", "my-space"},
				{"OK"},
				{"Creating", "my-new-app.foo.cf-app.com"},
				{"OK"},
				{"Binding", "my-new-app.foo.cf-app.com"},
				{"OK"},
				{"Uploading my-new-app"},
				{"OK"},
			})

			assert.Equal(mr.T(), deps.stopper.AppToStop.Guid, "")
			assert.Equal(mr.T(), deps.starter.AppToStart.Guid, "my-new-app-guid")
			assert.Equal(mr.T(), deps.starter.AppToStart.Name, "my-new-app")
			assert.Equal(mr.T(), deps.starter.Timeout, 111)
		})

		It("TestPushingAppWithACrazyName", func() {
			deps := getPushDependencies()

			deps.routeRepo.FindByHostAndDomainErr = true
			deps.appRepo.ReadNotFound = true

			ui := callPush(mr.T(), []string{"-t", "111", "Tim's 1st-Crazy__app!"}, deps)
			assert.Equal(mr.T(), *deps.appRepo.CreatedAppParams().Name, "Tim's 1st-Crazy__app!")

			assert.Equal(mr.T(), deps.routeRepo.FindByHostAndDomainHost, "tims-1st-crazy-app")
			assert.Equal(mr.T(), deps.routeRepo.CreatedHost, "tims-1st-crazy-app")

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Creating", "tims-1st-crazy-app.foo.cf-app.com"},
				{"Binding", "tims-1st-crazy-app.foo.cf-app.com"},
			})
			testassert.SliceDoesNotContain(mr.T(), ui.Outputs, testassert.Lines{
				{"FAILED"},
			})
		})
		It("TestPushingAppWhenItDoesNotExistButRouteExists", func() {

			deps := getPushDependencies()

			route := models.Route{}
			route.Guid = "my-route-guid"
			route.Host = "my-new-app"
			route.Domain = deps.domainRepo.ListSharedDomainsDomains[0].DomainFields

			deps.routeRepo.FindByHostAndDomainRoute = route
			deps.appRepo.ReadNotFound = true

			ui := callPush(mr.T(), []string{"my-new-app"}, deps)

			assert.Empty(mr.T(), deps.routeRepo.CreatedHost)
			assert.Empty(mr.T(), deps.routeRepo.CreatedDomainGuid)
			assert.Equal(mr.T(), deps.routeRepo.FindByHostAndDomainHost, "my-new-app")
			assert.Equal(mr.T(), deps.routeRepo.BoundAppGuid, "my-new-app-guid")
			assert.Equal(mr.T(), deps.routeRepo.BoundRouteGuid, "my-route-guid")

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Using", "my-new-app.foo.cf-app.com"},
				{"Binding", "my-new-app.foo.cf-app.com"},
				{"OK"},
			})
		})
		It("TestPushingAppWithCustomFlags", func() {

			deps := getPushDependencies()
			domain := models.Domain{}
			domain.Name = "bar.cf-app.com"
			domain.Guid = "bar-domain-guid"
			stack := models.Stack{}
			stack.Name = "customLinux"
			stack.Guid = "custom-linux-guid"

			deps.domainRepo.FindByNameDomain = domain
			deps.routeRepo.FindByHostAndDomainErr = true
			deps.stackRepo.FindByNameStack = stack
			deps.appRepo.ReadNotFound = true

			ui := callPush(mr.T(), []string{
				"-c", "unicorn -c config/unicorn.rb -D",
				"-d", "bar.cf-app.com",
				"-n", "my-hostname",
				"-i", "3",
				"-m", "2G",
				"-b", "https://github.com/heroku/heroku-buildpack-play.git",
				"-s", "customLinux",
				"-t", "1",
				"--no-start",
				"my-new-app",
			}, deps)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
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

			assert.Equal(mr.T(), deps.stackRepo.FindByNameName, "customLinux")

			assert.Equal(mr.T(), *deps.appRepo.CreatedAppParams().Name, "my-new-app")
			assert.Equal(mr.T(), *deps.appRepo.CreatedAppParams().Command, "unicorn -c config/unicorn.rb -D")
			assert.Equal(mr.T(), *deps.appRepo.CreatedAppParams().InstanceCount, 3)
			assert.Equal(mr.T(), *deps.appRepo.CreatedAppParams().Memory, uint64(2048))
			assert.Equal(mr.T(), *deps.appRepo.CreatedAppParams().StackGuid, "custom-linux-guid")
			assert.Equal(mr.T(), *deps.appRepo.CreatedAppParams().HealthCheckTimeout, 1)
			assert.Equal(mr.T(), *deps.appRepo.CreatedAppParams().BuildpackUrl, "https://github.com/heroku/heroku-buildpack-play.git")

			assert.Equal(mr.T(), deps.domainRepo.FindByNameInCurrentSpaceName, "bar.cf-app.com")

			assert.Equal(mr.T(), deps.routeRepo.CreatedHost, "my-hostname")
			assert.Equal(mr.T(), deps.routeRepo.CreatedDomainGuid, "bar-domain-guid")
			assert.Equal(mr.T(), deps.routeRepo.BoundAppGuid, "my-new-app-guid")
			assert.Equal(mr.T(), deps.routeRepo.BoundRouteGuid, "my-hostname-route-guid")

			assert.Equal(mr.T(), deps.appBitsRepo.UploadedAppGuid, "my-new-app-guid")

			assert.Equal(mr.T(), deps.starter.AppToStart.Name, "")
		})
		It("TestPushingAppWithInvalidTimeout", func() {

			deps := getPushDependencies()
			deps.appRepo.ReadNotFound = true

			ui := callPush(mr.T(), []string{
				"-t", "FooeyTimeout",
				"my-new-app",
			}, deps)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"FAILED"},
				{"invalid", "timeout"},
			})
		})
		It("TestPushingAppToResetStartCommand", func() {

			deps := getPushDependencies()

			existingApp := models.Application{}
			existingApp.Name = "existing-app"
			existingApp.Guid = "existing-app-guid"
			existingApp.Command = "unicorn -c config/unicorn.rb -D"

			deps.appRepo.ReadApp = existingApp

			args := []string{
				"-c", "null",
				"existing-app",
			}
			_ = callPush(mr.T(), args, deps)

			assert.Equal(mr.T(), *deps.appRepo.UpdateParams.Command, "")
		})

		It("TestPushingAppMergesManifestEnvVarsWithExistingEnvVars", func() {

			deps := getPushDependencies()

			existingApp := maker.NewApp(maker.Overrides{"name": "existing-app"})
			existingApp.EnvironmentVars = map[string]string{
				"crazy": "pants",
				"FOO":   "NotYoBaz",
				"foo":   "manchu",
			}
			deps.appRepo.ReadApp = existingApp

			deps.manifestRepo.ReadManifestManifest = singleAppManifest()

			_ = callPush(mr.T(), []string{"existing-app"}, deps)

			updatedAppEnvVars := *deps.appRepo.UpdateParams.EnvironmentVars
			assert.Equal(mr.T(), updatedAppEnvVars["crazy"], "pants")
			assert.Equal(mr.T(), updatedAppEnvVars["FOO"], "baz")
			assert.Equal(mr.T(), updatedAppEnvVars["foo"], "manchu")
			assert.Equal(mr.T(), updatedAppEnvVars["PATH"], "/u/apps/my-app/bin")
		})

		It("TestPushingAppWithSingleAppManifest", func() {
			deps := getPushDependencies()
			domain := models.Domain{}
			domain.Name = "manifest-example.com"
			domain.Guid = "bar-domain-guid"
			deps.domainRepo.FindByNameDomain = domain
			deps.routeRepo.FindByHostAndDomainErr = true
			deps.appRepo.ReadNotFound = true

			deps.manifestRepo.ReadManifestManifest = singleAppManifest()

			ui := callPush(mr.T(), []string{}, deps)
			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Creating route", "manifest-host.manifest-example.com"},
				{"OK"},
				{"Binding", "manifest-host.manifest-example.com"},
				{"manifest-app-name"},
			})

			assert.Equal(mr.T(), *deps.appRepo.CreatedAppParams().Name, "manifest-app-name")
			assert.Equal(mr.T(), *deps.appRepo.CreatedAppParams().Memory, uint64(128))
			assert.Equal(mr.T(), *deps.appRepo.CreatedAppParams().InstanceCount, 1)
			assert.Equal(mr.T(), *deps.appRepo.CreatedAppParams().StackName, "custom-stack")
			assert.Equal(mr.T(), *deps.appRepo.CreatedAppParams().BuildpackUrl, "some-buildpack")
			assert.Equal(mr.T(), *deps.appRepo.CreatedAppParams().Command, "JAVA_HOME=$PWD/.openjdk JAVA_OPTS=\"-Xss995K\" ./bin/start.sh run")
			assert.Equal(mr.T(), deps.appBitsRepo.UploadedDir, "/some/path/from/manifest")

			envVars := *deps.appRepo.CreatedAppParams().EnvironmentVars
			assert.Equal(mr.T(), envVars, map[string]string{
				"PATH": "/u/apps/my-app/bin",
				"FOO":  "baz",
			})
		})

		It("TestPushingAppManifestWithErrors", func() {
			deps := getPushDependencies()
			deps.appRepo.ReadNotFound = true
			deps.manifestRepo.ReadManifestErrors = manifest.ManifestErrors{
				errors.New("buildpack should not be null"),
				errors.New("disk_quota should not be null"),
			}

			ui := callPush(mr.T(), []string{}, deps)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"FAILED"},
				{"Error", "reading", "manifest"},
				{"buildpack should not be null"},
				{"disk_quota should not be null"},
			})
		})

		It("TestPushingManyAppsFromManifest", func() {
			deps := getPushDependencies()
			deps.appRepo.ReadNotFound = true
			deps.manifestRepo.ReadManifestManifest = manifestWithServicesAndEnv()

			ui := callPush(mr.T(), []string{}, deps)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Creating", "app1"},
				{"Creating", "app2"},
			})
			assert.Equal(mr.T(), len(deps.appRepo.CreateAppParams), 2)

			firstApp := deps.appRepo.CreateAppParams[0]
			secondApp := deps.appRepo.CreateAppParams[1]
			assert.Equal(mr.T(), *firstApp.Name, "app1")
			assert.Equal(mr.T(), *secondApp.Name, "app2")

			envVars := *firstApp.EnvironmentVars
			assert.Equal(mr.T(), envVars["SOMETHING"], "definitely-something")

			envVars = *secondApp.EnvironmentVars
			assert.Equal(mr.T(), envVars["SOMETHING"], "nothing")
		})

		It("TestPushingASingleAppFromAManifestWithManyApps", func() {
			deps := getPushDependencies()
			deps.appRepo.ReadNotFound = true
			deps.manifestRepo.ReadManifestManifest = manifestWithServicesAndEnv()

			ui := callPush(mr.T(), []string{"app2"}, deps)
			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Creating", "app2"},
			})
			testassert.SliceDoesNotContain(mr.T(), ui.Outputs, testassert.Lines{
				{"Creating", "app1"},
			})
			assert.Equal(mr.T(), len(deps.appRepo.CreateAppParams), 1)
			assert.Equal(mr.T(), *deps.appRepo.CreateAppParams[0].Name, "app2")
		})

		It("TestNamedAppInAManifestIsNotFound", func() {
			deps := getPushDependencies()
			deps.appRepo.ReadNotFound = true
			deps.manifestRepo.ReadManifestManifest = manifestWithServicesAndEnv()

			ui := callPush(mr.T(), []string{"non-existant-app"}, deps)
			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Failed"},
			})
			assert.Equal(mr.T(), len(deps.appRepo.CreateAppParams), 0)
		})
		It("TestPushingWithBindingMergedServices", func() {

			deps := getPushDependencies()
			deps.appRepo.ReadNotFound = true

			deps.serviceRepo.FindInstanceByNameMap = generic.NewMap(map[interface{}]interface{}{
				"global-service": maker.NewServiceInstance("global-service"),
				"app1-service":   maker.NewServiceInstance("app1-service"),
				"app2-service":   maker.NewServiceInstance("app2-service"),
			})

			deps.manifestRepo.ReadManifestManifest = manifestWithServicesAndEnv()

			ui := callPush(mr.T(), []string{}, deps)
			assert.Equal(mr.T(), len(deps.binder.AppsToBind), 4)
			assert.Equal(mr.T(), deps.binder.AppsToBind[0].Name, "app1")
			assert.Equal(mr.T(), deps.binder.AppsToBind[1].Name, "app1")
			assert.Equal(mr.T(), deps.binder.InstancesToBindTo[0].Name, "app1-service")
			assert.Equal(mr.T(), deps.binder.InstancesToBindTo[1].Name, "global-service")

			assert.Equal(mr.T(), deps.binder.AppsToBind[2].Name, "app2")
			assert.Equal(mr.T(), deps.binder.AppsToBind[3].Name, "app2")
			assert.Equal(mr.T(), deps.binder.InstancesToBindTo[2].Name, "app2-service")
			assert.Equal(mr.T(), deps.binder.InstancesToBindTo[3].Name, "global-service")

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Creating", "app1"},
				{"OK"},
				{"Binding service", "app1-service", "app1", "my-org", "my-space", "my-user"},
				{"OK"},
				{"Binding service", "global-service", "app1", "my-org", "my-space", "my-user"},
				{"OK"},
				{"Creating", "app2"},
				{"OK"},
				{"Binding service", "app2-service", "app2", "my-org", "my-space", "my-user"},
				{"OK"},
				{"Binding service", "global-service", "app2", "my-org", "my-space", "my-user"},
				{"OK"},
			})
		})
		It("TestPushWithServicesThatAreNotFound", func() {

			deps := getPushDependencies()
			deps.routeRepo.FindByHostAndDomainErr = true
			deps.serviceRepo.FindInstanceByNameErr = true
			deps.manifestRepo.ReadManifestManifest = manifestWithServicesAndEnv()

			ui := callPush(mr.T(), []string{}, deps)
			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"FAILED"},
				{"Could not find service", "app1-service", "app1"},
			})
		})
		It("TestPushingAppWithPath", func() {

			deps := getPushDependencies()
			deps.appRepo.ReadNotFound = true

			absPath, err := filepath.Abs("../../../fixtures/example-app")
			assert.NoError(mr.T(), err)

			callPush(mr.T(), []string{
				"-p", absPath,
				"app-with-path",
			}, deps)

			assert.Equal(mr.T(), deps.appBitsRepo.UploadedDir, absPath)
		})
		It("TestPushingAppWithPathToZipFile", func() {

			deps := getPushDependencies()
			deps.appRepo.ReadNotFound = true

			absPath, err := filepath.Abs("../../../fixtures/example-app.jar")
			assert.NoError(mr.T(), err)

			callPush(mr.T(), []string{
				"-p", absPath,
				"app-with-path",
			}, deps)

			assert.Equal(mr.T(), deps.appBitsRepo.UploadedDir, absPath)
		})
		It("TestPushingWithDefaultAppPath", func() {

			deps := getPushDependencies()
			deps.appRepo.ReadNotFound = true

			callPush(mr.T(), []string{"app-with-default-path"}, deps)

			dir, err := os.Getwd()
			assert.NoError(mr.T(), err)
			assert.Equal(mr.T(), deps.appBitsRepo.UploadedDir, dir)
		})
		It("TestPushingWithRelativeAppPath", func() {

			deps := getPushDependencies()
			deps.appRepo.ReadNotFound = true

			callPush(mr.T(), []string{
				"-p", "../../../fixtures/example-app",
				"app-with-relative-path",
			}, deps)

			dir, err := os.Getwd()
			assert.NoError(mr.T(), err)
			assert.Equal(mr.T(), deps.appBitsRepo.UploadedDir, filepath.Join(dir, "../../../fixtures/example-app"))
		})
		It("TestPushingWithBadManifestPath", func() {

			deps := getPushDependencies()
			deps.appRepo.ReadNotFound = true

			deps.manifestRepo.ReadManifestManifest = manifest.NewEmptyManifest()
			deps.manifestRepo.ManifestPathErr = errors.New("read manifest error")

			ui := callPush(mr.T(), []string{
				"-f", "bad/manifest/path",
			}, deps)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"FAILED"},
				{"read manifest error"},
			})
		})
		It("TestPushingWithDefaultManifestNotFound", func() {

			deps := getPushDependencies()
			deps.appRepo.ReadNotFound = true
			deps.manifestRepo.ReadManifestManifest = singleAppManifest()
			deps.manifestRepo.ReadManifestErrors = manifest.ManifestErrors{syscall.ENOENT}

			ui := callPush(mr.T(), []string{"--no-route", "app-name"}, deps)
			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Creating app", "app-name"},
				{"OK"},
				{"Uploading", "app-name"},
				{"OK"},
			})
			testassert.SliceDoesNotContain(mr.T(), ui.Outputs, testassert.Lines{
				{"FAILED"},
			})
		})
		It("TestPushingWithManifestInAppDirectory", func() {

			deps := getPushDependencies()
			deps.appRepo.ReadNotFound = true
			deps.manifestRepo.ReadManifestManifest = singleAppManifest()

			ui := callPush(mr.T(), []string{"-p", "some/relative/path"}, deps)
			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Using manifest file", "manifest.yml"},
			})

			assert.Equal(mr.T(), deps.manifestRepo.UserSpecifiedPath, "")
			assert.Equal(mr.T(), deps.manifestRepo.ReadManifestPath, "manifest.yml")
		})
		It("TestPushingWithNoManifestFlag", func() {

			deps := getPushDependencies()
			deps.appRepo.ReadNotFound = true

			ui := callPush(mr.T(), []string{"--no-route", "--no-manifest", "app-name"}, deps)

			testassert.SliceDoesNotContain(mr.T(), ui.Outputs, testassert.Lines{
				{"FAILED"},
				{"hacker-manifesto"},
			})

			assert.Equal(mr.T(), deps.manifestRepo.ReadManifestPath, "")
			assert.Equal(mr.T(), *deps.appRepo.CreatedAppParams().Name, "app-name")
		})
		It("TestPushingWithNoManifestFlagAndMissingAppName", func() {

			deps := getPushDependencies()
			deps.appRepo.ReadNotFound = true

			ui := callPush(mr.T(), []string{"--no-route", "--no-manifest"}, deps)
			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"FAILED"},
			})
		})
		It("TestPushingAppWithNoRoute", func() {

			deps := getPushDependencies()
			domain := models.Domain{}
			domain.Name = "bar.cf-app.com"
			domain.Guid = "bar-domain-guid"

			deps.domainRepo.FindByNameDomain = domain
			deps.routeRepo.FindByHostErr = true
			deps.appRepo.ReadNotFound = true

			callPush(mr.T(), []string{
				"--no-route",
				"my-new-app",
			}, deps)

			assert.Equal(mr.T(), *deps.appRepo.CreatedAppParams().Name, "my-new-app")
			assert.Equal(mr.T(), deps.routeRepo.CreatedHost, "")
			assert.Equal(mr.T(), deps.routeRepo.CreatedDomainGuid, "")
		})
		It("TestPushingAppWithNoHostname", func() {

			deps := getPushDependencies()
			domain := models.Domain{}
			domain.Name = "bar.cf-app.com"
			domain.Guid = "bar-domain-guid"
			domain.Shared = true

			deps.domainRepo.ListSharedDomainsDomains = []models.Domain{domain}
			deps.routeRepo.FindByHostAndDomainErr = true
			deps.appRepo.ReadNotFound = true

			callPush(mr.T(), []string{
				"--no-hostname",
				"my-new-app",
			}, deps)

			assert.Equal(mr.T(), *deps.appRepo.CreatedAppParams().Name, "my-new-app")
			assert.Equal(mr.T(), deps.routeRepo.CreatedHost, "")
			assert.Equal(mr.T(), deps.routeRepo.CreatedDomainGuid, "bar-domain-guid")
		})
		It("TestPushingAppAsWorker", func() {

			deps := getPushDependencies()
			deps.appRepo.ReadNotFound = true

			workerManifest := singleAppManifest()
			noRoute := true
			workerManifest.Applications[0].NoRoute = &noRoute
			deps.manifestRepo.ReadManifestManifest = workerManifest

			ui := callPush(mr.T(), []string{
				"worker-app",
			}, deps)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"worker-app", "is a worker", "skipping route creation"},
			})
			assert.Equal(mr.T(), deps.routeRepo.BoundAppGuid, "")
			assert.Equal(mr.T(), deps.routeRepo.BoundRouteGuid, "")
		})
		It("TestPushingAppWithMemoryInMegaBytes", func() {

			deps := getPushDependencies()
			deps.appRepo.ReadNotFound = true

			callPush(mr.T(), []string{
				"-m", "256M",
				"my-new-app",
			}, deps)

			assert.Equal(mr.T(), *deps.appRepo.CreatedAppParams().Memory, uint64(256))
		})
		It("TestPushingAppWithInvalidMemory", func() {

			deps := getPushDependencies()
			deps.appRepo.ReadNotFound = true

			ui := callPush(mr.T(), []string{
				"-m", "abcM",
				"my-new-app",
			}, deps)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"FAILED"},
				{"invalid", "memory"},
			})
		})
		It("TestPushingAppWhenItAlreadyExistsAndNothingIsSpecified", func() {

			deps := getPushDependencies()
			existingApp := maker.NewApp(maker.Overrides{"name": "existing-app"})
			deps.appRepo.ReadApp = existingApp
			deps.appRepo.UpdateAppResult = existingApp

			_ = callPush(mr.T(), []string{"existing-app"}, deps)

			assert.Equal(mr.T(), deps.stopper.AppToStop.Guid, existingApp.Guid)
			assert.Equal(mr.T(), deps.appBitsRepo.UploadedAppGuid, existingApp.Guid)
		})
		It("TestPushingAppWhenItIsStopped", func() {

			deps := getPushDependencies()
			stoppedApp := maker.NewApp(maker.Overrides{"state": "stopped", "name": "stopped-app"})

			deps.appRepo.ReadApp = stoppedApp
			deps.appRepo.UpdateAppResult = stoppedApp

			_ = callPush(mr.T(), []string{"stopped-app"}, deps)

			assert.Equal(mr.T(), deps.stopper.AppToStop.Guid, "")
		})

		It("TestPushingAppWhenItAlreadyExistsAndChangingOptions", func() {
			deps := getPushDependencies()

			existingRoute := models.RouteSummary{}
			existingRoute.Host = "existing-app"

			existingApp := models.Application{}
			existingApp.Name = "existing-app"
			existingApp.Guid = "existing-app-guid"
			existingApp.Routes = []models.RouteSummary{existingRoute}

			deps.appRepo.ReadApp = existingApp

			stack := models.Stack{}
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
			_ = callPush(mr.T(), args, deps)

			assert.Equal(mr.T(), *deps.appRepo.UpdateParams.Command, "different start command")
			assert.Equal(mr.T(), *deps.appRepo.UpdateParams.InstanceCount, 10)
			assert.Equal(mr.T(), *deps.appRepo.UpdateParams.Memory, uint64(1024))
			assert.Equal(mr.T(), *deps.appRepo.UpdateParams.BuildpackUrl, "https://github.com/heroku/heroku-buildpack-different.git")
			assert.Equal(mr.T(), *deps.appRepo.UpdateParams.StackGuid, "differentStack-guid")
		})

		It("TestPushingAppWhenItAlreadyExistsAndDomainIsSpecifiedIsAlreadyBound", func() {

			deps := getPushDependencies()

			domain := models.DomainFields{}
			domain.Name = "example.com"
			domain.Guid = "domain-guid"

			existingRoute := models.RouteSummary{}
			existingRoute.Host = "existing-app"
			existingRoute.Domain = domain

			existingApp := models.Application{}
			existingApp.Name = "existing-app"
			existingApp.Guid = "existing-app-guid"
			existingApp.Routes = []models.RouteSummary{existingRoute}

			foundRoute := models.Route{}
			foundRoute.RouteFields = existingRoute.RouteFields
			foundRoute.Domain = existingRoute.Domain

			deps.appRepo.ReadApp = existingApp
			deps.appRepo.UpdateAppResult = existingApp
			deps.routeRepo.FindByHostAndDomainRoute = foundRoute

			ui := callPush(mr.T(), []string{"-d", "example.com", "existing-app"}, deps)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Using route", "existing-app", "example.com"},
			})
			assert.Equal(mr.T(), deps.appBitsRepo.UploadedAppGuid, "existing-app-guid")
		})
		It("TestPushingAppWhenItAlreadyExistsAndDomainSpecifiedIsNotBound", func() {

			deps := getPushDependencies()

			domain := models.DomainFields{}
			domain.Name = "example.com"

			existingRoute := models.RouteSummary{}
			existingRoute.Host = "existing-app"
			existingRoute.Domain = domain

			existingApp := models.Application{}
			existingApp.Name = "existing-app"
			existingApp.Guid = "existing-app-guid"
			existingApp.Routes = []models.RouteSummary{existingRoute}

			foundDomain := models.Domain{}
			foundDomain.Guid = "domain-guid"
			foundDomain.Name = "newdomain.com"

			deps.appRepo.ReadApp = existingApp
			deps.appRepo.UpdateAppResult = existingApp
			deps.routeRepo.FindByHostAndDomainNotFound = true
			deps.domainRepo.FindByNameDomain = foundDomain

			ui := callPush(mr.T(), []string{"-d", "newdomain.com", "existing-app"}, deps)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Creating route", "existing-app.newdomain.com"},
				{"OK"},
				{"Binding", "existing-app.newdomain.com"},
			})

			assert.Equal(mr.T(), deps.appBitsRepo.UploadedAppGuid, "existing-app-guid")
			assert.Equal(mr.T(), deps.domainRepo.FindByNameInCurrentSpaceName, "newdomain.com")
			assert.Equal(mr.T(), deps.routeRepo.FindByHostAndDomainDomain, "newdomain.com")
			assert.Equal(mr.T(), deps.routeRepo.FindByHostAndDomainHost, "existing-app")
			assert.Equal(mr.T(), deps.routeRepo.CreatedHost, "existing-app")
			assert.Equal(mr.T(), deps.routeRepo.CreatedDomainGuid, "domain-guid")
		})
		It("TestPushingAppWithNoFlagsWhenAppIsAlreadyBoundToDomain", func() {

			deps := getPushDependencies()

			domain := models.DomainFields{}
			domain.Name = "example.com"

			existingRoute := models.RouteSummary{}
			existingRoute.Host = "foo"
			existingRoute.Domain = domain

			existingApp := models.Application{}
			existingApp.Name = "existing-app"
			existingApp.Guid = "existing-app-guid"
			existingApp.Routes = []models.RouteSummary{existingRoute}

			deps.appRepo.ReadApp = existingApp
			deps.appRepo.UpdateAppResult = existingApp

			_ = callPush(mr.T(), []string{"existing-app"}, deps)

			assert.Equal(mr.T(), deps.appBitsRepo.UploadedAppGuid, "existing-app-guid")
			assert.Equal(mr.T(), deps.domainRepo.FindByNameInCurrentSpaceName, "")
			assert.Equal(mr.T(), deps.routeRepo.FindByHostAndDomainDomain, "")
			assert.Equal(mr.T(), deps.routeRepo.FindByHostAndDomainHost, "")
			assert.Equal(mr.T(), deps.routeRepo.CreatedHost, "")
			assert.Equal(mr.T(), deps.routeRepo.CreatedDomainGuid, "")
		})
		It("TestPushingAppWhenItAlreadyExistsAndHostIsSpecified", func() {

			deps := getPushDependencies()

			domain := models.Domain{}
			domain.Name = "example.com"
			domain.Guid = "domain-guid"
			domain.Shared = true

			existingRoute := models.RouteSummary{}
			existingRoute.Host = "existing-app"
			existingRoute.Domain = domain.DomainFields

			existingApp := models.Application{}
			existingApp.Name = "existing-app"
			existingApp.Guid = "existing-app-guid"
			existingApp.Routes = []models.RouteSummary{existingRoute}

			deps.appRepo.ReadApp = existingApp
			deps.appRepo.UpdateAppResult = existingApp
			deps.routeRepo.FindByHostAndDomainNotFound = true
			deps.domainRepo.ListSharedDomainsDomains = []models.Domain{domain}

			ui := callPush(mr.T(), []string{"-n", "new-host", "existing-app"}, deps)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Creating route", "new-host.example.com"},
				{"OK"},
				{"Binding", "new-host.example.com"},
			})

			assert.Equal(mr.T(), deps.routeRepo.FindByHostAndDomainDomain, "example.com")
			assert.Equal(mr.T(), deps.routeRepo.FindByHostAndDomainHost, "new-host")
			assert.Equal(mr.T(), deps.routeRepo.CreatedHost, "new-host")
			assert.Equal(mr.T(), deps.routeRepo.CreatedDomainGuid, "domain-guid")
		})
		It("TestPushingAppWhenItAlreadyExistsAndNoRouteFlagIsPresent", func() {

			deps := getPushDependencies()
			existingApp := models.Application{}
			existingApp.Name = "existing-app"
			existingApp.Guid = "existing-app-guid"

			deps.appRepo.ReadApp = existingApp
			deps.appRepo.UpdateAppResult = existingApp

			ui := callPush(mr.T(), []string{"--no-route", "existing-app"}, deps)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Uploading", "existing-app"},
				{"OK"},
			})

			assert.Equal(mr.T(), deps.appBitsRepo.UploadedAppGuid, "existing-app-guid")
			assert.Equal(mr.T(), deps.domainRepo.FindByNameInCurrentSpaceName, "")
			assert.Equal(mr.T(), deps.routeRepo.FindByHostAndDomainDomain, "")
			assert.Equal(mr.T(), deps.routeRepo.FindByHostAndDomainHost, "")
			assert.Equal(mr.T(), deps.routeRepo.CreatedHost, "")
			assert.Equal(mr.T(), deps.routeRepo.CreatedDomainGuid, "")
		})
		It("TestPushingAppWhenItAlreadyExistsAndNoHostFlagIsPresent", func() {

			deps := getPushDependencies()

			domain := models.Domain{}
			domain.Name = "example.com"
			domain.Guid = "domain-guid"
			domain.Shared = true

			existingRoute := models.RouteSummary{}
			existingRoute.Host = "existing-app"
			existingRoute.Domain = domain.DomainFields

			existingApp := models.Application{}
			existingApp.Name = "existing-app"
			existingApp.Guid = "existing-app-guid"
			existingApp.Routes = []models.RouteSummary{existingRoute}

			deps.appRepo.ReadApp = existingApp
			deps.appRepo.UpdateAppResult = existingApp
			deps.routeRepo.FindByHostAndDomainNotFound = true
			deps.domainRepo.ListSharedDomainsDomains = []models.Domain{domain}

			ui := callPush(mr.T(), []string{"--no-hostname", "existing-app"}, deps)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Creating route", "example.com"},
				{"OK"},
				{"Binding", "example.com"},
			})
			testassert.SliceDoesNotContain(mr.T(), ui.Outputs, testassert.Lines{
				{"existing-app.example.com"},
			})

			assert.Equal(mr.T(), deps.routeRepo.FindByHostAndDomainDomain, "example.com")
			assert.Equal(mr.T(), deps.routeRepo.FindByHostAndDomainHost, "")
			assert.Equal(mr.T(), deps.routeRepo.CreatedHost, "")
			assert.Equal(mr.T(), deps.routeRepo.CreatedDomainGuid, "domain-guid")
		})
		It("TestPushingAppWhenItAlreadyExistsWithoutARouteCreatesADefaultDomain", func() {

			deps := getPushDependencies()

			sharedDomain := models.Domain{}
			sharedDomain.Name = "foo.cf-app.com"
			sharedDomain.Shared = true
			sharedDomain.Guid = "foo-domain-guid"

			deps.routeRepo.FindByHostAndDomainErr = true
			deps.domainRepo.ListSharedDomainsDomains = []models.Domain{sharedDomain}
			deps.appRepo.ReadApp = maker.NewApp(maker.Overrides{"name": "existing-app", "guid": "existing-app-guid"})
			deps.appRepo.UpdateAppResult = deps.appRepo.ReadApp

			ui := callPush(mr.T(), []string{"-t", "111", "existing-app"}, deps)
			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Creating route", "existing-app.foo.cf-app.com"},
				{"OK"},
				{"Binding", "existing-app.foo.cf-app.com"},
				{"OK"},
				{"Uploading"},
				{"OK"},
			})

			assert.Equal(mr.T(), deps.routeRepo.FindByHostAndDomainDomain, "foo.cf-app.com")
			assert.Equal(mr.T(), deps.routeRepo.FindByHostAndDomainHost, "existing-app")

			assert.Equal(mr.T(), deps.routeRepo.CreatedHost, "existing-app")
			assert.Equal(mr.T(), deps.routeRepo.CreatedDomainGuid, "foo-domain-guid")

			assert.Equal(mr.T(), deps.routeRepo.BoundAppGuid, "existing-app-guid")
			assert.Equal(mr.T(), deps.routeRepo.BoundRouteGuid, "existing-app-route-guid")
		})
		It("TestPushingAppWithInvalidPath", func() {

			deps := getPushDependencies()
			deps.appBitsRepo.UploadAppErr = true

			ui := callPush(mr.T(), []string{"app"}, deps)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Uploading"},
				{"FAILED"},
			})
		})
		It("TestPushingAppDescribesUpload", func() {

			deps := getPushDependencies()

			deps.appRepo.ReadNotFound = true
			deps.appBitsRepo.CallbackPath = "path/to/app"
			deps.appBitsRepo.CallbackZipSize = 61 * 1024 * 1024
			deps.appBitsRepo.CallbackFileCount = 11

			ui := callPush(mr.T(), []string{"appName"}, deps)
			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Uploading", "path/to/app"},
				{"61M", "11 files"},
			})
		})
		It("TestPushingWithNoManifestAndNoName", func() {

			deps := getPushDependencies()

			ui := callPush(mr.T(), []string{}, deps)
			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Incorrect Usage"},
			})
		})
	})
}
