package application_test

import (
	. "cf/commands/application"
	"cf/manifest"
	"cf/models"
	"cf/net"
	"errors"
	"generic"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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

var _ = Describe("Push Command", func() {
	It("TestPushingRequirements", func() {
		ui := new(testterm.FakeUI)
		configRepo := testconfig.NewRepository()
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

		cmd := NewPush(ui, configRepo, manifestRepo, starter, stopper, binder, appRepo, domainRepo, routeRepo, stackRepo, serviceRepo, appBitsRepo)
		ctxt := testcmd.NewContext("push", []string{})

		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}
		testcmd.RunCommand(cmd, ctxt, reqFactory)
		Expect(testcmd.CommandDidPassRequirements).To(BeTrue())

		reqFactory = &testreq.FakeReqFactory{LoginSuccess: false, TargetedSpaceSuccess: true}
		testcmd.RunCommand(cmd, ctxt, reqFactory)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())

		testcmd.CommandDidPassRequirements = true

		reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: false}
		testcmd.RunCommand(cmd, ctxt, reqFactory)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
	})

	It("TestPushingAppWithOldV2DomainsEndpoint", func() {
		deps := getPushDependencies()

		privateDomain := models.DomainFields{}
		privateDomain.Shared = false
		privateDomain.Name = "private.cf-app.com"
		privateDomain.Guid = "private-domain-guid"

		sharedDomain := models.DomainFields{}
		sharedDomain.Name = "shared.cf-app.com"
		sharedDomain.Shared = true
		sharedDomain.Guid = "shared-domain-guid"

		deps.domainRepo.ListSharedDomainsApiResponse = net.NewNotFoundApiResponse("whoopsie")
		deps.domainRepo.ListDomainsDomains = []models.DomainFields{privateDomain, sharedDomain}
		deps.routeRepo.FindByHostAndDomainErr = true
		deps.appRepo.ReadNotFound = true

		ui := callPush(mr.T(), []string{"-t", "111", "my-new-app"}, deps)

		Expect(deps.routeRepo.FindByHostAndDomainHost).To(Equal("my-new-app"))
		Expect(deps.routeRepo.CreatedHost).To(Equal("my-new-app"))
		Expect(deps.routeRepo.CreatedDomainGuid).To(Equal("shared-domain-guid"))
		Expect(deps.routeRepo.BoundAppGuid).To(Equal("my-new-app-guid"))
		Expect(deps.routeRepo.BoundRouteGuid).To(Equal("my-new-app-route-guid"))

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
		Expect(*deps.appRepo.CreatedAppParams().Name).To(Equal("my-new-app"))
		Expect(*deps.appRepo.CreatedAppParams().SpaceGuid).To(Equal("my-space-guid"))

		Expect(deps.routeRepo.FindByHostAndDomainHost).To(Equal("my-new-app"))
		Expect(deps.routeRepo.CreatedHost).To(Equal("my-new-app"))
		Expect(deps.routeRepo.CreatedDomainGuid).To(Equal("foo-domain-guid"))
		Expect(deps.routeRepo.BoundAppGuid).To(Equal("my-new-app-guid"))
		Expect(deps.routeRepo.BoundRouteGuid).To(Equal("my-new-app-route-guid"))

		Expect(deps.appBitsRepo.UploadedAppGuid).To(Equal("my-new-app-guid"))

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

		Expect(deps.stopper.AppToStop.Guid).To(Equal(""))
		Expect(deps.starter.AppToStart.Guid).To(Equal("my-new-app-guid"))
		Expect(deps.starter.AppToStart.Name).To(Equal("my-new-app"))
		Expect(deps.starter.Timeout).To(Equal(111))
	})

	It("TestPushingAppWithACrazyName", func() {
		deps := getPushDependencies()

		deps.routeRepo.FindByHostAndDomainErr = true
		deps.appRepo.ReadNotFound = true

		ui := callPush(mr.T(), []string{"-t", "111", "Tim's 1st-Crazy__app!"}, deps)
		Expect(*deps.appRepo.CreatedAppParams().Name).To(Equal("Tim's 1st-Crazy__app!"))

		Expect(deps.routeRepo.FindByHostAndDomainHost).To(Equal("tims-1st-crazy-app"))
		Expect(deps.routeRepo.CreatedHost).To(Equal("tims-1st-crazy-app"))

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
		route.Domain = deps.domainRepo.ListSharedDomainsDomains[0]

		deps.routeRepo.FindByHostAndDomainRoute = route
		deps.appRepo.ReadNotFound = true

		ui := callPush(mr.T(), []string{"my-new-app"}, deps)

		Expect(deps.routeRepo.CreatedHost).To(BeEmpty())
		Expect(deps.routeRepo.CreatedDomainGuid).To(BeEmpty())
		Expect(deps.routeRepo.FindByHostAndDomainHost).To(Equal("my-new-app"))
		Expect(deps.routeRepo.BoundAppGuid).To(Equal("my-new-app-guid"))
		Expect(deps.routeRepo.BoundRouteGuid).To(Equal("my-route-guid"))

		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"Using", "my-new-app.foo.cf-app.com"},
			{"Binding", "my-new-app.foo.cf-app.com"},
			{"OK"},
		})
	})

	It("TestPushingAppWithCustomFlags", func() {
		deps := getPushDependencies()
		deps.domainRepo.FindByNameInOrgDomain = models.DomainFields{
			Name: "bar.cf-app.com",
			Guid: "bar-domain-guid",
		}
		deps.routeRepo.FindByHostAndDomainErr = true
		deps.stackRepo.FindByNameStack = models.Stack{
			Name: "customLinux",
			Guid: "custom-linux-guid",
		}
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

		Expect(deps.stackRepo.FindByNameName).To(Equal("customLinux"))

		Expect(*deps.appRepo.CreatedAppParams().Name).To(Equal("my-new-app"))
		Expect(*deps.appRepo.CreatedAppParams().Command).To(Equal("unicorn -c config/unicorn.rb -D"))
		Expect(*deps.appRepo.CreatedAppParams().InstanceCount).To(Equal(3))
		Expect(*deps.appRepo.CreatedAppParams().Memory).To(Equal(uint64(2048)))
		Expect(*deps.appRepo.CreatedAppParams().StackGuid).To(Equal("custom-linux-guid"))
		Expect(*deps.appRepo.CreatedAppParams().HealthCheckTimeout).To(Equal(1))
		Expect(*deps.appRepo.CreatedAppParams().BuildpackUrl).To(Equal("https://github.com/heroku/heroku-buildpack-play.git"))

		Expect(deps.domainRepo.FindByNameInOrgName).To(Equal("bar.cf-app.com"))
		Expect(deps.domainRepo.FindByNameInOrgGuid).To(Equal("my-org-guid"))

		Expect(deps.routeRepo.CreatedHost).To(Equal("my-hostname"))
		Expect(deps.routeRepo.CreatedDomainGuid).To(Equal("bar-domain-guid"))
		Expect(deps.routeRepo.BoundAppGuid).To(Equal("my-new-app-guid"))
		Expect(deps.routeRepo.BoundRouteGuid).To(Equal("my-hostname-route-guid"))

		Expect(deps.appBitsRepo.UploadedAppGuid).To(Equal("my-new-app-guid"))

		Expect(deps.starter.AppToStart.Name).To(Equal(""))
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

		Expect(*deps.appRepo.UpdateParams.Command).To(Equal(""))
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

		deps.manifestRepo.ReadManifestReturns.Manifest = singleAppManifest()

		_ = callPush(mr.T(), []string{"existing-app"}, deps)

		updatedAppEnvVars := *deps.appRepo.UpdateParams.EnvironmentVars
		Expect(updatedAppEnvVars["crazy"]).To(Equal("pants"))
		Expect(updatedAppEnvVars["FOO"]).To(Equal("baz"))
		Expect(updatedAppEnvVars["foo"]).To(Equal("manchu"))
		Expect(updatedAppEnvVars["PATH"]).To(Equal("/u/apps/my-app/bin"))
	})

	It("TestPushingAppWithSingleAppManifest", func() {
		deps := getPushDependencies()
		domain := models.DomainFields{}
		domain.Name = "manifest-example.com"
		domain.Guid = "bar-domain-guid"
		deps.domainRepo.FindByNameInOrgDomain = domain
		deps.routeRepo.FindByHostAndDomainErr = true
		deps.appRepo.ReadNotFound = true

		deps.manifestRepo.ReadManifestReturns.Manifest = singleAppManifest()

		ui := callPush(mr.T(), []string{}, deps)
		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"Creating route", "manifest-host.manifest-example.com"},
			{"OK"},
			{"Binding", "manifest-host.manifest-example.com"},
			{"manifest-app-name"},
		})

		Expect(*deps.appRepo.CreatedAppParams().Name).To(Equal("manifest-app-name"))
		Expect(*deps.appRepo.CreatedAppParams().Memory).To(Equal(uint64(128)))
		Expect(*deps.appRepo.CreatedAppParams().InstanceCount).To(Equal(1))
		Expect(*deps.appRepo.CreatedAppParams().StackName).To(Equal("custom-stack"))
		Expect(*deps.appRepo.CreatedAppParams().BuildpackUrl).To(Equal("some-buildpack"))
		Expect(*deps.appRepo.CreatedAppParams().Command).To(Equal("JAVA_HOME=$PWD/.openjdk JAVA_OPTS=\"-Xss995K\" ./bin/start.sh run"))
		Expect(deps.appBitsRepo.UploadedDir).To(Equal("/some/path/from/manifest"))

		envVars := *deps.appRepo.CreatedAppParams().EnvironmentVars
		Expect(envVars).To(Equal(map[string]string{
			"PATH": "/u/apps/my-app/bin",
			"FOO":  "baz",
		}))
	})

	It("TestPushingAppManifestWithErrors", func() {
		deps := getPushDependencies()
		deps.appRepo.ReadNotFound = true
		deps.manifestRepo.ReadManifestReturns.Path = "/some-path/"
		deps.manifestRepo.ReadManifestReturns.Errors = manifest.ManifestErrors{
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
		deps.manifestRepo.ReadManifestReturns.Manifest = manifestWithServicesAndEnv()

		ui := callPush(mr.T(), []string{}, deps)

		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"Creating", "app1"},
			{"Creating", "app2"},
		})
		Expect(len(deps.appRepo.CreateAppParams)).To(Equal(2))

		firstApp := deps.appRepo.CreateAppParams[0]
		secondApp := deps.appRepo.CreateAppParams[1]
		Expect(*firstApp.Name).To(Equal("app1"))
		Expect(*secondApp.Name).To(Equal("app2"))

		envVars := *firstApp.EnvironmentVars
		Expect(envVars["SOMETHING"]).To(Equal("definitely-something"))

		envVars = *secondApp.EnvironmentVars
		Expect(envVars["SOMETHING"]).To(Equal("nothing"))
	})

	It("TestPushingASingleAppFromAManifestWithManyApps", func() {
		deps := getPushDependencies()
		deps.appRepo.ReadNotFound = true
		deps.manifestRepo.ReadManifestReturns.Manifest = manifestWithServicesAndEnv()

		ui := callPush(mr.T(), []string{"app2"}, deps)
		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"Creating", "app2"},
		})
		testassert.SliceDoesNotContain(mr.T(), ui.Outputs, testassert.Lines{
			{"Creating", "app1"},
		})
		Expect(len(deps.appRepo.CreateAppParams)).To(Equal(1))
		Expect(*deps.appRepo.CreateAppParams[0].Name).To(Equal("app2"))
	})

	It("TestNamedAppInAManifestIsNotFound", func() {
		deps := getPushDependencies()
		deps.appRepo.ReadNotFound = true
		deps.manifestRepo.ReadManifestReturns.Manifest = manifestWithServicesAndEnv()

		ui := callPush(mr.T(), []string{"non-existant-app"}, deps)
		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"Failed"},
		})
		Expect(len(deps.appRepo.CreateAppParams)).To(Equal(0))
	})

	It("TestPushingWithBindingMergedServices", func() {
		deps := getPushDependencies()
		deps.appRepo.ReadNotFound = true

		deps.serviceRepo.FindInstanceByNameMap = generic.NewMap(map[interface{}]interface{}{
			"global-service": maker.NewServiceInstance("global-service"),
			"app1-service":   maker.NewServiceInstance("app1-service"),
			"app2-service":   maker.NewServiceInstance("app2-service"),
		})

		deps.manifestRepo.ReadManifestReturns.Manifest = manifestWithServicesAndEnv()

		ui := callPush(mr.T(), []string{}, deps)
		Expect(len(deps.binder.AppsToBind)).To(Equal(4))
		Expect(deps.binder.AppsToBind[0].Name).To(Equal("app1"))
		Expect(deps.binder.AppsToBind[1].Name).To(Equal("app1"))
		Expect(deps.binder.InstancesToBindTo[0].Name).To(Equal("app1-service"))
		Expect(deps.binder.InstancesToBindTo[1].Name).To(Equal("global-service"))

		Expect(deps.binder.AppsToBind[2].Name).To(Equal("app2"))
		Expect(deps.binder.AppsToBind[3].Name).To(Equal("app2"))
		Expect(deps.binder.InstancesToBindTo[2].Name).To(Equal("app2-service"))
		Expect(deps.binder.InstancesToBindTo[3].Name).To(Equal("global-service"))

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
		deps.manifestRepo.ReadManifestReturns.Manifest = manifestWithServicesAndEnv()

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
		Expect(err).NotTo(HaveOccurred())

		callPush(mr.T(), []string{
			"-p", absPath,
			"app-with-path",
		}, deps)

		Expect(deps.appBitsRepo.UploadedDir).To(Equal(absPath))
	})

	It("TestPushingAppWithPathToZipFile", func() {
		deps := getPushDependencies()
		deps.appRepo.ReadNotFound = true

		absPath, err := filepath.Abs("../../../fixtures/example-app.jar")
		Expect(err).NotTo(HaveOccurred())

		callPush(mr.T(), []string{
			"-p", absPath,
			"app-with-path",
		}, deps)

		Expect(deps.appBitsRepo.UploadedDir).To(Equal(absPath))
	})

	It("TestPushingWithDefaultAppPath", func() {
		deps := getPushDependencies()
		deps.appRepo.ReadNotFound = true

		callPush(mr.T(), []string{"app-with-default-path"}, deps)

		dir, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		Expect(deps.appBitsRepo.UploadedDir).To(Equal(dir))
	})

	It("TestPushingWithRelativeAppPath", func() {
		deps := getPushDependencies()
		deps.appRepo.ReadNotFound = true

		callPush(mr.T(), []string{
			"-p", "../../../fixtures/example-app",
			"app-with-relative-path",
		}, deps)

		dir, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		Expect(deps.appBitsRepo.UploadedDir).To(Equal(filepath.Join(dir, "../../../fixtures/example-app")))
	})

	It("TestPushingWithBadManifestPath", func() {
		deps := getPushDependencies()
		deps.appRepo.ReadNotFound = true

		deps.manifestRepo.ReadManifestReturns.Manifest = manifest.NewEmptyManifest()
		deps.manifestRepo.ReadManifestReturns.Errors = []error{errors.New("read manifest error")}

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
		deps.manifestRepo.ReadManifestReturns.Manifest = singleAppManifest()
		deps.manifestRepo.ReadManifestReturns.Errors = manifest.ManifestErrors{syscall.ENOENT}
		deps.manifestRepo.ReadManifestReturns.Path = ""

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
		deps.manifestRepo.ReadManifestReturns.Manifest = singleAppManifest()
		deps.manifestRepo.ReadManifestReturns.Path = "manifest.yml"

		ui := callPush(mr.T(), []string{"-p", "some/relative/path"}, deps)
		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"Using manifest file", "manifest.yml"},
		})

		cwd, _ := os.Getwd()
		Expect(deps.manifestRepo.ReadManifestArgs.Path).To(Equal(cwd))
	})

	It("TestPushingWithNoManifestFlag", func() {
		deps := getPushDependencies()
		deps.appRepo.ReadNotFound = true

		ui := callPush(mr.T(), []string{"--no-route", "--no-manifest", "app-name"}, deps)

		testassert.SliceDoesNotContain(mr.T(), ui.Outputs, testassert.Lines{
			{"FAILED"},
			{"hacker-manifesto"},
		})

		Expect(deps.manifestRepo.ReadManifestArgs.Path).To(Equal(""))
		Expect(*deps.appRepo.CreatedAppParams().Name).To(Equal("app-name"))
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
		domain := models.DomainFields{}
		domain.Name = "bar.cf-app.com"
		domain.Guid = "bar-domain-guid"

		deps.domainRepo.FindByNameInOrgDomain = domain
		deps.routeRepo.FindByHostErr = true
		deps.appRepo.ReadNotFound = true

		callPush(mr.T(), []string{
			"--no-route",
			"my-new-app",
		}, deps)

		Expect(*deps.appRepo.CreatedAppParams().Name).To(Equal("my-new-app"))
		Expect(deps.routeRepo.CreatedHost).To(Equal(""))
		Expect(deps.routeRepo.CreatedDomainGuid).To(Equal(""))
	})

	It("TestPushingAppWithNoHostname", func() {
		deps := getPushDependencies()
		domain := models.DomainFields{}
		domain.Name = "bar.cf-app.com"
		domain.Guid = "bar-domain-guid"
		domain.Shared = true

		deps.domainRepo.ListSharedDomainsDomains = []models.DomainFields{domain}
		deps.routeRepo.FindByHostAndDomainErr = true
		deps.appRepo.ReadNotFound = true

		callPush(mr.T(), []string{
			"--no-hostname",
			"my-new-app",
		}, deps)

		Expect(*deps.appRepo.CreatedAppParams().Name).To(Equal("my-new-app"))
		Expect(deps.routeRepo.CreatedHost).To(Equal(""))
		Expect(deps.routeRepo.CreatedDomainGuid).To(Equal("bar-domain-guid"))
	})

	It("TestPushingAppAsWorker", func() {
		deps := getPushDependencies()
		deps.appRepo.ReadNotFound = true

		workerManifest := singleAppManifest()
		noRoute := true
		workerManifest.Applications[0].NoRoute = &noRoute
		deps.manifestRepo.ReadManifestReturns.Manifest = workerManifest

		ui := callPush(mr.T(), []string{
			"worker-app",
		}, deps)

		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"worker-app", "is a worker", "skipping route creation"},
		})
		Expect(deps.routeRepo.BoundAppGuid).To(Equal(""))
		Expect(deps.routeRepo.BoundRouteGuid).To(Equal(""))
	})

	It("TestPushingAppWithMemoryInMegaBytes", func() {
		deps := getPushDependencies()
		deps.appRepo.ReadNotFound = true

		callPush(mr.T(), []string{
			"-m", "256M",
			"my-new-app",
		}, deps)

		Expect(*deps.appRepo.CreatedAppParams().Memory).To(Equal(uint64(256)))
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

		Expect(deps.stopper.AppToStop.Guid).To(Equal(existingApp.Guid))
		Expect(deps.appBitsRepo.UploadedAppGuid).To(Equal(existingApp.Guid))
	})

	It("TestPushingAppWhenItIsStopped", func() {
		deps := getPushDependencies()
		stoppedApp := maker.NewApp(maker.Overrides{"state": "stopped", "name": "stopped-app"})

		deps.appRepo.ReadApp = stoppedApp
		deps.appRepo.UpdateAppResult = stoppedApp

		_ = callPush(mr.T(), []string{"stopped-app"}, deps)

		Expect(deps.stopper.AppToStop.Guid).To(Equal(""))
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

		Expect(*deps.appRepo.UpdateParams.Command).To(Equal("different start command"))
		Expect(*deps.appRepo.UpdateParams.InstanceCount).To(Equal(10))
		Expect(*deps.appRepo.UpdateParams.Memory).To(Equal(uint64(1024)))
		Expect(*deps.appRepo.UpdateParams.BuildpackUrl).To(Equal("https://github.com/heroku/heroku-buildpack-different.git"))
		Expect(*deps.appRepo.UpdateParams.StackGuid).To(Equal("differentStack-guid"))
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
		Expect(deps.appBitsRepo.UploadedAppGuid).To(Equal("existing-app-guid"))
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

		foundDomain := models.DomainFields{}
		foundDomain.Guid = "domain-guid"
		foundDomain.Name = "newdomain.com"

		deps.appRepo.ReadApp = existingApp
		deps.appRepo.UpdateAppResult = existingApp
		deps.routeRepo.FindByHostAndDomainNotFound = true
		deps.domainRepo.FindByNameInOrgDomain = foundDomain

		ui := callPush(mr.T(), []string{"-d", "newdomain.com", "existing-app"}, deps)

		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"Creating route", "existing-app.newdomain.com"},
			{"OK"},
			{"Binding", "existing-app.newdomain.com"},
		})

		Expect(deps.appBitsRepo.UploadedAppGuid).To(Equal("existing-app-guid"))
		Expect(deps.domainRepo.FindByNameInOrgName).To(Equal("newdomain.com"))
		Expect(deps.domainRepo.FindByNameInOrgGuid).To(Equal("my-org-guid"))
		Expect(deps.routeRepo.FindByHostAndDomainDomain).To(Equal("newdomain.com"))
		Expect(deps.routeRepo.FindByHostAndDomainHost).To(Equal("existing-app"))
		Expect(deps.routeRepo.CreatedHost).To(Equal("existing-app"))
		Expect(deps.routeRepo.CreatedDomainGuid).To(Equal("domain-guid"))
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

		Expect(deps.appBitsRepo.UploadedAppGuid).To(Equal("existing-app-guid"))
		Expect(deps.domainRepo.FindByNameInOrgName).To(Equal(""))
		Expect(deps.routeRepo.FindByHostAndDomainDomain).To(Equal(""))
		Expect(deps.routeRepo.FindByHostAndDomainHost).To(Equal(""))
		Expect(deps.routeRepo.CreatedHost).To(Equal(""))
		Expect(deps.routeRepo.CreatedDomainGuid).To(Equal(""))
	})

	It("TestPushingAppWhenItAlreadyExistsAndHostIsSpecified", func() {
		deps := getPushDependencies()

		domain := models.DomainFields{}
		domain.Name = "example.com"
		domain.Guid = "domain-guid"
		domain.Shared = true

		existingRoute := models.RouteSummary{}
		existingRoute.Host = "existing-app"
		existingRoute.Domain = domain

		existingApp := models.Application{}
		existingApp.Name = "existing-app"
		existingApp.Guid = "existing-app-guid"
		existingApp.Routes = []models.RouteSummary{existingRoute}

		deps.appRepo.ReadApp = existingApp
		deps.appRepo.UpdateAppResult = existingApp
		deps.routeRepo.FindByHostAndDomainNotFound = true
		deps.domainRepo.ListSharedDomainsDomains = []models.DomainFields{domain}

		ui := callPush(mr.T(), []string{"-n", "new-host", "existing-app"}, deps)

		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"Creating route", "new-host.example.com"},
			{"OK"},
			{"Binding", "new-host.example.com"},
		})

		Expect(deps.routeRepo.FindByHostAndDomainDomain).To(Equal("example.com"))
		Expect(deps.routeRepo.FindByHostAndDomainHost).To(Equal("new-host"))
		Expect(deps.routeRepo.CreatedHost).To(Equal("new-host"))
		Expect(deps.routeRepo.CreatedDomainGuid).To(Equal("domain-guid"))
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

		Expect(deps.appBitsRepo.UploadedAppGuid).To(Equal("existing-app-guid"))
		Expect(deps.domainRepo.FindByNameInOrgName).To(Equal(""))
		Expect(deps.routeRepo.FindByHostAndDomainDomain).To(Equal(""))
		Expect(deps.routeRepo.FindByHostAndDomainHost).To(Equal(""))
		Expect(deps.routeRepo.CreatedHost).To(Equal(""))
		Expect(deps.routeRepo.CreatedDomainGuid).To(Equal(""))
	})

	It("TestPushingAppWhenItAlreadyExistsAndNoHostFlagIsPresent", func() {
		deps := getPushDependencies()

		domain := models.DomainFields{}
		domain.Name = "example.com"
		domain.Guid = "domain-guid"
		domain.Shared = true

		existingRoute := models.RouteSummary{}
		existingRoute.Host = "existing-app"
		existingRoute.Domain = domain

		existingApp := models.Application{}
		existingApp.Name = "existing-app"
		existingApp.Guid = "existing-app-guid"
		existingApp.Routes = []models.RouteSummary{existingRoute}

		deps.appRepo.ReadApp = existingApp
		deps.appRepo.UpdateAppResult = existingApp
		deps.routeRepo.FindByHostAndDomainNotFound = true
		deps.domainRepo.ListSharedDomainsDomains = []models.DomainFields{domain}

		ui := callPush(mr.T(), []string{"--no-hostname", "existing-app"}, deps)

		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"Creating route", "example.com"},
			{"OK"},
			{"Binding", "example.com"},
		})
		testassert.SliceDoesNotContain(mr.T(), ui.Outputs, testassert.Lines{
			{"existing-app.example.com"},
		})

		Expect(deps.routeRepo.FindByHostAndDomainDomain).To(Equal("example.com"))
		Expect(deps.routeRepo.FindByHostAndDomainHost).To(Equal(""))
		Expect(deps.routeRepo.CreatedHost).To(Equal(""))
		Expect(deps.routeRepo.CreatedDomainGuid).To(Equal("domain-guid"))
	})

	It("TestPushingAppWhenItAlreadyExistsWithoutARouteCreatesADefaultDomain", func() {
		deps := getPushDependencies()

		sharedDomain := models.DomainFields{}
		sharedDomain.Name = "foo.cf-app.com"
		sharedDomain.Shared = true
		sharedDomain.Guid = "foo-domain-guid"

		deps.routeRepo.FindByHostAndDomainErr = true
		deps.domainRepo.ListSharedDomainsDomains = []models.DomainFields{sharedDomain}
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

		Expect(deps.routeRepo.FindByHostAndDomainDomain).To(Equal("foo.cf-app.com"))
		Expect(deps.routeRepo.FindByHostAndDomainHost).To(Equal("existing-app"))

		Expect(deps.routeRepo.CreatedHost).To(Equal("existing-app"))
		Expect(deps.routeRepo.CreatedDomainGuid).To(Equal("foo-domain-guid"))

		Expect(deps.routeRepo.BoundAppGuid).To(Equal("existing-app-guid"))
		Expect(deps.routeRepo.BoundRouteGuid).To(Equal("existing-app-route-guid"))
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
	sharedDomain := maker.NewSharedDomainFields(maker.Overrides{"name": "foo.cf-app.com", "guid": "foo-domain-guid"})
	deps.domainRepo.ListSharedDomainsDomains = []models.DomainFields{sharedDomain}

	deps.routeRepo = &testapi.FakeRouteRepository{}
	deps.stackRepo = &testapi.FakeStackRepository{}
	deps.appBitsRepo = &testapi.FakeApplicationBitsRepository{}
	deps.serviceRepo = &testapi.FakeServiceRepo{}

	return
}

func callPush(t mr.TestingT, args []string, deps pushDependencies) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("push", args)

	configRepo := testconfig.NewRepositoryWithDefaults()

	cmd := NewPush(ui, configRepo, deps.manifestRepo, deps.starter,
		deps.stopper, deps.binder, deps.appRepo, deps.domainRepo,
		deps.routeRepo, deps.stackRepo, deps.serviceRepo, deps.appBitsRepo)

	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}
	testcmd.RunCommand(cmd, ctxt, reqFactory)

	return
}
