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

/*
Possibly helpful code taken from ccv3/droplet
	FWhen("there is an error reading the buildpack", func() {
			var (
				fakeReader  *ccv3fakes.FakeReader
				expectedErr error
			)

			BeforeEach(func() {
				expectedErr = errors.New("droplet read error")
				fakeReader = new(ccv3fakes.FakeReader)
				fakeReader.ReadReturns(0, expectedErr)
				dropletFile = fakeReader
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError(expectedErr))
			})
		})
 */
