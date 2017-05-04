package util_test

import (
	. "code.cloudfoundry.org/cli/util"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DeterminePathType", func() {
	Context("when the string passed is not a proper HTTP URL", func() {
		It("returns that it is a PluginFilePath", func() {
			Expect(DeterminePathType("http//example.com")).To(Equal(PluginFilePath))
		})
	})

	Context("when the string passed is a proper HTTP URL", func() {
		It("returns that it is an HTTP path", func() {
			Expect(DeterminePathType("http://example.com")).To(Equal(PluginHTTPPath))
		})
	})

	Context("when the string passed is a proper HTTPS URL", func() {
		It("returns that it is an HTTP path", func() {
			Expect(DeterminePathType("https://example.com")).To(Equal(PluginHTTPPath))
		})
	})

	Context("when the string passed is a proper FTP URL", func() {
		It("returns that it is of an unsupported path type", func() {
			Expect(DeterminePathType("ftp://example.com")).To(Equal(PluginUnsupportedPathType))
		})
	})

	Context("when the string passed is a local file name", func() {
		It("returns that it is a PluginFilePath", func() {
			Expect(DeterminePathType("some-path")).To(Equal(PluginFilePath))
		})
	})

	Context("when the string passed is a UNIX path", func() {
		It("returns that it is a PluginFilePath", func() {
			Expect(DeterminePathType("/some/path")).To(Equal(PluginFilePath))
		})
	})

	Context("when the string passed is a Windows path", func() {
		It("returns that it is a PluginFilePath", func() {
			Expect(DeterminePathType("C:\\some\\path")).To(Equal(PluginFilePath))
		})
	})
})
