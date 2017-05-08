package util_test

import (
	. "code.cloudfoundry.org/cli/util"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("util", func() {

	DescribeTable("URL's",
		func(path string, isHTTPScheme bool) {
			Expect(IsHTTPScheme(path)).To(Equal(isHTTPScheme))
		},

		Entry("proper HTTP URL", "http://example.com", true),
		Entry("proper HTTPS URL", "https://example.com", true),
		Entry("not proper HTTP URL", "http//example.com", false),
		Entry("proper FTP URL", "ftp://example.com", false),
		Entry("local file name", "some-path", false),
		Entry("UNIX path", "/some/path", false),
		Entry("Windows path", "C:\\some\\path", false),
	)

	DescribeTable("IsUnsupportedScheme",
		func(path string, isUnsupportedScheme bool) {
			Expect(IsUnsupportedURLScheme(path)).To(Equal(isUnsupportedScheme))
		},

		Entry("proper HTTP URL", "http://example.com", false),
		Entry("proper HTTPS URL", "https://example.com", false),
		Entry("not proper HTTP URL", "http//example.com", false),
		Entry("proper FTP URL", "ftp://example.com", true),
		Entry("local file name", "some-path", false),
		Entry("UNIX path", "/some/path", false),
		Entry("Windows path", "C:\\some\\path", false),
	)
})
