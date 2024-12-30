package cfnetv1_test

import (
	"net/http"

	. "code.cloudfoundry.org/cfnetworking-cli-api/cfnetworking/cfnetv1"
	"code.cloudfoundry.org/cfnetworking-cli-api/cfnetworking/networkerror"
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

	Describe("ListPolicies", func() {
		var expectedPolicies []Policy
		Context("when the policies are found", func() {
			BeforeEach(func() {
				response := `{
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
								"protocol": "tcp",
								"ports": {
									"start": 4321,
									"end": 5321
								}
							}
						}
					]
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/policies"),
						RespondWith(http.StatusOK, response),
					),
				)
				expectedPolicies = []Policy{
					{
						Source: PolicySource{
							ID: "source-id-1",
						},
						Destination: PolicyDestination{
							ID:       "destination-id-1",
							Protocol: "tcp",
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
							Protocol: "tcp",
							Ports: Ports{
								Start: 4321,
								End:   5321,
							},
						},
					},
				}
			})

			It("returns the policies correctly", func() {
				policies, err := client.ListPolicies()
				Expect(policies).To(Equal(expectedPolicies))
				Expect(err).ToNot(HaveOccurred())
				Expect(server.ReceivedRequests()).To(HaveLen(1))
			})

			Context("when an app guid is passed", func() {
				It("makes the query correctly", func() {
					policies, err := client.ListPolicies("source-id-1")
					Expect(policies).To(Equal(expectedPolicies))
					Expect(err).ToNot(HaveOccurred())

					requests := server.ReceivedRequests()
					Expect(requests).To(HaveLen(1))
					Expect(requests[0].RequestURI).To(Equal("/policies?id=source-id-1"))
				})
			})

			Context("when multiple app guid are passed", func() {
				It("makes the query correctly", func() {
					policies, err := client.ListPolicies("source-id-1", "source-id-2")
					Expect(policies).To(Equal(expectedPolicies))
					Expect(err).ToNot(HaveOccurred())

					requests := server.ReceivedRequests()
					Expect(requests).To(HaveLen(1))
					Expect(requests[0].RequestURI).To(Equal("/policies?id=source-id-1%2Csource-id-2"))
				})
			})

		})

		Context("when the client returns an error", func() {
			BeforeEach(func() {
				response := `{
					"error": "Oh Noes"
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/policies"),
						RespondWith(http.StatusBadRequest, response),
					),
				)
			})

			It("returns the error", func() {
				_, err := client.ListPolicies()
				Expect(err).To(MatchError(networkerror.BadRequestError{
					Message: "Oh Noes",
				}))
			})
		})
	})

	Describe("RemovePolicies", func() {
		Context("when the policy is found", func() {
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
						VerifyRequest(http.MethodPost, "/policies/delete"),
						VerifyJSON(expectedBody),
						RespondWith(http.StatusOK, ""),
					),
				)
			})

			It("passes the body correctly", func() {
				err := client.RemovePolicies([]Policy{
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
						VerifyRequest(http.MethodPost, "/policies/delete"),
						RespondWith(http.StatusBadRequest, response),
					),
				)
			})

			It("returns the error", func() {
				err := client.RemovePolicies(nil)
				Expect(err).To(MatchError(networkerror.BadRequestError{
					Message: "Oh Noes",
				}))
			})
		})
	})
})
