package cfnetv1_test

import (
	"net/http"

	. "code.cloudfoundry.org/cli/api/cfnetworking/cfnetv1"
	"code.cloudfoundry.org/cli/api/cfnetworking/networkerror"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Policy", func() {
	var client *Client

	BeforeEach(func() {
		client = NewTestClient()
	})

	Describe("CreatePolicies", func() {
		Context("when the stack is found", func() {
			BeforeEach(func() {
				expectedBody := `{
					"policies": [
						{
							"source": {
								"id": "source-id-1"
							},
							"destination": {
								"id": "destination-id-1",
								"protocol": "tcp",
								"ports": {
									"start": 1234,
									"end": 1235
								}
							}
						},
						{
							"source": {
								"id": "source-id-2"
							},
							"destination": {
								"id": "destination-id-2",
								"protocol": "udp",
								"ports": {
									"start": 1234,
									"end": 1235
								}
							}
						}
					]
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/policies"),
						VerifyJSON(expectedBody),
						RespondWith(http.StatusOK, ""),
					),
				)
			})

			It("passes the body correctly", func() {
				err := client.CreatePolicies([]Policy{
					{
						Source: PolicySource{
							ID: "source-id-1",
						},
						Destination: PolicyDestination{
							ID:       "destination-id-1",
							Protocol: PolicyProtocolTCP,
							Ports: Ports{
								Start: 1234,
								End:   1235,
							},
						},
					},
					{
						Source: PolicySource{
							ID: "source-id-2",
						},
						Destination: PolicyDestination{
							ID:       "destination-id-2",
							Protocol: PolicyProtocolUDP,
							Ports: Ports{
								Start: 1234,
								End:   1235,
							},
						},
					},
				})
				Expect(err).ToNot(HaveOccurred())
				Expect(server.ReceivedRequests()).To(HaveLen(1))
			})
		})

		Context("when the client returns an error", func() {
			BeforeEach(func() {
				response := `{
					"error": "Oh Noes"
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/policies"),
						RespondWith(http.StatusBadRequest, response),
					),
				)
			})

			It("returns the error and warnings", func() {
				err := client.CreatePolicies(nil)
				Expect(err).To(MatchError(networkerror.BadRequestError{
					Message: "Oh Noes",
				}))
			})
		})
	})
})
