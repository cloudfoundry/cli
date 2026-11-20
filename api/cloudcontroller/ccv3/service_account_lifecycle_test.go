package ccv3_test

import (
	"net/http"

	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/v9/resources"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Service Account Lifecycle API", func() {
	var client *Client
	BeforeEach(func() { client, _ = NewTestClient() })

	It("fetches account detail by immutable GUID", func() {
		server.AppendHandlers(CombineHandlers(VerifyRequest(http.MethodGet, "/v3/service_accounts/account-guid"),
			RespondWith(http.StatusOK, `{"guid":"account-guid","name":"workers","description":"Workers","enabled":false,"status":"ready"}`, http.Header{"X-Cf-Warnings": {"detail warning"}})))
		account, warnings, err := client.GetServiceAccount("account-guid")
		Expect(err).NotTo(HaveOccurred())
		Expect(account).To(Equal(resources.ServiceAccount{GUID: "account-guid", Name: "workers", Description: "Workers", Enabled: false, Status: "ready"}))
		Expect(warnings).To(ConsistOf("detail warning"))
	})

	DescribeTable("encodes lifecycle changes and returns the asynchronous job location", func(operation, method, path, body string) {
		handlers := []http.HandlerFunc{VerifyRequest(method, path)}
		if body != "" {
			handlers = append(handlers, VerifyJSON(body))
		}
		handlers = append(handlers, RespondWith(http.StatusAccepted, "", http.Header{"Location": {server.URL() + "/v3/jobs/job-guid"}, "X-Cf-Warnings": {"operation warning"}}))
		server.AppendHandlers(CombineHandlers(handlers...))
		var job JobURL
		var warnings Warnings
		var err error
		switch operation {
		case "delete":
			job, warnings, err = client.DeleteServiceAccount("account-guid")
		case "enable":
			job, warnings, err = client.UpdateServiceAccountEnabled("account-guid", true)
		case "disable":
			job, warnings, err = client.UpdateServiceAccountEnabled("account-guid", false)
		case "bind":
			job, warnings, err = client.UpdateApplicationServiceAccount("app-guid", resources.Relationship{GUID: "account-guid"})
		case "unbind":
			job, warnings, err = client.UpdateApplicationServiceAccount("app-guid", resources.Relationship{})
		}
		Expect(err).NotTo(HaveOccurred())
		Expect(job).To(Equal(JobURL(server.URL() + "/v3/jobs/job-guid")))
		Expect(warnings).To(ConsistOf("operation warning"))
	},
		Entry("delete", "delete", http.MethodDelete, "/v3/service_accounts/account-guid", ""),
		Entry("enable", "enable", http.MethodPatch, "/v3/service_accounts/account-guid", `{"enabled":true}`),
		Entry("disable", "disable", http.MethodPatch, "/v3/service_accounts/account-guid", `{"enabled":false}`),
		Entry("bind", "bind", http.MethodPatch, "/v3/apps/app-guid/relationships/service_account", `{"data":{"guid":"account-guid"}}`),
		Entry("unbind", "unbind", http.MethodPatch, "/v3/apps/app-guid/relationships/service_account", `{"data":null}`),
	)

	It("supports synchronous unbinding and retains restart warnings", func() {
		server.AppendHandlers(CombineHandlers(VerifyRequest(http.MethodPatch, "/v3/apps/app-guid/relationships/service_account"), VerifyJSON(`{"data":null}`),
			RespondWith(http.StatusOK, `{"data":null}`, http.Header{"X-Cf-Warnings": {"Restart workers"}})))
		job, warnings, err := client.UpdateApplicationServiceAccount("app-guid", resources.Relationship{})
		Expect(err).NotTo(HaveOccurred())
		Expect(job).To(BeEmpty())
		Expect(warnings).To(ConsistOf("Restart workers"))
	})

	DescribeTable("preserves server errors and warnings", func(operation, method, path string) {
		server.AppendHandlers(CombineHandlers(VerifyRequest(method, path), RespondWith(http.StatusForbidden,
			`{"errors":[{"code":10003,"title":"CF-NotAuthorized","detail":"denied"}]}`, http.Header{"X-Cf-Warnings": {"warning"}})))
		var warnings Warnings
		var err error
		switch operation {
		case "show":
			_, warnings, err = client.GetServiceAccount("account-guid")
		case "delete":
			_, warnings, err = client.DeleteServiceAccount("account-guid")
		case "enable":
			_, warnings, err = client.UpdateServiceAccountEnabled("account-guid", true)
		case "bind":
			_, warnings, err = client.UpdateApplicationServiceAccount("app-guid", resources.Relationship{GUID: "account-guid"})
		}
		Expect(err).To(MatchError(ccerror.ForbiddenError{Message: "denied"}))
		Expect(warnings).To(ConsistOf("warning"))
	},
		Entry("show denial", "show", http.MethodGet, "/v3/service_accounts/account-guid"),
		Entry("delete denial", "delete", http.MethodDelete, "/v3/service_accounts/account-guid"),
		Entry("enable denial", "enable", http.MethodPatch, "/v3/service_accounts/account-guid"),
		Entry("assignment denial", "bind", http.MethodPatch, "/v3/apps/app-guid/relationships/service_account"),
	)
})
