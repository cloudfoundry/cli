package uploads_test

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/uploads"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Upload", func() {
	Describe("CalculateRequestSize", func() {
		var (
			fileSize  int64
			path      string
			fieldName string
		)
		BeforeEach(func() {
			fileSize = 12
			path = "some/fake-droplet.tgz"
			fieldName = "bits"
		})

		It("", func() {
			contentLength, err := uploads.CalculateRequestSize(fileSize, path, fieldName)

			Expect(contentLength).To(Equal(260))
			Expect(err).To(BeNil())
		})

	})
})
