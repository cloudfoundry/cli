// fakeservicebroker is a test helper for service broker tests. It can push an App that acts as a service broker, and
// as an optimisation, it will reuse the same App so that multiple pushes aren't necessary. This saves significant time
// when running tests.
//
// To make use of this optimisation, set 'KEEP_FAKE_SERVICE_BROKERS' to 'true'
package fakeservicebroker

import (
	"fmt"
	"os"
	"strings"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

// Setup should be run in a SynchronisedBeforeSuite() to create the org and space,
// otherwise the tests become flaky.
func Setup() {
	helpers.WithRandomHomeDir(func() {
		helpers.SetAPI()
		helpers.LoginCF()
		helpers.CreateOrgAndSpaceUnlessExists(itsOrg, itsSpace)
	})
}

// Cleanup should be run in a SynchronizedAfterSuite() to clean up the reusable broker apps
func Cleanup() {
	helpers.WithRandomHomeDir(func() {
		helpers.SetAPI()
		helpers.LoginCF()
		helpers.CreateOrgAndSpaceUnlessExists(itsOrg, itsSpace)
		helpers.TargetOrgAndSpace(itsOrg, itsSpace)

		broker := New()
		otherBroker := NewAlternate()

		if os.Getenv("KEEP_FAKE_SERVICE_BROKERS") != "true" {
			broker.stopReusing()
			otherBroker.stopReusing()
		}

		broker.cleanup()
		otherBroker.cleanup()
	})
}

type FakeServiceBroker struct {
	name      string
	username  string
	password  string
	Services  []service
	domain    string
	behaviors behaviors
	reusable  bool
}

// New creates a default object that can then be configured
func New() *FakeServiceBroker {
	return defaultConfig()
}

// NewAlternate returns a reusable broker with another name. Can be used in conjunction with New() if you need two brokers in the same test.
func NewAlternate() *FakeServiceBroker {
	f := defaultConfig()
	f.name = generateReusableBrokerName("other-")
	return f
}

// WithName has the side-effect that the broker is not reusable
func (f *FakeServiceBroker) WithName(name string) *FakeServiceBroker {
	f.name = name
	f.reusable = false
	return f
}

func (f *FakeServiceBroker) WithAsyncBehaviour() *FakeServiceBroker {
	f.behaviors.Provision["default"] = asyncResponse()
	f.behaviors.Update["default"] = asyncResponse()
	f.behaviors.Deprovision["default"] = asyncResponse()
	f.behaviors.Bind["default"] = asyncResponse().asyncOnly()
	f.behaviors.Unbind["default"] = asyncResponse().asyncOnly()
	return f
}

// EnsureAppIsDeployed makes the fake service broker app available and does not run 'cf create-service-broker'
func (f *FakeServiceBroker) EnsureAppIsDeployed() *FakeServiceBroker {
	f.pushAppIfNecessary()
	f.deregister()
	f.configure()
	return f
}

// EnsureBrokerIsAvailable makes the service broker app available and runs 'cf create-service-broker'
func (f *FakeServiceBroker) EnsureBrokerIsAvailable() *FakeServiceBroker {
	f.EnsureAppIsDeployed()
	f.register()
	return f
}

func (f *FakeServiceBroker) EnableServiceAccess() {
	session := helpers.CF("enable-service-access", f.ServiceName())
	Eventually(session).Should(Exit(0))
}

func (f *FakeServiceBroker) Update() *FakeServiceBroker {
	f.configure()
	f.update()
	return f
}

// Destroy always deletes the broker and app, meaning it can't be reused. In general you should not destroy brokers
// as it makes tests slower.
func (f *FakeServiceBroker) Destroy() {
	f.deregister()
	f.deleteApp()
}

func (f *FakeServiceBroker) URL(paths ...string) string {
	return fmt.Sprintf("http://%s.%s%s", f.name, f.domain, strings.Join(paths, ""))
}

func (f *FakeServiceBroker) ServiceName() string {
	return f.Services[0].Name
}

func (f *FakeServiceBroker) ServiceDescription() string {
	return f.Services[0].Description
}

func (f *FakeServiceBroker) ServicePlanDescription() string {
	return f.Services[0].Plans[0].Description
}

func (f *FakeServiceBroker) ServicePlanName() string {
	return f.Services[0].Plans[0].Name
}

func (f *FakeServiceBroker) Name() string {
	return f.name
}

func (f *FakeServiceBroker) Username() string {
	return f.username
}

func (f *FakeServiceBroker) Password() string {
	return f.password
}
