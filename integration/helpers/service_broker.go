package helpers

import (
	"fmt"
	"net/http"
	"strings"

	"io/ioutil"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

const (
	DefaultMemoryLimit = "256M"
	DefaultDiskLimit   = "1G"
)

type Plan struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type ServiceBroker struct {
	Name       string
	Path       string
	AppsDomain string
	Service    struct {
		Name            string `json:"name"`
		ID              string `json:"id"`
		DashboardClient struct {
			ID          string `json:"id"`
			Secret      string `json:"secret"`
			RedirectUri string `json:"redirect_uri"`
		}
	}
	SyncPlans  []Plan
	AsyncPlans []Plan
}

func NewServiceBroker(name string, path string, appsDomain string, serviceName string, planName string) ServiceBroker {
	b := ServiceBroker{}
	b.Path = path
	b.Name = name
	b.AppsDomain = appsDomain
	b.Service.Name = serviceName
	b.Service.ID = RandomName()
	b.SyncPlans = []Plan{
		{Name: planName, ID: RandomName()},
		{Name: RandomName(), ID: RandomName()},
	}
	b.AsyncPlans = []Plan{
		{Name: RandomName(), ID: RandomName()},
		{Name: RandomName(), ID: RandomName()},
	}
	b.Service.DashboardClient.ID = RandomName()
	b.Service.DashboardClient.Secret = RandomName()
	b.Service.DashboardClient.RedirectUri = RandomName()
	return b
}

func (b ServiceBroker) Push() {
	Eventually(CF(
		"push", b.Name,
		"--no-start",
		"-m", DefaultMemoryLimit,
		"-p", b.Path,
		"-d", b.AppsDomain,
	)).Should(Exit(0))

	Eventually(CF("start", b.Name)).Should(Exit(0))
}

func (b ServiceBroker) Configure() {
	uri := fmt.Sprintf("http://%s.%s%s", b.Name, b.AppsDomain, "/config")
	body := strings.NewReader(b.ToJSON())
	req, err := http.NewRequest("POST", uri, body)
	Expect(err).ToNot(HaveOccurred())
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	Expect(err).ToNot(HaveOccurred())
	defer resp.Body.Close()
}

func (b ServiceBroker) Create() {
	appURI := fmt.Sprintf("http://%s.%s", b.Name, b.AppsDomain)
	Eventually(CF("create-service-broker", b.Name, "username", "password", appURI)).Should(Exit(0))
	Eventually(CF("service-brokers")).Should(And(Exit(0), Say(b.Name)))
}

func (b ServiceBroker) Delete() {
	Eventually(CF("delete-service-broker", b.Name, "-f")).Should(Exit(0))
	Eventually(CF("service-brokers")).Should(And(Exit(0), Not(Say(b.Name))))
}

func (b ServiceBroker) Destroy() {
	Eventually(CF("purge-service-offering", b.Service.Name, "-f")).Should(Exit(0))
	b.Delete()
	Eventually(CF("delete", b.Name, "-f", "-r")).Should(Exit(0))
}

func (b ServiceBroker) ToJSON() string {
	bytes, err := ioutil.ReadFile(NewAssets().ServiceBroker + "/cats.json")
	Expect(err).To(BeNil())

	replacer := strings.NewReplacer(
		"<fake-service>", b.Service.Name,
		"<fake-service-guid>", b.Service.ID,
		"<sso-test>", b.Service.DashboardClient.ID,
		"<sso-secret>", b.Service.DashboardClient.Secret,
		"<sso-redirect-uri>", b.Service.DashboardClient.RedirectUri,
		"<fake-plan>", b.SyncPlans[0].Name,
		"<fake-plan-guid>", b.SyncPlans[0].ID,
		"<fake-plan-2>", b.SyncPlans[1].Name,
		"<fake-plan-2-guid>", b.SyncPlans[1].ID,
		"<fake-async-plan>", b.AsyncPlans[0].Name,
		"<fake-async-plan-guid>", b.AsyncPlans[0].ID,
		"<fake-async-plan-2>", b.AsyncPlans[1].Name,
		"<fake-async-plan-2-guid>", b.AsyncPlans[1].ID,
	)

	return replacer.Replace(string(bytes))
}

func GetAppGuid(appName string) string {
	session := CF("app", appName, "--guid")
	Eventually(session).Should(Exit(0))

	appGuid := strings.TrimSpace(string(session.Out.Contents()))
	Expect(appGuid).NotTo(Equal(""))
	return appGuid
}

type Assets struct {
	ServiceBroker string
}

func NewAssets() Assets {
	return Assets{
		ServiceBroker: "../assets/service_broker",
	}
}
