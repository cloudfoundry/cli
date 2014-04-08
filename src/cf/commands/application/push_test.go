package application_test

import (
	. "cf/commands/application"
	"cf/configuration"
	"cf/errors"
	"cf/manifest"
	"cf/models"
	"generic"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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
	testwords "testhelpers/words"
	"words"
)

var _ = Describe("Push Command", func() {
	var (
		cmd           *Push
		ui            *testterm.FakeUI
		configRepo    configuration.ReadWriter
		manifestRepo  *testmanifest.FakeManifestRepository
		starter       *testcmd.FakeAppStarter
		stopper       *testcmd.FakeAppStopper
		binder        *testcmd.FakeAppBinder
		appRepo       *testapi.FakeApplicationRepository
		domainRepo    *testapi.FakeDomainRepository
		routeRepo     *testapi.FakeRouteRepository
		stackRepo     *testapi.FakeStackRepository
		appBitsRepo   *testapi.FakeApplicationBitsRepository
		serviceRepo   *testapi.FakeServiceRepo
		wordGenerator words.WordGenerator
	)

	BeforeEach(func() {
		manifestRepo = &testmanifest.FakeManifestRepository{}
		starter = &testcmd.FakeAppStarter{}
		stopper = &testcmd.FakeAppStopper{}
		binder = &testcmd.FakeAppBinder{}
		appRepo = &testapi.FakeApplicationRepository{}

		domainRepo = &testapi.FakeDomainRepository{}
		sharedDomain := maker.NewSharedDomainFields(maker.Overrides{"name": "foo.cf-app.com", "guid": "foo-domain-guid"})
		domainRepo.ListDomainsForOrgDomains = []models.DomainFields{sharedDomain}

		routeRepo = &testapi.FakeRouteRepository{}
		stackRepo = &testapi.FakeStackRepository{}
		appBitsRepo = &testapi.FakeApplicationBitsRepository{CallbackZipSize: 1, CallbackFileCount: 1}
		serviceRepo = &testapi.FakeServiceRepo{}
		wordGenerator = testwords.NewFakeWordGenerator("laughing-cow")

		ui = new(testterm.FakeUI)
		configRepo = testconfig.NewRepositoryWithDefaults()

		cmd = NewPush(ui, configRepo, manifestRepo, starter, stopper, binder, appRepo, domainRepo, routeRepo, stackRepo, serviceRepo, appBitsRepo, wordGenerator)
	})

	callPush := func(args ...string) {
		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}
		testcmd.RunCommand(cmd, testcmd.NewContext("push", args), reqFactory)
	}

	Describe("requirements", func() {
		It("passes when logged in and a space is targeted", func() {
			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}
			testcmd.RunCommand(cmd, testcmd.NewContext("push", []string{}), reqFactory)
			Expect(testcmd.CommandDidPassRequirements).To(BeTrue())
		})

		It("fails when not logged in", func() {
			reqFactory := &testreq.FakeReqFactory{LoginSuccess: false, TargetedSpaceSuccess: true}
			testcmd.RunCommand(cmd, testcmd.NewContext("push", []string{}), reqFactory)
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})

		It("fails when a space is not targeted", func() {
			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: false}
			testcmd.RunCommand(cmd, testcmd.NewContext("push", []string{}), reqFactory)
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})

		// yes, we're aware that the args here should probably be provided in a different order
		// erg: app-name -p some/path some-extra-arg
		// but the test infrastructure for parsing args and flags is sorely lacking
		It("fails when provided too many args", func() {
			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}
			testcmd.RunCommand(cmd, testcmd.NewContext("push", []string{"-p", "path", "too-much", "app-name"}), reqFactory)
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})
	})

	Describe("when pushing a new app", func() {
		BeforeEach(func() {
			appRepo.ReadReturns.Error = errors.NewModelNotFoundError("App", "the-app")
		})

		It("creates an app", func() {
			routeRepo.FindByHostAndDomainErr = true

			callPush("-t", "111", "my-new-app")

			Expect(*appRepo.CreatedAppParams().Name).To(Equal("my-new-app"))
			Expect(*appRepo.CreatedAppParams().SpaceGuid).To(Equal("my-space-guid"))

			Expect(routeRepo.FindByHostAndDomainHost).To(Equal("my-new-app"))
			Expect(routeRepo.CreatedHost).To(Equal("my-new-app"))
			Expect(routeRepo.CreatedDomainGuid).To(Equal("foo-domain-guid"))
			Expect(routeRepo.BoundAppGuid).To(Equal("my-new-app-guid"))
			Expect(routeRepo.BoundRouteGuid).To(Equal("my-new-app-route-guid"))

			Expect(appBitsRepo.UploadedAppGuid).To(Equal("my-new-app-guid"))

			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"Creating app", "my-new-app", "my-org", "my-space"},
				{"OK"},
				{"Creating", "my-new-app.foo.cf-app.com"},
				{"OK"},
				{"Binding", "my-new-app.foo.cf-app.com"},
				{"OK"},
				{"Uploading my-new-app"},
				{"OK"},
			})

			Expect(stopper.AppToStop.Guid).To(Equal(""))
			Expect(starter.AppToStart.Guid).To(Equal("my-new-app-guid"))
			Expect(starter.AppToStart.Name).To(Equal("my-new-app"))
			Expect(starter.Timeout).To(Equal(111))
		})

		It("strips special characters when creating a default route", func() {
			routeRepo.FindByHostAndDomainErr = true

			callPush("-t", "111", "Tim's 1st-Crazy__app!")
			Expect(*appRepo.CreatedAppParams().Name).To(Equal("Tim's 1st-Crazy__app!"))

			Expect(routeRepo.FindByHostAndDomainHost).To(Equal("tims-1st-crazy-app"))
			Expect(routeRepo.CreatedHost).To(Equal("tims-1st-crazy-app"))

			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"Creating", "tims-1st-crazy-app.foo.cf-app.com"},
				{"Binding", "tims-1st-crazy-app.foo.cf-app.com"},
			})
			testassert.SliceDoesNotContain(ui.Outputs, testassert.Lines{
				{"FAILED"},
			})
		})

		It("binds to existing routes", func() {
			route := models.Route{}
			route.Guid = "my-route-guid"
			route.Host = "my-new-app"
			route.Domain = domainRepo.ListDomainsForOrgDomains[0]

			routeRepo.FindByHostAndDomainRoute = route

			callPush("my-new-app")

			Expect(routeRepo.CreatedHost).To(BeEmpty())
			Expect(routeRepo.CreatedDomainGuid).To(BeEmpty())
			Expect(routeRepo.FindByHostAndDomainHost).To(Equal("my-new-app"))
			Expect(routeRepo.BoundAppGuid).To(Equal("my-new-app-guid"))
			Expect(routeRepo.BoundRouteGuid).To(Equal("my-route-guid"))

			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"Using", "my-new-app.foo.cf-app.com"},
				{"Binding", "my-new-app.foo.cf-app.com"},
				{"OK"},
			})
		})

		It("sets the app params from the flags", func() {
			domainRepo.FindByNameInOrgDomain = models.DomainFields{
				Name: "bar.cf-app.com",
				Guid: "bar-domain-guid",
			}
			routeRepo.FindByHostAndDomainErr = true
			stackRepo.FindByNameStack = models.Stack{
				Name: "customLinux",
				Guid: "custom-linux-guid",
			}

			callPush(
				"-c", "unicorn -c config/unicorn.rb -D",
				"-d", "bar.cf-app.com",
				"-n", "my-hostname",
				"-k", "4G",
				"-i", "3",
				"-m", "2G",
				"-b", "https://github.com/heroku/heroku-buildpack-play.git",
				"-s", "customLinux",
				"-t", "1",
				"--no-start",
				"my-new-app",
			)

			testassert.SliceContains(ui.Outputs, testassert.Lines{
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

			Expect(stackRepo.FindByNameName).To(Equal("customLinux"))

			Expect(*appRepo.CreatedAppParams().Name).To(Equal("my-new-app"))
			Expect(*appRepo.CreatedAppParams().Command).To(Equal("unicorn -c config/unicorn.rb -D"))
			Expect(*appRepo.CreatedAppParams().InstanceCount).To(Equal(3))
			Expect(*appRepo.CreatedAppParams().DiskQuota).To(Equal(uint64(4096)))
			Expect(*appRepo.CreatedAppParams().Memory).To(Equal(uint64(2048)))
			Expect(*appRepo.CreatedAppParams().StackGuid).To(Equal("custom-linux-guid"))
			Expect(*appRepo.CreatedAppParams().HealthCheckTimeout).To(Equal(1))
			Expect(*appRepo.CreatedAppParams().BuildpackUrl).To(Equal("https://github.com/heroku/heroku-buildpack-play.git"))

			Expect(domainRepo.FindByNameInOrgName).To(Equal("bar.cf-app.com"))
			Expect(domainRepo.FindByNameInOrgGuid).To(Equal("my-org-guid"))

			Expect(routeRepo.CreatedHost).To(Equal("my-hostname"))
			Expect(routeRepo.CreatedDomainGuid).To(Equal("bar-domain-guid"))
			Expect(routeRepo.BoundAppGuid).To(Equal("my-new-app-guid"))
			Expect(routeRepo.BoundRouteGuid).To(Equal("my-hostname-route-guid"))

			Expect(appBitsRepo.UploadedAppGuid).To(Equal("my-new-app-guid"))

			Expect(starter.AppToStart.Name).To(Equal(""))
		})

		Context("when there is a shared domain", func() {
			It("creates a route with the shared domain and maps it to the app", func() {
				privateDomain := models.DomainFields{
					Shared: false,
					Name:   "private.cf-app.com",
					Guid:   "private-domain-guid",
				}
				sharedDomain := models.DomainFields{
					Name:   "shared.cf-app.com",
					Shared: true,
					Guid:   "shared-domain-guid",
				}

				domainRepo.ListDomainsForOrgDomains = []models.DomainFields{privateDomain, sharedDomain}
				routeRepo.FindByHostAndDomainErr = true

				callPush("-t", "111", "my-new-app")

				Expect(routeRepo.FindByHostAndDomainHost).To(Equal("my-new-app"))
				Expect(routeRepo.CreatedHost).To(Equal("my-new-app"))
				Expect(routeRepo.CreatedDomainGuid).To(Equal("shared-domain-guid"))
				Expect(routeRepo.BoundAppGuid).To(Equal("my-new-app-guid"))
				Expect(routeRepo.BoundRouteGuid).To(Equal("my-new-app-route-guid"))

				testassert.SliceContains(ui.Outputs, testassert.Lines{
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
		})

		Context("when there is no shared domain but there is a private domain in the targeted org", func() {
			It("creates a route with the private domain and maps it to the app", func() {
				privateDomain := models.DomainFields{
					Shared: false,
					Name:   "private.cf-app.com",
					Guid:   "private-domain-guid",
				}

				domainRepo.ListDomainsForOrgDomains = []models.DomainFields{privateDomain}
				routeRepo.FindByHostAndDomainErr = true
				appRepo.ReadReturns.Error = errors.NewModelNotFoundError("App", "the-app")

				callPush("-t", "111", "my-new-app")

				Expect(routeRepo.FindByHostAndDomainHost).To(Equal("my-new-app"))
				Expect(routeRepo.CreatedHost).To(Equal("my-new-app"))
				Expect(routeRepo.CreatedDomainGuid).To(Equal("private-domain-guid"))
				Expect(routeRepo.BoundAppGuid).To(Equal("my-new-app-guid"))
				Expect(routeRepo.BoundRouteGuid).To(Equal("my-new-app-route-guid"))

				testassert.SliceContains(ui.Outputs, testassert.Lines{
					{"Creating app", "my-new-app", "my-org", "my-space"},
					{"OK"},
					{"Creating", "my-new-app.private.cf-app.com"},
					{"OK"},
					{"Binding", "my-new-app.private.cf-app.com"},
					{"OK"},
					{"Uploading my-new-app"},
					{"OK"},
				})
			})
		})

		Describe("randomized hostnames", func() {
			var manifestApp generic.Map

			BeforeEach(func() {
				manifest := singleAppManifest()
				manifestApp = manifest.Data.Get("applications").([]interface{})[0].(generic.Map)
				manifestApp.Delete("host")
				manifestRepo.ReadManifestReturns.Manifest = manifest
			})

			It("provides a random hostname when the --random-route flag is passed", func() {
				callPush("--random-route", "my-new-app")
				Expect(routeRepo.FindByHostAndDomainHost).To(Equal("my-new-app-laughing-cow"))
			})

			It("provides a random hostname when the random-route option is set in the manifest", func() {
				manifestApp.Set("random-route", true)

				callPush("my-new-app")

				Expect(routeRepo.FindByHostAndDomainHost).To(Equal("my-new-app-laughing-cow"))
			})
		})

		It("pushes the contents of the directory specified using the -p flag", func() {
			callPush("-p", "../some/path-to/an-app", "app-with-path")
			Expect(appBitsRepo.UploadedDir).To(Equal("../some/path-to/an-app"))
		})

		It("pushes the contents of the current working directory by default", func() {
			callPush("app-with-default-path")
			dir, _ := os.Getwd()
			Expect(appBitsRepo.UploadedDir).To(Equal(dir))
		})

		It("fails when given a bad manifest path", func() {
			manifestRepo.ReadManifestReturns.Manifest = manifest.NewEmptyManifest()
			manifestRepo.ReadManifestReturns.Error = errors.New("read manifest error")

			callPush("-f", "bad/manifest/path")

			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"FAILED"},
				{"read manifest error"},
			})
		})

		It("does not fail when the current working directory does not contain a manifest", func() {
			manifestRepo.ReadManifestReturns.Manifest = singleAppManifest()
			manifestRepo.ReadManifestReturns.Error = syscall.ENOENT
			manifestRepo.ReadManifestReturns.Manifest.Path = ""

			callPush("--no-route", "app-name")

			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"Creating app", "app-name"},
				{"OK"},
				{"Uploading", "app-name"},
				{"OK"},
			})
			testassert.SliceDoesNotContain(ui.Outputs, testassert.Lines{
				{"FAILED"},
			})
		})

		It("uses the manifest in the current directory by default", func() {
			manifestRepo.ReadManifestReturns.Manifest = singleAppManifest()
			manifestRepo.ReadManifestReturns.Manifest.Path = "manifest.yml"

			callPush("-p", "some/relative/path")

			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"Using manifest file", "manifest.yml"},
			})

			cwd, _ := os.Getwd()
			Expect(manifestRepo.ReadManifestArgs.Path).To(Equal(cwd))
		})

		It("does not use a manifest if the 'no-manifest' flag is passed", func() {
			callPush("--no-route", "--no-manifest", "app-name")

			testassert.SliceDoesNotContain(ui.Outputs, testassert.Lines{
				{"FAILED"},
				{"hacker-manifesto"},
			})

			Expect(manifestRepo.ReadManifestArgs.Path).To(Equal(""))
			Expect(*appRepo.CreatedAppParams().Name).To(Equal("app-name"))
		})

		It("pushes an app when provided a manifest with one app defined", func() {
			domain := models.DomainFields{}
			domain.Name = "manifest-example.com"
			domain.Guid = "bar-domain-guid"
			domainRepo.FindByNameInOrgDomain = domain
			routeRepo.FindByHostAndDomainErr = true

			manifestRepo.ReadManifestReturns.Manifest = singleAppManifest()

			callPush()
			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"Creating route", "manifest-host.manifest-example.com"},
				{"OK"},
				{"Binding", "manifest-host.manifest-example.com"},
				{"manifest-app-name"},
			})

			Expect(*appRepo.CreatedAppParams().Name).To(Equal("manifest-app-name"))
			Expect(*appRepo.CreatedAppParams().Memory).To(Equal(uint64(128)))
			Expect(*appRepo.CreatedAppParams().InstanceCount).To(Equal(1))
			Expect(*appRepo.CreatedAppParams().StackName).To(Equal("custom-stack"))
			Expect(*appRepo.CreatedAppParams().BuildpackUrl).To(Equal("some-buildpack"))
			Expect(*appRepo.CreatedAppParams().Command).To(Equal("JAVA_HOME=$PWD/.openjdk JAVA_OPTS=\"-Xss995K\" ./bin/start.sh run"))
			Expect(appBitsRepo.UploadedDir).To(Equal(filepath.Clean("some/path/from/manifest")))

			envVars := *appRepo.CreatedAppParams().EnvironmentVars
			Expect(envVars).To(Equal(map[string]string{
				"PATH": "/u/apps/my-app/bin",
				"FOO":  "baz",
			}))
		})

		It("fails when parsing the manifest has errors", func() {
			manifestRepo.ReadManifestReturns.Manifest = &manifest.Manifest{Path: "/some-path/"}
			manifestRepo.ReadManifestReturns.Error = errors.New("buildpack should not be null")

			callPush()

			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"FAILED"},
				{"Error", "reading", "manifest"},
				{"buildpack should not be null"},
			})
		})

		It("does not create a route when provided the --no-route flag", func() {
			domain := models.DomainFields{}
			domain.Name = "bar.cf-app.com"
			domain.Guid = "bar-domain-guid"

			domainRepo.FindByNameInOrgDomain = domain
			routeRepo.FindByHostAndDomainErr = true

			callPush("--no-route", "my-new-app")

			Expect(*appRepo.CreatedAppParams().Name).To(Equal("my-new-app"))
			Expect(routeRepo.CreatedHost).To(Equal(""))
			Expect(routeRepo.CreatedDomainGuid).To(Equal(""))
		})

		It("maps the root domain route to the app when given the --no-hostname flag", func() {
			domain := models.DomainFields{}
			domain.Name = "bar.cf-app.com"
			domain.Guid = "bar-domain-guid"
			domain.Shared = true

			domainRepo.ListDomainsForOrgDomains = []models.DomainFields{domain}
			routeRepo.FindByHostAndDomainErr = true

			callPush("--no-hostname", "my-new-app")

			Expect(*appRepo.CreatedAppParams().Name).To(Equal("my-new-app"))
			Expect(routeRepo.CreatedHost).To(Equal(""))
			Expect(routeRepo.CreatedDomainGuid).To(Equal("bar-domain-guid"))
		})

		It("Does not create a route when the no-route property is in the manifest", func() {
			workerManifest := singleAppManifest()
			workerManifest.Data.Get("applications").([]interface{})[0].(generic.Map).Set("no-route", true)
			manifestRepo.ReadManifestReturns.Manifest = workerManifest

			callPush("worker-app")

			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"worker-app", "is a worker", "skipping route creation"},
			})
			Expect(routeRepo.BoundAppGuid).To(Equal(""))
			Expect(routeRepo.BoundRouteGuid).To(Equal(""))
		})

		It("fails when given an invalid memory limit", func() {
			callPush("-m", "abcM", "my-new-app")

			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"FAILED"},
				{"invalid", "memory"},
			})
		})
	})

	Describe("re-pushing an existing app", func() {
		BeforeEach(func() {
			existingApp := models.Application{}
			existingApp.Name = "existing-app"
			existingApp.Guid = "existing-app-guid"
			existingApp.Command = "unicorn -c config/unicorn.rb -D"
			existingApp.EnvironmentVars = map[string]string{
				"crazy": "pants",
				"FOO":   "NotYoBaz",
				"foo":   "manchu",
			}

			appRepo.ReadReturns.App = existingApp
		})

		// HERE

		It("resets the app's buildpack when the -b flag is provided as 'null'", func() {
			callPush("-b", "null", "existing-app")
			Expect(*appRepo.UpdateParams.BuildpackUrl).To(Equal(""))
		})

		It("resets the app's command when the -c flag is provided as 'null'", func() {
			callPush("-c", "null", "existing-app")
			Expect(*appRepo.UpdateParams.Command).To(Equal(""))
		})

		It("merges env vars from the manifest with those from the server", func() {
			manifestRepo.ReadManifestReturns.Manifest = singleAppManifest()

			callPush("existing-app")

			updatedAppEnvVars := *appRepo.UpdateParams.EnvironmentVars
			Expect(updatedAppEnvVars["crazy"]).To(Equal("pants"))
			Expect(updatedAppEnvVars["FOO"]).To(Equal("baz"))
			Expect(updatedAppEnvVars["foo"]).To(Equal("manchu"))
			Expect(updatedAppEnvVars["PATH"]).To(Equal("/u/apps/my-app/bin"))
		})
	})

	Describe("checking for bad flags", func() {
		It("fails when non-positive value is given for memory limit", func() {
			callPush("-m", "0G", "my-new-app")

			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"FAILED"},
				{"memory"},
				{"positive integer"},
			})
		})

		It("fails when non-positive value is given for instances", func() {
			callPush("-i", "0", "my-new-app")

			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"FAILED"},
				{"instance count"},
				{"positive integer"},
			})
		})

		It("fails when non-positive value is given for disk quota", func() {
			callPush("-k", "-1G", "my-new-app")

			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"FAILED"},
				{"disk quota"},
				{"positive integer"},
			})
		})

		It("fails when a non-numeric start timeout is given", func() {
			appRepo.ReadReturns.Error = errors.NewModelNotFoundError("App", "the-app")

			callPush("-t", "FooeyTimeout", "my-new-app")

			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"FAILED"},
				{"invalid", "timeout"},
			})
		})
	})

	Context("when a manifest has many apps", func() {
		BeforeEach(func() {
			appRepo.ReadReturns.Error = errors.NewModelNotFoundError("App", "the-app")
			manifestRepo.ReadManifestReturns.Manifest = manifestWithServicesAndEnv()
		})

		It("pushes each app", func() {
			callPush()

			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"Creating", "app1"},
				{"Creating", "app2"},
			})
			Expect(len(appRepo.CreateAppParams)).To(Equal(2))

			firstApp := appRepo.CreateAppParams[0]
			secondApp := appRepo.CreateAppParams[1]
			Expect(*firstApp.Name).To(Equal("app1"))
			Expect(*secondApp.Name).To(Equal("app2"))

			envVars := *firstApp.EnvironmentVars
			Expect(envVars["SOMETHING"]).To(Equal("definitely-something"))

			envVars = *secondApp.EnvironmentVars
			Expect(envVars["SOMETHING"]).To(Equal("nothing"))
		})

		It("pushes a single app when given the name of a single app in the manifest", func() {
			callPush("app2")

			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"Creating", "app2"},
			})
			testassert.SliceDoesNotContain(ui.Outputs, testassert.Lines{
				{"Creating", "app1"},
			})
			Expect(len(appRepo.CreateAppParams)).To(Equal(1))
			Expect(*appRepo.CreateAppParams[0].Name).To(Equal("app2"))
		})

		It("fails when given the name of an app that is not in the manifest", func() {
			callPush("non-existant-app")
			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"Failed"},
			})
			Expect(len(appRepo.CreateAppParams)).To(Equal(0))
		})
	})

	It("binds service instances to the app", func() {
		appRepo.ReadReturns.Error = errors.NewModelNotFoundError("App", "the-app")

		serviceRepo.FindInstanceByNameMap = generic.NewMap(map[interface{}]interface{}{
			"global-service": maker.NewServiceInstance("global-service"),
			"app1-service":   maker.NewServiceInstance("app1-service"),
			"app2-service":   maker.NewServiceInstance("app2-service"),
		})

		manifestRepo.ReadManifestReturns.Manifest = manifestWithServicesAndEnv()

		callPush()
		Expect(len(binder.AppsToBind)).To(Equal(4))
		Expect(binder.AppsToBind[0].Name).To(Equal("app1"))
		Expect(binder.AppsToBind[1].Name).To(Equal("app1"))
		Expect(binder.InstancesToBindTo[0].Name).To(Equal("app1-service"))
		Expect(binder.InstancesToBindTo[1].Name).To(Equal("global-service"))

		Expect(binder.AppsToBind[2].Name).To(Equal("app2"))
		Expect(binder.AppsToBind[3].Name).To(Equal("app2"))
		Expect(binder.InstancesToBindTo[2].Name).To(Equal("app2-service"))
		Expect(binder.InstancesToBindTo[3].Name).To(Equal("global-service"))

		testassert.SliceContains(ui.Outputs, testassert.Lines{
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

	It("fails when the service instances can't be found", func() {
		routeRepo.FindByHostAndDomainErr = true
		serviceRepo.FindInstanceByNameErr = true
		manifestRepo.ReadManifestReturns.Manifest = manifestWithServicesAndEnv()

		callPush()
		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"FAILED"},
			{"Could not find service", "app1-service", "app1"},
		})
	})

	It("stops the app, achieving a full-downtime deploy!", func() {
		existingApp := maker.NewApp(maker.Overrides{"name": "existing-app"})
		appRepo.ReadReturns.App = existingApp
		appRepo.UpdateAppResult = existingApp

		callPush("existing-app")

		Expect(stopper.AppToStop.Guid).To(Equal(existingApp.Guid))
		Expect(appBitsRepo.UploadedAppGuid).To(Equal(existingApp.Guid))
	})

	It("does not stop the app when it is already stopped", func() {
		stoppedApp := maker.NewApp(maker.Overrides{"state": "stopped", "name": "stopped-app"})

		appRepo.ReadReturns.App = stoppedApp
		appRepo.UpdateAppResult = stoppedApp

		callPush("stopped-app")

		Expect(stopper.AppToStop.Guid).To(Equal(""))
	})

	It("updates the app if it already exists", func() {
		existingRoute := models.RouteSummary{}
		existingRoute.Host = "existing-app"

		existingApp := models.Application{}
		existingApp.Name = "existing-app"
		existingApp.Guid = "existing-app-guid"
		existingApp.Routes = []models.RouteSummary{existingRoute}

		appRepo.ReadReturns.App = existingApp

		stack := models.Stack{}
		stack.Name = "differentStack"
		stack.Guid = "differentStack-guid"
		stackRepo.FindByNameStack = stack

		callPush(
			"-c", "different start command",
			"-i", "10",
			"-m", "1G",
			"-b", "https://github.com/heroku/heroku-buildpack-different.git",
			"-s", "differentStack",
			"existing-app",
		)

		Expect(*appRepo.UpdateParams.Command).To(Equal("different start command"))
		Expect(*appRepo.UpdateParams.InstanceCount).To(Equal(10))
		Expect(*appRepo.UpdateParams.Memory).To(Equal(uint64(1024)))
		Expect(*appRepo.UpdateParams.BuildpackUrl).To(Equal("https://github.com/heroku/heroku-buildpack-different.git"))
		Expect(*appRepo.UpdateParams.StackGuid).To(Equal("differentStack-guid"))
	})

	Describe("when the app already exists and has a route bound", func() {
		BeforeEach(func() {
			domain := models.DomainFields{
				Name:   "example.com",
				Guid:   "domain-guid",
				Shared: true,
			}

			domainRepo.ListDomainsForOrgDomains = []models.DomainFields{domain}

			existingRoute := models.RouteSummary{}
			existingRoute.Host = "existing-app"
			existingRoute.Domain = domain

			routeRepo.FindByHostAndDomainRoute = models.Route{
				RouteSummary: existingRoute,
			}

			existingApp := models.Application{}
			existingApp.Name = "existing-app"
			existingApp.Guid = "existing-app-guid"
			existingApp.Routes = []models.RouteSummary{existingRoute}

			appRepo.ReadReturns.App = existingApp
			appRepo.UpdateAppResult = existingApp
		})

		It("uses the existing route when an app already has it bound", func() {
			callPush("-d", "example.com", "existing-app")

			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"Using route", "existing-app", "example.com"},
			})
		})

		It("does not add a route to the app if no route-related flags are given", func() {
			callPush("existing-app")

			Expect(appBitsRepo.UploadedAppGuid).To(Equal("existing-app-guid"))
			Expect(domainRepo.FindByNameInOrgName).To(Equal(""))
			Expect(routeRepo.FindByHostAndDomainDomain).To(Equal(""))
			Expect(routeRepo.FindByHostAndDomainHost).To(Equal(""))
			Expect(routeRepo.CreatedHost).To(Equal(""))
			Expect(routeRepo.CreatedDomainGuid).To(Equal(""))
		})

		It("creates and binds a route when a different domain is specified", func() {
			newDomain := models.DomainFields{Guid: "domain-guid", Name: "newdomain.com"}
			routeRepo.FindByHostAndDomainNotFound = true
			domainRepo.FindByNameInOrgDomain = newDomain

			callPush("-d", "newdomain.com", "existing-app")

			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"Creating route", "existing-app.newdomain.com"},
				{"OK"},
				{"Binding", "existing-app.newdomain.com"},
			})

			Expect(domainRepo.FindByNameInOrgName).To(Equal("newdomain.com"))
			Expect(domainRepo.FindByNameInOrgGuid).To(Equal("my-org-guid"))
			Expect(routeRepo.FindByHostAndDomainDomain).To(Equal("newdomain.com"))
			Expect(routeRepo.FindByHostAndDomainHost).To(Equal("existing-app"))
			Expect(routeRepo.CreatedHost).To(Equal("existing-app"))
			Expect(routeRepo.CreatedDomainGuid).To(Equal("domain-guid"))
		})

		It("creates and binds a route when a different hostname is specified", func() {
			routeRepo.FindByHostAndDomainNotFound = true

			callPush("-n", "new-host", "existing-app")

			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"Creating route", "new-host.example.com"},
				{"OK"},
				{"Binding", "new-host.example.com"},
			})

			Expect(routeRepo.FindByHostAndDomainDomain).To(Equal("example.com"))
			Expect(routeRepo.FindByHostAndDomainHost).To(Equal("new-host"))
			Expect(routeRepo.CreatedHost).To(Equal("new-host"))
			Expect(routeRepo.CreatedDomainGuid).To(Equal("domain-guid"))
		})

		It("does not create a route when the --no-route flag is given", func() {
			callPush("--no-route", "existing-app")

			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"Uploading", "existing-app"},
				{"OK"},
			})

			Expect(appBitsRepo.UploadedAppGuid).To(Equal("existing-app-guid"))
			Expect(domainRepo.FindByNameInOrgName).To(Equal(""))
			Expect(routeRepo.FindByHostAndDomainDomain).To(Equal(""))
			Expect(routeRepo.FindByHostAndDomainHost).To(Equal(""))
			Expect(routeRepo.CreatedHost).To(Equal(""))
			Expect(routeRepo.CreatedDomainGuid).To(Equal(""))
		})

		It("binds the root domain route to an app with a pre-existing route when the --no-hostname flag is given", func() {
			routeRepo.FindByHostAndDomainNotFound = true

			callPush("--no-hostname", "existing-app")

			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"Creating route", "example.com"},
				{"OK"},
				{"Binding", "example.com"},
			})
			testassert.SliceDoesNotContain(ui.Outputs, testassert.Lines{
				{"existing-app.example.com"},
			})

			Expect(routeRepo.FindByHostAndDomainDomain).To(Equal("example.com"))
			Expect(routeRepo.FindByHostAndDomainHost).To(Equal(""))
			Expect(routeRepo.CreatedHost).To(Equal(""))
			Expect(routeRepo.CreatedDomainGuid).To(Equal("domain-guid"))
		})
	})

	Describe("displaying information about files being uploaded", func() {
		It("displays information about the files being uploaded", func() {
			appRepo.ReadReturns.Error = errors.NewModelNotFoundError("App", "the-app")
			appBitsRepo.CallbackPath = "path/to/app"
			appBitsRepo.CallbackZipSize = 61 * 1024 * 1024
			appBitsRepo.CallbackFileCount = 11

			callPush("appName")
			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"Uploading", "path/to/app"},
				{"61M", "11 files"},
			})
		})

		It("omits the size when there are no files being uploaded", func() {
			appRepo.ReadReturns.Error = errors.NewModelNotFoundError("App", "the-app")
			appBitsRepo.CallbackPath = "path/to/app"
			appBitsRepo.CallbackFileCount = 0

			callPush("appName")
			testassert.SliceContains(ui.WarnOutputs, testassert.Lines{
				{"None of your application files have changed", "nothing will be uploaded"},
			})
		})
	})

	It("fails when the app can't be uploaded", func() {
		appBitsRepo.UploadAppErr = true

		callPush("app")

		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Uploading"},
			{"FAILED"},
		})
	})

	Describe("when binding the route fails", func() {
		BeforeEach(func() {
			routeRepo.FindByHostAndDomainRoute.Host = "existing-app"
			routeRepo.FindByHostAndDomainRoute.Domain = models.DomainFields{Name: "foo.cf-app.com"}
		})

		It("suggests using 'random-route' if the default route is taken", func() {
			routeRepo.BindErr = errors.NewHttpError(400, errors.INVALID_RELATION, "The URL not available")

			callPush("existing-app")

			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"FAILED"},
				{"existing-app.foo.cf-app.com", "already in use"},
				{"TIP", "random-route"},
			})
		})

		It("does not suggest using 'random-route' for other failures", func() {
			routeRepo.BindErr = errors.NewHttpError(500, "some-code", "exception happened")

			callPush("existing-app")

			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"FAILED"},
			})

			testassert.SliceDoesNotContain(ui.Outputs, testassert.Lines{
				{"TIP", "random-route"},
			})
		})
	})

	It("fails when neither a manifest nor a name is given", func() {
		manifestRepo.ReadManifestReturns.Error = errors.New("No such manifest")
		callPush()
		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"FAILED"},
			{"App name"},
		})
	})
})

func singleAppManifest() *manifest.Manifest {
	return &manifest.Manifest{
		Path: "manifest.yml",
		Data: generic.NewMap(map[interface{}]interface{}{
			"applications": []interface{}{
				generic.NewMap(map[interface{}]interface{}{
					"name":      "manifest-app-name",
					"memory":    "128MB",
					"instances": 1,
					"host":      "manifest-host",
					"domain":    "manifest-example.com",
					"stack":     "custom-stack",
					"timeout":   360,
					"buildpack": "some-buildpack",
					"command":   `JAVA_HOME=$PWD/.openjdk JAVA_OPTS="-Xss995K" ./bin/start.sh run`,
					"path":      filepath.Clean("some/path/from/manifest"),
					"env": generic.NewMap(map[interface{}]interface{}{
						"FOO":  "baz",
						"PATH": "/u/apps/my-app/bin",
					}),
				}),
			},
		}),
	}
}

func manifestWithServicesAndEnv() *manifest.Manifest {
	return &manifest.Manifest{
		Data: generic.NewMap(map[interface{}]interface{}{
			"applications": []interface{}{
				generic.NewMap(map[interface{}]interface{}{
					"name":     "app1",
					"services": []interface{}{"app1-service", "global-service"},
					"env": generic.NewMap(map[interface{}]interface{}{
						"SOMETHING": "definitely-something",
					}),
				}),
				generic.NewMap(map[interface{}]interface{}{
					"name":     "app2",
					"services": []interface{}{"app2-service", "global-service"},
					"env": generic.NewMap(map[interface{}]interface{}{
						"SOMETHING": "nothing",
					}),
				}),
			},
		}),
	}
}
