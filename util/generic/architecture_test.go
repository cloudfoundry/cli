package generic_test

import (
	. "code.cloudfoundry.org/cli/util/generic"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("architecture", func() {
	DescribeTable("GeneratePlatform",
		func(runtimeGOOS string, runtimeGOARCH string, platform string) {
			Expect(GeneratePlatform(runtimeGOOS, runtimeGOARCH)).To(Equal(platform))
		},

		Entry("linux64", "linux", "amd64", "linux64"),
		Entry("linux32", "linux", "386", "linux32"),
		Entry("win64", "windows", "amd64", "win64"),
		Entry("win32", "windows", "386", "win32"),
		Entry("osx", "darwin", "this is ignored", "osx"),

		Entry("linux64", "linux", "arm64", ""),
		Entry("linux32", "linux", "arm", ""),
		Entry("win64", "windows", "arm64", ""),
		Entry("win32", "windows", "arm", ""),
		Entry("osx", "darwin", "amd64", "osx"),
	)
})
