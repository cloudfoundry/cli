package routes_test

import (
	"encoding/json"

	"code.cloudfoundry.org/diego-ssh/routes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Diego SSH Route", func() {
	var route routes.SSHRoute

	BeforeEach(func() {
		route = routes.SSHRoute{
			ContainerPort:   2222,
			HostFingerprint: "my-key-fingerprint",
			User:            "user",
			Password:        "password",
			PrivateKey:      "FAKE_PEM_ENCODED_KEY",
		}
	})

	Describe("JSON Marshalling", func() {
		Context("when the user and password are missing", func() {
			var expectedJson string

			BeforeEach(func() {
				route.User = ""
				route.Password = ""

				expectedJson = `{
					"container_port": 2222,
					"host_fingerprint": "my-key-fingerprint",
					"private_key": "FAKE_PEM_ENCODED_KEY"
				}`
			})

			It("marshals the structure correctly", func() {
				payload, err := json.Marshal(route)
				Expect(err).NotTo(HaveOccurred())

				Expect(payload).To(MatchJSON(expectedJson))
			})
		})

		Context("when the private key and host fingerprint are empty", func() {
			var expectedJson string

			BeforeEach(func() {
				route.PrivateKey = ""
				route.HostFingerprint = ""

				expectedJson = `{
					"container_port": 2222,
					"user": "user",
					"password": "password"
				}`
			})

			It("marshals the structure correctly", func() {
				payload, err := json.Marshal(route)
				Expect(err).NotTo(HaveOccurred())

				Expect(payload).To(MatchJSON(expectedJson))
			})
		})
	})

	Describe("Round Trip Marshalling", func() {
		It("successfully marshals and unmarshals", func() {
			payload, err := json.Marshal(route)
			Expect(err).NotTo(HaveOccurred())

			var result routes.SSHRoute
			err = json.Unmarshal(payload, &result)
			Expect(err).NotTo(HaveOccurred())

			Expect(result).To(Equal(route))
		})
	})
})
