package resources_test

import (
	"encoding/json"

	. "github.com/cloudfoundry/cli/cf/api/resources"
	testtime "github.com/cloudfoundry/cli/testhelpers/time"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Application resources", func() {
	var resource *ApplicationResource

	Describe("New Application", func() {
		BeforeEach(func() {
			resource = new(ApplicationResource)
		})

		It("Adds a packageUpdatedAt timestamp", func() {
			err := json.Unmarshal([]byte(`
			{
				"metadata": {
					"guid":"application-1-guid"
				},
				"entity": {
					"package_updated_at": "2013-10-07T16:51:07+00:00"
				}
			}`), &resource)

			Expect(err).NotTo(HaveOccurred())

			applicationModel := resource.ToModel()
			Expect(*applicationModel.PackageUpdatedAt).To(Equal(testtime.MustParse(eventTimestampFormat, "2013-10-07T16:51:07+00:00")))
		})
	})
})
