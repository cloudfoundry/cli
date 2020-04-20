package isolated

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("logs command", func() {

	var server *ghttp.Server

	BeforeEach(func() {
		server = helpers.StartAndTargetMockServerWithAPIVersions(helpers.DefaultV2Version, helpers.DefaultV3Version)
		helpers.AddLoginRoutes(server)

		helpers.AddHandler(server,
			http.MethodGet,
			"/v3/organizations?order_by=name",
			http.StatusOK,
			[]byte(
				`{
				 "total_results": 1,
				 "total_pages": 1,
				 "resources": [
					 {
						 "guid": "f3ea75ba-ea6b-439f-8889-b07abf718e6a",
						 "name": "some-fake-org"
					 }
				 ]}`),
		)

		// The v6 version of this command makes the below request when logging in.
		// See below for comparison with v7 version.
		helpers.AddHandler(server,
			http.MethodGet,
			"/v3/spaces?organization_guids=f3ea75ba-ea6b-439f-8889-b07abf718e6a",
			http.StatusOK,
			[]byte(
				`{
					 "total_results": 1,
					 "total_pages": 1,
					 "resources": [
						 {
							 "guid": "1704b4e7-14bb-4b7b-bc23-0b8d23a60238",
							 "name": "some-fake-space"
						 }
					 ]}`),
		)

		// The v7 version of this command makes the below request when logging in,
		// which is similar to the v6 version above except for the additional 'order_by'
		// query parameter. Rather than split these tests across two files, we just add
		// a handler for both routes (with and without 'order_by').
		helpers.AddHandler(server,
			http.MethodGet,
			"/v3/spaces?order_by=name&organization_guids=f3ea75ba-ea6b-439f-8889-b07abf718e6a",
			http.StatusOK,
			[]byte(
				`{
					 "total_results": 1,
					 "total_pages": 1,
					 "resources": [
						 {
							 "guid": "1704b4e7-14bb-4b7b-bc23-0b8d23a60238",
							 "name": "some-fake-space"
						 }
					 ]}`),
		)

		helpers.AddHandler(server,
			http.MethodGet,
			"/v2/apps?q=name%3Asome-fake-app&q=space_guid%3A1704b4e7-14bb-4b7b-bc23-0b8d23a60238",
			http.StatusOK,
			[]byte(
				`{
					 "total_results": 1,
					 "total_pages": 1,
					 "resources": [
						 {
							 "metadata": {
									"guid": "d5d27772-315f-474b-8673-57e34ce2db2c"
							 },
							 "entity": {
									"name": "some-fake-app"
							 }
						 }
					 ]}`),
		)

		helpers.AddHandler(server,
			http.MethodGet,
			"/v3/apps?names=some-fake-app&space_guids=1704b4e7-14bb-4b7b-bc23-0b8d23a60238",
			http.StatusOK,
			[]byte(
				`{
					 "total_results": 1,
					 "total_pages": 1,
					 "resources": [
						 {
							 "guid": "d5d27772-315f-474b-8673-57e34ce2db2c",
							 "name": "some-fake-app"
						 }
					 ]}`),
		)

		helpers.AddHandler(server,
			http.MethodGet,
			"/api/v1/info",
			http.StatusOK,
			[]byte(`{"version":"2.6.8"}`),
		)
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("streaming logs", func() {

		const logMessage = "hello from log-cache"
		var returnEmptyEnvelope bool

		onWindows := runtime.GOOS == "windows"

		BeforeEach(func() {
			latestEnvelopeTimestamp := "1581447006352020890"
			nextEnvelopeTimestamp := "1581447009352020890"
			nextEnvelopeTimestampPlusOneNanosecond := "1581447009352020891"

			server.RouteToHandler(
				http.MethodGet,
				"/api/v1/read/d5d27772-315f-474b-8673-57e34ce2db2c",
				func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					switch r.URL.RawQuery {
					case fmt.Sprintf("descending=true&limit=1&start_time=%s", strconv.FormatInt(time.Time{}.UnixNano(), 10)):
						if returnEmptyEnvelope {
							_, err := w.Write([]byte(`{}`))
							Expect(err).ToNot(HaveOccurred())
							returnEmptyEnvelope = false // Allow the CLI to continue after receiving an empty envelope
						} else {
							_, err := w.Write([]byte(fmt.Sprintf(`
						{
							"envelopes": {
								"batch": [
									{
										"timestamp": "%s",
										"source_id": "d5d27772-315f-474b-8673-57e34ce2db2c"
									}
								]
							}
						}`, latestEnvelopeTimestamp)))
							Expect(err).ToNot(HaveOccurred())
						}
					case fmt.Sprintf("envelope_types=LOG&start_time=%s", latestEnvelopeTimestamp):
						_, err := w.Write([]byte(fmt.Sprintf(`{
							"envelopes": {
								"batch": [
									{
										"timestamp": "%s",
										"source_id": "d5d27772-315f-474b-8673-57e34ce2db2c",
										"tags": {
											"__v1_type": "LogMessage"
										},
										"log": {
											"payload": "%s",
											"type": "OUT"
										}
									}
								]
							}
							}`, nextEnvelopeTimestamp, base64.StdEncoding.EncodeToString([]byte(logMessage)))))
						Expect(err).ToNot(HaveOccurred())
					case fmt.Sprintf("envelope_types=LOG&start_time=%s", nextEnvelopeTimestampPlusOneNanosecond):
						_, err := w.Write([]byte("{}"))
						Expect(err).ToNot(HaveOccurred())
					default:
						Fail(fmt.Sprintf("Unhandled log-cache api query string: %s", r.URL.RawQuery))
					}
				})
		})

		When("there already is an envelope in the log cache", func() {
			JustBeforeEach(func() {
				returnEmptyEnvelope = false
			})

			It("fetches logs with a timestamp just prior to the latest log envelope", func() {
				username, password := helpers.GetCredentials()
				session := helpers.CF("login", "-a", server.URL(), "-u", username, "-p", password, "--skip-ssl-validation")
				Eventually(session).Should(Exit(0))

				session = helpers.CF("logs", "some-fake-app")
				Eventually(session).Should(Say(logMessage))
				if onWindows {
					session.Kill()
					Eventually(session).Should(Exit())
				} else {
					session.Interrupt()
					Eventually(session).Should(Exit(0), "Interrupt should be handled and fail gracefully")
				}
			})
		})

		When("there is not yet an envelope in the log cache", func() {
			JustBeforeEach(func() {
				returnEmptyEnvelope = true
			})

			// TODO: the case where log-cache has no envelopes yet may be "special": we may want to switch to "start from your oldest envelope" approach.
			It("retries until there is an initial envelope, and then fetches logs with a timestamp just prior to the latest log envelope", func() {
				username, password := helpers.GetCredentials()
				session := helpers.CF("login", "-a", server.URL(), "-u", username, "-p", password, "--skip-ssl-validation")
				Eventually(session).Should(Exit(0))

				session = helpers.CF("logs", "some-fake-app")
				Eventually(session).Should(Say(logMessage))
				if onWindows {
					session.Kill()
					Eventually(session).Should(Exit())
				} else {
					session.Interrupt()
					Eventually(session).Should(Exit(0))
				}
			})
		})
	})
})
