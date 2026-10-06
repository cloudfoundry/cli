package ccv3_test

import (
	"fmt"
	"net/http"

	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/v9/resources"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Service Accounts", func() {
	var client *Client
	BeforeEach(func() { client, _ = NewTestClient() })

	Describe("GetServiceAccounts", func() {
		It("filters by owning space, follows pagination and accumulates warnings", func() {
			server.AppendHandlers(
				CombineHandlers(VerifyRequest(http.MethodGet, "/v3/service_accounts", "space_guids=space-guid"),
					RespondWith(http.StatusOK, fmt.Sprintf(`{"pagination":{"next":{"href":"%s/v3/service_accounts?space_guids=space-guid&page=2"}},"resources":[{"guid":"one","name":"first-account","enabled":false,"status":"unprovisioned"}]}`, server.URL()), http.Header{"X-Cf-Warnings": {"first warning"}})),
				CombineHandlers(VerifyRequest(http.MethodGet, "/v3/service_accounts", "space_guids=space-guid&page=2"),
					RespondWith(http.StatusOK, `{"pagination":{"next":null},"resources":[{"guid":"two","name":"second-account","enabled":true,"status":"ready"}]}`, http.Header{"X-Cf-Warnings": {"second warning"}})),
			)
			accounts, warnings, err := client.GetServiceAccounts(Query{Key: SpaceGUIDFilter, Values: []string{"space-guid"}})
			Expect(err).NotTo(HaveOccurred())
			Expect(accounts).To(Equal([]resources.ServiceAccount{
				{GUID: "one", Name: "first-account", Enabled: false, Status: "unprovisioned"},
				{GUID: "two", Name: "second-account", Enabled: true, Status: "ready"},
			}))
			Expect(warnings).To(ConsistOf("first warning", "second warning"))
			Expect(server.ReceivedRequests()).To(HaveLen(2))
		})
		It("propagates list denial and warnings", func() {
			server.AppendHandlers(CombineHandlers(VerifyRequest(http.MethodGet, "/v3/service_accounts"),
				RespondWith(http.StatusForbidden, `{"errors":[{"code":10003,"title":"CF-NotAuthorized","detail":"denied"}]}`, http.Header{"X-Cf-Warnings": {"warning"}})))
			_, warnings, err := client.GetServiceAccounts()
			Expect(err).To(MatchError(ccerror.ForbiddenError{Message: "denied"}))
			Expect(warnings).To(ConsistOf("warning"))
		})
	})

	Describe("CreateServiceAccount", func() {
		It("posts only account input and space ownership and returns the platform identity and warnings", func() {
			server.AppendHandlers(CombineHandlers(
				VerifyRequest(http.MethodPost, "/v3/service_accounts"),
				VerifyJSON(`{"name":"shared-worker","description":"Background workers","relationships":{"space":{"data":{"guid":"space-guid"}}}}`),
				RespondWith(http.StatusCreated, `{"guid":"account-guid","name":"shared-worker","description":"Background workers","client_id":"cf:service-account:shared-worker","certificate_dns_san":"shared-worker.svc.identity","enabled":true,"status":"unprovisioned","relationships":{"space":{"data":{"guid":"space-guid"}}}}`, http.Header{"X-Cf-Warnings": {"account warning"}}),
			))
			account, warnings, err := client.CreateServiceAccount(resources.ServiceAccount{
				Name: "shared-worker", Description: "Background workers",
				Relationships: resources.Relationships{constant.RelationshipTypeSpace: {GUID: "space-guid"}},
				// These are response-only fields, never client-controlled creation input.
				GUID: "injected-guid", ClientID: "injected-client", CertificateDNSSAN: "injected-san", Status: "ready", Enabled: true,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(server.ReceivedRequests()).To(HaveLen(1))
			Expect(warnings).To(ConsistOf("account warning"))
			Expect(account).To(Equal(resources.ServiceAccount{
				GUID: "account-guid", Name: "shared-worker", Description: "Background workers",
				ClientID: "cf:service-account:shared-worker", CertificateDNSSAN: "shared-worker.svc.identity", Enabled: true, Status: "unprovisioned",
				Relationships: resources.Relationships{constant.RelationshipTypeSpace: {GUID: "space-guid"}},
			}))
		})

		It("omits an unspecified description", func() {
			server.AppendHandlers(CombineHandlers(
				VerifyRequest(http.MethodPost, "/v3/service_accounts"),
				VerifyJSON(`{"name":"shared-worker","relationships":{"space":{"data":{"guid":"space-guid"}}}}`),
				RespondWith(http.StatusCreated, `{"guid":"account-guid","name":"shared-worker"}`),
			))
			_, _, err := client.CreateServiceAccount(resources.ServiceAccount{Name: "shared-worker", Relationships: resources.Relationships{constant.RelationshipTypeSpace: {GUID: "space-guid"}}})
			Expect(err).NotTo(HaveOccurred())
			Expect(server.ReceivedRequests()).To(HaveLen(1))
		})

		It("returns CAPI denial and warnings without treating it as success", func() {
			server.AppendHandlers(CombineHandlers(
				VerifyRequest(http.MethodPost, "/v3/service_accounts"),
				RespondWith(http.StatusForbidden, `{"errors":[{"code":10003,"title":"CF-NotAuthorized","detail":"You are not authorized to perform the requested action"}]}`, http.Header{"X-Cf-Warnings": {"denial warning"}}),
			))
			_, warnings, err := client.CreateServiceAccount(resources.ServiceAccount{Name: "shared-worker"})
			Expect(err).To(MatchError(ccerror.ForbiddenError{Message: "You are not authorized to perform the requested action"}))
			Expect(warnings).To(ConsistOf("denial warning"))
		})
	})
})
