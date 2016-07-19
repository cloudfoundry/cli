package resources_test

import (
	"encoding/json"

	. "code.cloudfoundry.org/cli/cf/api/resources"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ServiceKeyResource", func() {
	var resource ServiceKeyResource

	BeforeEach(func() {
		err := json.Unmarshal([]byte(`
    {
      "metadata": {
        "guid": "fake-service-key-guid",
        "url": "/v2/service_keys/fake-guid",
        "created_at": "2015-01-13T18:52:08+00:00",
        "updated_at": null
      },
      "entity": {
        "name": "fake-service-key-name",
        "service_instance_guid":"fake-service-instance-guid",
        "service_instance_url":"http://fake/service/instance/url",
        "credentials": {
          "username": "fake-username",
          "password": "fake-password",
          "host": "fake-host",
          "port": 3306,
          "database": "fake-db-name",
          "uri": "mysql://fake-user:fake-password@fake-host:3306/fake-db-name"
        }
      }
    }`), &resource)

		Expect(err).ToNot(HaveOccurred())
	})

	Context("Brokers unmarshall service keys", func() {
		Describe("#ToFields", func() {
			It("unmarshalls the fields of a service key resource", func() {
				fields := resource.ToFields()

				Expect(fields.GUID).To(Equal("fake-service-key-guid"))
				Expect(fields.Name).To(Equal("fake-service-key-name"))
			})
		})

		Describe("#ToModel", func() {
			It("unmarshalls the service instance resource model", func() {
				instance := resource.ToModel()

				Expect(instance.Fields.Name).To(Equal("fake-service-key-name"))
				Expect(instance.Fields.GUID).To(Equal("fake-service-key-guid"))
				Expect(instance.Fields.URL).To(Equal("/v2/service_keys/fake-guid"))
				Expect(instance.Fields.ServiceInstanceGUID).To(Equal("fake-service-instance-guid"))
				Expect(instance.Fields.ServiceInstanceURL).To(Equal("http://fake/service/instance/url"))

				Expect(instance.Credentials).To(HaveKeyWithValue("username", "fake-username"))
				Expect(instance.Credentials).To(HaveKeyWithValue("password", "fake-password"))
				Expect(instance.Credentials).To(HaveKeyWithValue("host", "fake-host"))
				Expect(instance.Credentials).To(HaveKeyWithValue("port", float64(3306)))
				Expect(instance.Credentials).To(HaveKeyWithValue("database", "fake-db-name"))
				Expect(instance.Credentials).To(HaveKeyWithValue("uri", "mysql://fake-user:fake-password@fake-host:3306/fake-db-name"))
			})
		})
	})
})
